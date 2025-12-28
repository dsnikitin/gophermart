package service

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/dsnikitin/gophermart/internal/models"
	accrualstatus "github.com/dsnikitin/gophermart/internal/pkg/consts/accrual"
	orderstatus "github.com/dsnikitin/gophermart/internal/pkg/consts/order"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const restartingStaleOrdersInterval time.Duration = time.Minute * 5
const handlersCount int = 100
const attempts int = 5

type AccrualRepository interface {
	GetOldestOrderWithLock(ctx context.Context, status orderstatus.Status) (models.Order, error)
	UpdateOrder(ctx context.Context, number string, status orderstatus.Status, accrual *float64) error
	GetStaleOrders(ctx context.Context, threshold time.Time, status orderstatus.Status) ([]models.Order, error)
}

type AccrualTxProvider interface {
	Do(ctx context.Context, fn func(AccrualRepository) error) error
}

type AccrualService struct {
	accrualSystemAddr string
	client            *http.Client

	r  AccrualRepository
	tx AccrualTxProvider

	startCh  chan struct{}
	ordersCh chan models.Order

	ctx    context.Context
	cancel context.CancelFunc
	eg     *errgroup.Group
}

func NewAccrualService(accrualSystemAddr string, r AccrualRepository, tx AccrualTxProvider) *AccrualService {
	ctx, cancel := context.WithCancel(context.Background())

	s := &AccrualService{
		accrualSystemAddr: accrualSystemAddr,
		client:            initClient(),
		r:                 r,
		tx:                tx,
		startCh:           make(chan struct{}, 1),
		ordersCh:          make(chan models.Order),
		eg:                &errgroup.Group{},
		ctx:               ctx,
		cancel:            cancel,
	}

	for range handlersCount {
		s.eg.Go(s.ordersHandler)
	}

	s.eg.Go(s.ordersProducer)
	s.eg.Go(s.staleOrdersProducer)

	s.NotifyOrderUploaded()
	return s
}

func (s *AccrualService) NotifyOrderUploaded() {
	select {
	case s.startCh <- struct{}{}:
	default:
	}
}

func (s *AccrualService) Stop() {
	s.cancel()
	s.eg.Wait()
	logger.Log.Info("Accrual service stopped")
}

func (s *AccrualService) ordersProducer() error {
	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-s.startCh:
			for {
				if err := s.produceOrder(); err != nil {
					if errors.Is(err, errx.ErrNotFound) {
						break
					}

					err := errors.Wrap(err, "produce order")
					logger.Log.Warnw("Failed to produce next order", "error", err.Error())
				}
			}
		}
	}
}

func (s *AccrualService) staleOrdersProducer() error {
	logger.Log.Info("Restarting stale orders...")
	if err := s.restartStaleOrders(time.Now().UTC()); err != nil {
		err = errors.Wrap(err, "restart stale orders")
		logger.Log.Errorw("Failed to restart stale orders on start", "error", err.Error())
	}

	ticker := time.NewTicker(restartingStaleOrdersInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-ticker.C:
			err := s.restartStaleOrders(time.Now().UTC().Add(-restartingStaleOrdersInterval))
			if err != nil {
				err = errors.Wrap(err, "restart stale orders")
				logger.Log.Errorw("Failed to restart stale orders", "error", err.Error())
			}
		}
	}
}

func (s *AccrualService) ordersHandler() error {
	for {
		select {
		case <-s.ctx.Done():
			return nil
		case order, ok := <-s.ordersCh:
			if !ok {
				return nil
			}

			if err := s.handleOrder(order); err != nil {
				if errors.Is(err, errx.ErrUnregisteredOrder) {
					logger.Log.Warnw("Order is not registered in accrual system", "orderNumber", order.Number)
					break
				}

				err = errors.Wrap(err, "handle order")
				logger.Log.Errorw("Failed to handle order", "orderNumber", order.Number, "error", err)
			}

			select {
			case s.startCh <- struct{}{}:
			default:
			}
		}
	}
}

func (s *AccrualService) produceOrder() error {
	err := s.tx.Do(s.ctx, func(rTx AccrualRepository) error {
		order, err := rTx.GetOldestOrderWithLock(s.ctx, orderstatus.New)
		if err != nil {
			return errors.Wrap(err, "get olders order with lock")
		}

		if err := rTx.UpdateOrder(s.ctx, order.Number, orderstatus.Processing, nil); err != nil {
			return errors.Wrap(err, "update order")
		}

		if ok := s.sendToChan(order); !ok {
			return errx.ErrAllWorkersBusy
		}

		return nil
	})

	return errors.Wrap(err, "do tx")
}

func (s *AccrualService) restartStaleOrders(stalingThreshold time.Time) error {
	orders, err := s.r.GetStaleOrders(s.ctx, stalingThreshold, orderstatus.Processing)
	if err != nil {
		return errors.Wrap(err, "get stale orders")
	}

	for _, order := range orders {
		if ok := s.sendToChan(order); !ok {
			return errx.ErrAllWorkersBusy
		}
	}

	return nil
}

func (s *AccrualService) handleOrder(order models.Order) error {
	for range attempts {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			accrual, err := s.getAccrual(order.Number)
			if err != nil {
				if errors.Is(err, errx.ErrToManyRequests) {
					time.Sleep(time.Second * time.Duration(rand.Intn(10)))
					continue
				}

				return errors.Wrap(err, "get accrual")
			}

			switch accrual.Status {
			case accrualstatus.Invalid:
				err := s.r.UpdateOrder(s.ctx, order.Number, orderstatus.Invalid, nil)
				return errors.Wrap(err, "update order status to invalid")
			case accrualstatus.Processed:
				err := s.r.UpdateOrder(s.ctx, order.Number, orderstatus.Processed, &accrual.Accrual)
				return errors.Wrap(err, "update acrrual and order status to processed")
			}
		}
	}

	return nil
}

func (s *AccrualService) getAccrual(number string) (models.Accrual, error) {
	endpoint := s.accrualSystemAddr + "/api/orders/" + number
	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return models.Accrual{}, errors.Wrap(err, "new http request")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return models.Accrual{}, errors.Wrap(err, "do http request")
	}
	defer resp.Body.Close()

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
		return models.Accrual{}, errors.Wrap(errx.ErrInternalServer, "accrual system internal error")
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

func (s *AccrualService) sendToChan(order models.Order) bool {
	for range attempts {
		select {
		case s.ordersCh <- order:
			return true
		default:
			time.Sleep(time.Millisecond * 500)
		}
	}

	return false
}
