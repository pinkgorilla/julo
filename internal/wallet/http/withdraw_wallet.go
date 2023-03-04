package http

import (
	"julo/internal/auth"
	httphelper "julo/internal/http"
	"julo/internal/wallet"
	"net/http"
	"strconv"
	"time"
)

func WithdrawWalletHandler(wallets wallet.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response httphelper.Response
		session := auth.SessionFromContext(r.Context())
		if session == nil {
			httphelper.WriteErrorJSON(w, http.StatusUnauthorized, auth.ErrSessionNotFound)
			return
		}

		refid := r.FormValue("reference_id")
		iamount, err := strconv.ParseInt(r.FormValue("amount"), 10, 32)
		if err != nil {
			httphelper.WriteErrorJSON(w, http.StatusBadRequest, err)
			return
		}

		wal, err := wallets.GetWalletByXID(r.Context(), session.Account.XID)
		if err != nil && err == wallet.ErrWalletNotFound {
			httphelper.WriteErrorJSON(w, http.StatusBadRequest, wallet.ErrWalletDisabled)
			return
		} else if err != nil {
			httphelper.WriteErrorJSON(w, http.StatusInternalServerError, err)
			return
		}

		if wal == nil || wal.Status == wallet.WalletStatusDisabled {
			httphelper.WriteErrorJSON(w, http.StatusBadRequest, wallet.ErrWalletDisabled)
			return
		}

		result, err := wallets.WithdrawWallet(r.Context(), wallet.WalletTransactionParam{
			ActorXID:    session.Account.XID,
			OwnerXID:    wal.OwnerXID,
			ReferenceID: refid,
			Amount:      int(iamount),
		})
		if err != nil {
			ve, ok := err.(wallet.ValidationError)
			if ok {
				response.Status = "failed"
				response.Data = ve.GetErrors()
				httphelper.WriteJSON(w, http.StatusBadRequest, response)
			} else {
				httphelper.WriteErrorJSON(w, http.StatusBadRequest, err)
			}
			return
		}

		response.Status = "success"
		response.Data = map[string]interface{}{
			"withdrawal": struct {
				ID          string    `json:"id"`
				Depositedby string    `json:"deposited_by"`
				Status      string    `json:"status"`
				DepositedAt time.Time `json:"deposited_at"`
				Amount      int       `json:"amount"`
				ReferenceID string    `json:"reference_id"`
			}{
				ID:          result.ID,
				Depositedby: result.DepositedBy,
				Status:      result.Status,
				DepositedAt: result.DepositedAt,
				Amount:      result.Amount,
				ReferenceID: result.ReferenceID,
			},
		}
		httphelper.WriteJSON(w, http.StatusOK, response)
	})
}
