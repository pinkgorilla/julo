package http

import (
	"julo/internal/auth"
	httphelper "julo/internal/http"
	"julo/internal/wallet"
	"net/http"
	"strconv"
	"time"
)

func DisableWalletHandler(wallets wallet.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wal *wallet.Wallet
		var err error
		var response httphelper.Response

		session := auth.SessionFromContext(r.Context())
		if session == nil {
			httphelper.WriteErrorJSON(w, http.StatusUnauthorized, auth.ErrSessionNotFound)
			return
		}

		isDisabled, err := strconv.ParseBool(r.FormValue("is_disabled"))
		if err != nil {
			httphelper.WriteErrorJSON(w, http.StatusUnprocessableEntity, err)
			return
		}

		if !isDisabled {
			wal, err = wallets.EnableWallet(r.Context(), wallet.EnableWalletParam{
				OwnerXID: session.Account.XID,
			})
			if err != nil && err == wallet.ErrWalletEnabled {
				httphelper.WriteErrorJSON(w, http.StatusBadRequest, err)
				return
			} else if err != nil {
				httphelper.WriteErrorJSON(w, http.StatusInternalServerError, err)
				return
			}
		} else {
			wal, err = wallets.DisableWallet(r.Context(), wallet.DisableWalletParam{
				OwnerXID: session.Account.XID,
			})
			if err != nil && err == wallet.ErrWalletDisabled || err == wallet.ErrWalletNotFound {
				httphelper.WriteErrorJSON(w, http.StatusBadRequest, wallet.ErrWalletDisabled)
				return
			} else if err != nil {
				httphelper.WriteErrorJSON(w, http.StatusInternalServerError, err)
				return
			}
		}

		response.Status = "success"
		response.Data = map[string]interface{}{
			"wallet": struct {
				ID        string    `json:"id"`
				OwnedBy   string    `json:"owned_by"`
				Status    string    `json:"status"`
				EnabledAt time.Time `json:"enabled_at"`
				Balance   int       `json:"balance"`
			}{
				ID:        wal.ID,
				OwnedBy:   wal.OwnerXID,
				Status:    string(wal.Status),
				EnabledAt: wal.EnabledAt,
				Balance:   wal.Balance,
			},
		}
		httphelper.WriteJSON(w, http.StatusOK, response)
	})
}

type WalletResponse struct {
	Status string              `json:"status"`
	Data   *WalletResponseData `json:"data"`
}

type WalletResponseData struct {
	Withdrawal string
}
