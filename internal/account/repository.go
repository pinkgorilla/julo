package account

import (
	"context"
	"sync"
)

type Repository interface {
	CreateAccount(c context.Context, a Account) error
	GetAccount(c context.Context, xid string) (*Account, error)
}

type InMemoryRepository struct {
	store sync.Map
}

func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		store: sync.Map{},
	}
}

func (r *InMemoryRepository) CreateAccount(c context.Context, a Account) error {
	r.store.Store(a.XID, &a)
	return nil
}

func (r *InMemoryRepository) GetAccount(c context.Context, xid string) (*Account, error) {
	v, ok := r.store.Load(xid)
	if !ok {
		return nil, ErrAccountNotFound
	}
	return v.(*Account), nil
}
