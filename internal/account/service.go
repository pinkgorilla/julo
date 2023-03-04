package account

import (
	"context"

	"github.com/pkg/errors"
)

type Service interface {
	CreateAccount(context.Context, Account) error
	GetAccount(c context.Context, xid string) (*Account, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}
func (s *service) CreateAccount(c context.Context, a Account) error {
	acc, err := s.repo.GetAccount(c, a.XID)
	if err != nil && err != ErrAccountNotFound {
		return errors.Wrap(err, "failed getting account")
	}

	if acc != nil {
		return ErrAccountAlreadyExists
	}

	err = s.repo.CreateAccount(c, a)
	if err != nil {
		return errors.Wrap(err, "failed creating user")
	}

	return nil
}

func (s *service) GetAccount(c context.Context, xid string) (*Account, error) {
	return s.repo.GetAccount(c, xid)
}
