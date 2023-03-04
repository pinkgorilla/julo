package wallet_test

import (
	"context"
	"julo/internal/wallet"
	"testing"

	"github.com/google/uuid"
)

func TestEnableWallet(t *testing.T) {
	ctx := context.Background()
	service := wallet.NewService(wallet.NewInMemoryRepository())

	xid := uuid.NewString()
	t.Run("enable wallet for first time, should success", func(t *testing.T) {
		wal, err := service.EnableWallet(ctx, wallet.EnableWalletParam{
			OwnerXID: xid,
		})
		if err != nil {
			t.Fatal(err)
		}

		if wal == nil {
			t.Fatal("unexpected value nil for wal")
		}

		t.Run("enable already enabled wallet, should failed", func(t *testing.T) {
			wal2, err := service.EnableWallet(ctx, wallet.EnableWalletParam{
				OwnerXID: xid,
			})
			if err != wallet.ErrWalletEnabled {
				t.Fatalf("expecting error %s,  got %s", wallet.ErrWalletEnabled, err)
			}

			if wal2 != nil {
				t.Fatal("expecting nil value wal")
			}
		})
	})
}

func TestDisableWallet(t *testing.T) {
	ctx := context.Background()
	service := wallet.NewService(wallet.NewInMemoryRepository())

	xid := uuid.NewString()
	t.Run("enable wallet for first time, should success", func(t *testing.T) {
		wal, err := service.EnableWallet(ctx, wallet.EnableWalletParam{
			OwnerXID: xid,
		})
		if err != nil {
			t.Fatal(err)
		}

		if wal == nil {
			t.Fatal("unexpected value nil for wal")
		}

		if wal.Status != wallet.WalletStatusEnabled {
			t.Fatalf("expecting wallet status %s, got %s", wallet.WalletStatusEnabled, wal.Status)
		}

		t.Run("disable already enabled wallet, should success", func(t *testing.T) {
			wal2, err := service.DisableWallet(ctx, wallet.DisableWalletParam{
				OwnerXID: xid,
			})
			if err != nil {
				t.Fatal(err)
			}

			if wal2 == nil {
				t.Fatal("unexpected nil value for wal2")
			}

			if wal2.Status != wallet.WalletStatusDisabled {
				t.Fatalf("expecting wallet status %s, got %s", wallet.WalletStatusDisabled, wal2.Status)
			}
			t.Run("disable already disabled wallet, should fail", func(t *testing.T) {
				wal3, err := service.DisableWallet(ctx, wallet.DisableWalletParam{
					OwnerXID: xid,
				})
				if err != wallet.ErrWalletDisabled {
					t.Fatalf("expecting error %s,  got %s", wallet.ErrWalletDisabled, err)
				}

				if wal3 != nil {
					t.Fatal("expecting nil value wal")
				}
			})
		})
	})
}
