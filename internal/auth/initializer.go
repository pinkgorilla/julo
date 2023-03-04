package auth

import (
	"context"
	"julo/internal/account"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Initializer interface {
	Init(context.Context, InitParam) (*InitResult, error)
}

type InitParam struct {
	CustomerXID string
}
type InitResult struct {
	Session Session
}

type initializer struct {
	account account.Service
}

func NewInitializer(account account.Service) Initializer {
	return &initializer{
		account: account,
	}
}

func (i *initializer) Init(c context.Context, p InitParam) (*InitResult, error) {
	account := account.Account{
		XID: p.CustomerXID,
	}
	err := i.account.CreateAccount(c, account)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating account")
	}

	token := uuid.NewString()
	session := Session{
		Token:   token,
		Account: account,
	}

	err = StoreSession(c, session)
	if err != nil {
		return nil, errors.Wrap(err, "failed storing session")
	}

	return &InitResult{
		Session: session,
	}, nil
}
