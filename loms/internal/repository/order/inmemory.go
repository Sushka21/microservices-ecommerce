package order

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Sushka21/microservices-ecommerce/loms/internal/entity"
)

var ErrOrderNotFound = errors.New("order not found")

type inMemoryRepository struct {
	mx       sync.RWMutex
	orders   map[int64]entity.Order
	IDorders int64
}

func NewInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		mx:       sync.RWMutex{},
		orders:   map[int64]entity.Order{},
		IDorders: 0,
	}
}

func (r *inMemoryRepository) CreatOrder(_ context.Context, order entity.Order) (int64, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.IDorders++
	order.ID = r.IDorders
	r.orders[r.IDorders] = order
	return order.ID, nil
}

func (r *inMemoryRepository) GetOrderByID(_ context.Context, id int64) (entity.Order, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	v, ok := r.orders[id]
	if !ok {
		return entity.Order{}, entity.ErrOrderNotFound
	}
	return v, nil
}

func (r *inMemoryRepository) SetStatusByID(_ context.Context, id int64, status entity.OrderStatus) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	v, ok := r.orders[id]
	if !ok {
		return entity.ErrOrderNotFound
	}
	v.Status = status
	v.UpdatedAt = time.Now()
	r.orders[id] = v
	return nil
}



