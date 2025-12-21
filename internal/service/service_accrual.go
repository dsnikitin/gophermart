package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/dsnikitin/gophermart/internal/models"
	accrual_status "github.com/dsnikitin/gophermart/internal/pkg/consts/accrual"
	order_status "github.com/dsnikitin/gophermart/internal/pkg/consts/order"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"
	"github.com/pkg/errors"
)

const restartingStaleOrdersInterval time.Duration = time.Minute * 5
const handlersCount int = 100
const attempts int = 5

type AccrualRepository interface {
	GetOldestOrderWithLock(ctx context.Context, status order_status.Status) (models.Order, error)
	UpdateOrder(ctx context.Context, number string, status order_status.Status, accrual *float64) error
	GetStaleOrders(ctx context.Context, threshold time.Time, statuses ...order_status.Status) ([]models.Order, error)
}

type AccrualTxProvider interface {
	Do(ctx context.Context, fn func(AccrualRepository) error) error
}

type AccrualService struct {
	accrualSystemAddr string
	client            *http.Client

	r  AccrualRepository
	tx AccrualTxProvider

	wakeUpCh chan struct{}
	ordersCh chan models.Order

	wg     sync.WaitGroup
	stopCh chan struct{}
}

func NewAccrualService(accrualSystemAddr string, r AccrualRepository, tx AccrualTxProvider) *AccrualService {
	s := &AccrualService{
		accrualSystemAddr: accrualSystemAddr,
		client:            initClient(),
		r:                 r,
		tx:                tx,
		wakeUpCh:          make(chan struct{}, 1),
		ordersCh:          make(chan models.Order),
		stopCh:            make(chan struct{}),
	}

	for range handlersCount {
		s.wg.Add(1)
		go s.ordersHandler()
	}

	s.wg.Add(1)
	go s.ordersProducer()

	s.wg.Add(1)
	go s.staleOrdersProducer()

	return s
}

func (s *AccrualService) NotifyOrderUploaded() {
	fmt.Println("IN NotifyOrderUploaded")
	select {
	case s.wakeUpCh <- struct{}{}:
	default:
	}
}

func (s *AccrualService) Stop() {
	close(s.stopCh)
	s.wg.Wait()

	logger.Log.Info("Accrual service stopped")
}

func (s *AccrualService) ordersProducer() {
	defer s.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-s.stopCh:
			return
		case <-s.wakeUpCh:
			fmt.Println("IN ordersProducer")
			for {
				if err := s.produceOrder(ctx); err != nil {
					fmt.Println("AFTER produceOrder, err =", err)
					if errors.Is(err, errx.ErrNotFound) {
						break
					}

					if errors.Is(err, errx.ErrAllWorkersBusy) {
						logger.Log.Infow("All workers are busy for proccessing next order")
						break
					}

					logger.Log.Warnw("Failed to produce next order to processing", "error", err.Error())
				}
			}
		}
	}
}

func (s *AccrualService) staleOrdersProducer() {
	defer s.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.produceStaleOrders(ctx); err != nil {
		err = errors.Wrap(err, "produce stale orders")
		logger.Log.Warnw("Failed to produce stale orders on start", "error", err.Error())
	}

	ticker := time.NewTicker(restartingStaleOrdersInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.produceStaleOrders(ctx); err != nil {
				err = errors.Wrap(err, "produce stale orders")
				logger.Log.Warnw("Failed to produce stale orders", "error", err.Error())
			}
		}
	}
}

func (s *AccrualService) ordersHandler() {
	defer s.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-s.stopCh:
			return
		case order, ok := <-s.ordersCh:
			fmt.Println("IN ordersHandler")
			if !ok {
				return
			}

			if err := s.handleOrder(ctx, order); err != nil {
				err = errors.Wrap(err, "handle order")
				logger.Log.Errorw("Failed to handle order", "orderNumber", order.Number, "error", err)
			}

			select {
			case s.wakeUpCh <- struct{}{}:
			default:
			}
		}
	}
}

func (s *AccrualService) produceOrder(ctx context.Context) error {
	fmt.Println("IN produceOrder")

	err := s.tx.Do(ctx, func(rTx AccrualRepository) error {
		order, err := rTx.GetOldestOrderWithLock(ctx, order_status.New)
		if err != nil {
			return errors.Wrap(err, "get olders order with lock")
		}

		fmt.Printf("ORDER = %+v\n", order)

		if err := rTx.UpdateOrder(ctx, order.Number, order_status.Processing, nil); err != nil {
			return errors.Wrap(err, "update order")
		}

		for range attempts {
			select {
			case s.ordersCh <- order:
				return nil
			default:
				time.Sleep(time.Millisecond * 500)
			}
		}

		return errx.ErrAllWorkersBusy
	})

	return errors.Wrap(err, "do tx")
}

func (s *AccrualService) produceStaleOrders(ctx context.Context) error {
	threshold := time.Now().UTC().Add(-restartingStaleOrdersInterval)

	orders, err := s.r.GetStaleOrders(ctx, threshold, order_status.New, order_status.Processing)
	if err != nil {
		return errors.Wrap(err, "get stale orders")
	}

	fmt.Println("staleOrdersLen =", len(orders))

	for _, order := range orders {
		if order.Status == order_status.New {
			s.NotifyOrderUploaded()
			continue
		}

		for range attempts {
			select {
			case s.ordersCh <- order:
			default:
				time.Sleep(time.Millisecond * 500)
			}
		}
	}

	return nil
}

func (s *AccrualService) handleOrder(ctx context.Context, order models.Order) error {
	fmt.Println("IN handleOrder")
	for range attempts {
		select {
		case <-s.stopCh:
			return nil
		default:
			accrual, err := s.getAccrual(ctx, order.Number)
			if err != nil {
				err := errors.Wrap(err, "get accrual")

				switch {
				case errors.Is(err, errx.ErrToManyRequests):
					time.Sleep(time.Second * time.Duration(rand.Intn(10)))
					continue
				case errors.Is(err, errx.ErrUnregisteredOrder):
					logger.Log.Infow("Order is not registered in accrual system", "orderNumber", order.Number)
					err := s.r.UpdateOrder(ctx, order.Number, order_status.Invalid, nil)
					return errors.Wrap(err, "update non-registred order status")
				default:
					logger.Log.Warnw("Failed to get accrual", "orderNumber", order.Number, "error", err.Error())
					continue
				}
			}

			switch accrual.Status {
			case accrual_status.Invalid:
				err := s.r.UpdateOrder(ctx, order.Number, order_status.Invalid, nil)
				return errors.Wrap(err, "update invalid order status")
			case accrual_status.Processed:
				err := s.r.UpdateOrder(ctx, order.Number, order_status.Processed, &accrual.Accrual)
				return errors.Wrap(err, "update processed order status and acrrual")
			}
		}
	}

	return nil
}

func (s *AccrualService) getAccrual(ctx context.Context, number string) (models.Accrual, error) {
	fmt.Println("IN getAccural")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.accrualSystemAddr+"/api/orders/"+number, nil)
	if err != nil {
		return models.Accrual{}, errors.Wrap(err, "new http request")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return models.Accrual{}, errors.Wrap(err, "do http request")
	}
	defer resp.Body.Close()

	fmt.Println("RESP STATUS CODE =", resp.StatusCode)

	switch resp.StatusCode {
	case http.StatusOK:
		var accrual models.Accrual
		if err := json.NewDecoder(resp.Body).Decode(&accrual); err != nil {
			return models.Accrual{}, errors.Wrap(err, "decode accrual")
		}
		return accrual, nil
	case http.StatusNoContent:
		return models.Accrual{}, errx.ErrUnregisteredOrder
	case http.StatusTooManyRequests:
		return models.Accrual{}, errx.ErrToManyRequests
	default:
		return models.Accrual{}, errors.Wrap(errx.ErrInternalServer, "accrualer server error")
	}
}

func initClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		},
	}
}
