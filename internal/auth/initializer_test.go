package auth_test

import (
	"context"
	"julo/internal/account"
	"julo/internal/auth"
	"testing"

	"github.com/google/uuid"
)

type CreateAccountFunc func(context.Context, account.Account) error

func (f CreateAccountFunc) CreateAccount(c context.Context, account account.Account) error {
	return f(c, account)
}

func TestInitializer(t *testing.T) {
	c := context.Background()
	accounts := account.NewService(account.NewInMemoryRepository())
	initializer := auth.NewInitializer(accounts)

	t.Run("initialize", func(t *testing.T) {
		xid := uuid.NewString()
		result, err := initializer.Init(c, auth.InitParam{
			CustomerXID: xid,
		})
		if err != nil {
			t.Fatal(err)
		}

		if result == nil {
			t.Fatal("unexpected nil value for result")
		}

		t.Run("get session from initialized customer", func(t *testing.T) {
			session, err := auth.GetSession(c, result.Session.Token)
			if err != nil {
				t.Fatal(err)
			}
			if session == nil {
				t.Fatal("unexpected nil value for session")
			}
		})
	})
}

func TestSessionManager(t *testing.T) {
	c := context.Background()
	t.Run("get inexist session", func(t *testing.T) {
		result, err := auth.GetSession(c, "not-exist-token")
		if err != nil && err != auth.ErrSessionNotFound {
			t.Fatal("unexptected error", err)
		}

		if err == nil {
			t.Fatal("unexpected nil value for err")
		}

		if result != nil {
			t.Fatal("expecting nil value for result")
		}
	})
}
