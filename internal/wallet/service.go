package wallet

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type EnableWalletParam struct {
	OwnerXID string
}

type DisableWalletParam struct {
	OwnerXID string
}

type WalletTransactionParam struct {
	ActorXID    string
	OwnerXID    string
	ReferenceID string
	Amount      int
}

func (p WalletTransactionParam) Validate() error {
	var ve ValidationError
	if p.ActorXID == "" {
		ve.AddError("actor_xid", ErrMissingRequiredParameter)
	}
	if p.OwnerXID == "" {
		ve.AddError("owner_xid", ErrMissingRequiredParameter)
	}
	if p.ActorXID == "" {
		ve.AddError("reference_id", ErrMissingRequiredParameter)
	}
	if p.Amount <= 0 {
		ve.AddError("amount", ErrInvalidDepositAmount)
	}
	if len(ve.GetErrors()) > 0 {
		return ve
	}
	return nil
}

type WalletTransactionResult struct {
	ID          string
	DepositedAt time.Time
	DepositedBy string
	Amount      int
	Status      string
	ReferenceID string
}

type GetWalletTransactionsParam struct {
	WalletID string
}

type GetWalletTransactionsResult struct {
	Transactions []WalletTransaction
}

type Service interface {
	GetWalletByXID(ctx context.Context, xid string) (*Wallet, error)
	EnableWallet(ctx context.Context, param EnableWalletParam) (*Wallet, error)
	DisableWallet(ctx context.Context, param DisableWalletParam) (*Wallet, error)
	DepositWallet(ctx context.Context, param WalletTransactionParam) (*WalletTransactionResult, error)
	WithdrawWallet(ctx context.Context, param WalletTransactionParam) (*WalletTransactionResult, error)
	GetWalletTransactions(ctx context.Context, param GetWalletTransactionsParam) (*GetWalletTransactionsResult, error)
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{
		repo: r,
	}
}

func (s *service) DepositWallet(ctx context.Context, param WalletTransactionParam) (*WalletTransactionResult, error) {
	err := param.Validate()
	if err != nil {
		return nil, err
	}

	wal, err := s.repo.GetWalletByXID(ctx, param.OwnerXID)
	if err != nil && err != ErrWalletNotFound {
		return nil, errors.Wrap(err, "failed getting wallet")
	}

	if wal.Status == WalletStatusDisabled {
		return nil, ErrWalletDisabled
	}

	trx := WalletTransaction{
		ID:          uuid.NewString(),
		ActorXID:    param.ActorXID,
		WalletID:    wal.ID,
		ReferenceID: param.ReferenceID,
		Type:        "deposit",
		Date:        time.Now(),
		Amount:      param.Amount,
		Status:      "success",
	}

	err = s.repo.CreateTransaction(ctx, trx)
	if err != nil {
		return nil, errors.Wrap(err, "failed updating wallet")
	}

	wal.Balance += trx.Amount
	err = s.repo.UpdateWallet(ctx, *wal)
	if err != nil {
		return nil, errors.Wrap(err, "failed updating wallet")
	}

	return &WalletTransactionResult{
		ID:          trx.ID,
		DepositedAt: trx.Date,
		DepositedBy: trx.ActorXID,
		Amount:      trx.Amount,
		Status:      trx.Status,
		ReferenceID: trx.ReferenceID,
	}, nil
}

func (s *service) WithdrawWallet(ctx context.Context, param WalletTransactionParam) (*WalletTransactionResult, error) {
	err := param.Validate()
	if err != nil {
		return nil, err
	}

	wal, err := s.repo.GetWalletByXID(ctx, param.OwnerXID)
	if err != nil && err != ErrWalletNotFound {
		return nil, errors.Wrap(err, "failed getting wallet")
	}

	if wal.Status == WalletStatusDisabled {
		return nil, ErrWalletDisabled
	}

	if wal.Balance < param.Amount {
		return nil, ErrInsufficientBalance
	}

	trx := WalletTransaction{
		ID:          uuid.NewString(),
		ActorXID:    param.ActorXID,
		WalletID:    wal.ID,
		ReferenceID: param.ReferenceID,
		Type:        "withdrawal",
		Date:        time.Now(),
		Amount:      param.Amount,
		Status:      "success",
	}

	err = s.repo.CreateTransaction(ctx, trx)
	if err != nil {
		return nil, errors.Wrap(err, "failed updating wallet")
	}

	wal.Balance -= trx.Amount
	err = s.repo.UpdateWallet(ctx, *wal)
	if err != nil {
		return nil, errors.Wrap(err, "failed updating wallet")
	}

	return &WalletTransactionResult{
		ID:          trx.ID,
		DepositedAt: trx.Date,
		DepositedBy: trx.ActorXID,
		Amount:      trx.Amount,
		Status:      trx.Status,
		ReferenceID: trx.ReferenceID,
	}, nil
}

func (s *service) EnableWallet(ctx context.Context, param EnableWalletParam) (*Wallet, error) {
	wal, err := s.repo.GetWalletByXID(ctx, param.OwnerXID)
	if err != nil && err != ErrWalletNotFound {
		return nil, errors.Wrap(err, "failed getting wallet")
	}

	if wal == nil {
		wal = &Wallet{
			ID:       uuid.NewString(),
			OwnerXID: param.OwnerXID,
			Status:   WalletStatusDisabled,
			Balance:  0,
		}
		err = s.repo.CreateWallet(ctx, *wal)
		if err != nil {
			return nil, errors.Wrap(err, "failed creating wallet")
		}
	}
	if wal.Status == WalletStatusEnabled {
		return nil, ErrWalletEnabled
	}
	wal.Status = WalletStatusEnabled
	wal.EnabledAt = time.Now()

	err = s.repo.UpdateWallet(ctx, *wal)
	if err != nil {
		return nil, errors.Wrap(err, "failed updating wallet")
	}

	return wal, nil
}

func (s *service) DisableWallet(ctx context.Context, param DisableWalletParam) (*Wallet, error) {
	wal, err := s.repo.GetWalletByXID(ctx, param.OwnerXID)
	if err != nil && err == ErrWalletNotFound {
		return nil, ErrWalletNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "failed getting wallet")
	}

	if wal.Status == WalletStatusDisabled {
		return nil, ErrWalletDisabled
	}
	wal.Status = WalletStatusDisabled

	err = s.repo.UpdateWallet(ctx, *wal)
	if err != nil {
		return nil, errors.Wrap(err, "failed updating wallet")
	}

	return wal, nil
}

func (s *service) GetWalletByXID(ctx context.Context, xid string) (*Wallet, error) {
	return s.repo.GetWalletByXID(ctx, xid)
}

func (s *service) GetWalletTransactions(ctx context.Context, param GetWalletTransactionsParam) (*GetWalletTransactionsResult, error) {
	transactions, err := s.repo.GetTransactions(ctx, param.WalletID)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting wallet transactions")
	}

	return &GetWalletTransactionsResult{
		Transactions: transactions,
	}, nil
}
