package account

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestCreateAccount(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryRepository()
	service := NewService(repo)

	t.Run("create account, should success", func(t *testing.T) {
		xid := uuid.NewString()
		account := Account{
			XID: xid,
		}
		err := service.CreateAccount(ctx, account)
		if err != nil {
			t.Fatal(err)
		}

		t.Run("get created account, should success", func(t *testing.T) {

			acc, err := service.GetAccount(ctx, xid)
			if err != nil {
				t.Fatal(err)
			}
			if acc == nil {
				t.Fatal("unexpected nil value for acc")
			}
			if acc.XID != xid {
				t.Fatalf("expecting xid with value %s, got %s", xid, acc.XID)
			}
		})

		t.Run("create account with same xid, should failed", func(t *testing.T) {
			account := Account{
				XID: xid,
			}
			err := service.CreateAccount(ctx, account)
			if err != ErrAccountAlreadyExists {
				t.Fatalf("expecting error %s, got %s", ErrAccountAlreadyExists, err)
			}
		})
	})
}
