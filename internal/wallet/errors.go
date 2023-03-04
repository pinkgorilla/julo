package wallet

import "errors"

var (
	ErrWalletNotFound           = errors.New("wallet not found")
	ErrWalletEnabled            = errors.New("wallet is enabled")
	ErrWalletDisabled           = errors.New("wallet is disabled")
	ErrMissingRequiredParameter = errors.New("missing required parameter")
	ErrInvalidDepositAmount     = errors.New("invalid deposit amount")
	ErrInsufficientBalance      = errors.New("insufficient balance")
)

type ValidationError struct {
	messages map[string]error
}

func (e ValidationError) Error() string {
	return "validation error"
}

func (e ValidationError) AddError(field string, err error) {
	e.messages[field] = err
}

func (e ValidationError) GetErrors() map[string]error {
	return e.messages
}
