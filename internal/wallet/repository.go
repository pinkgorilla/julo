package wallet

import (
	"context"
	"sync"
	"time"
)

type Wallet struct {
	ID        string
	OwnerXID  string
	Balance   int
	EnabledAt time.Time
	Status    WalletStatus
}

type WalletStatus string

var (
	WalletStatusEnabled  = WalletStatus("enabled")
	WalletStatusDisabled = WalletStatus("disabled")
)

type WalletTransaction struct {
	ID          string `json:"id"`
	WalletID    string
	ActorXID    string
	ReferenceID string    `json:"reference_id"`
	Type        string    `json:"type"`
	Date        time.Time `json:"transacted_at"`
	Amount      int       `json:"amount"`
	Status      string    `json:"status"`
}

type Repository interface {
	GetWalletByXID(ctx context.Context, xid string) (*Wallet, error)
	CreateWallet(ctx context.Context, wallet Wallet) error
	UpdateWallet(ctx context.Context, wallet Wallet) error
	CreateTransaction(ctx context.Context, t WalletTransaction) error
	GetTransactions(ctx context.Context, walletID string) ([]WalletTransaction, error)
}

type InMemoryRepository struct {
	wallets      sync.Map
	transactions sync.Map
}

func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		wallets: sync.Map{},
	}
}

func (r *InMemoryRepository) CreateTransaction(ctx context.Context, t WalletTransaction) error {
	var transactions []WalletTransaction
	v, ok := r.transactions.Load(t.WalletID)
	if ok {
		transactions = v.([]WalletTransaction)
	} else {
		transactions = []WalletTransaction{}
	}
	transactions = append(transactions, t)
	r.transactions.Store(t.WalletID, transactions)
	return nil
}

func (r *InMemoryRepository) GetTransactions(ctx context.Context, walletID string) ([]WalletTransaction, error) {
	v, ok := r.transactions.Load(walletID)
	if !ok {
		return []WalletTransaction{}, nil
	}

	return v.([]WalletTransaction), nil
}

func (r *InMemoryRepository) CreateWallet(ctx context.Context, wallet Wallet) error {
	r.wallets.Store(wallet.OwnerXID, &wallet)
	return nil
}

func (r *InMemoryRepository) UpdateWallet(ctx context.Context, wallet Wallet) error {
	r.wallets.Store(wallet.OwnerXID, &wallet)
	return nil
}

func (r *InMemoryRepository) GetWalletByXID(ctx context.Context, xid string) (*Wallet, error) {
	v, ok := r.wallets.Load(xid)
	if !ok {
		return nil, ErrWalletNotFound
	}

	return v.(*Wallet), nil
}
