package http

import (
	"julo/internal/auth"
	httphelper "julo/internal/http"
	"julo/internal/wallet"
	"net/http"
	"time"
)

func EnableWalletHandler(wallets wallet.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response httphelper.Response
		session := auth.SessionFromContext(r.Context())
		if session == nil {
			httphelper.WriteErrorJSON(w, http.StatusUnauthorized, auth.ErrSessionNotFound)
			return
		}

		wal, err := wallets.EnableWallet(r.Context(), wallet.EnableWalletParam{
			OwnerXID: session.Account.XID,
		})
		if err != nil && err == wallet.ErrWalletEnabled {
			httphelper.WriteErrorJSON(w, http.StatusBadRequest, err)
			return
		} else if err != nil {
			httphelper.WriteErrorJSON(w, http.StatusInternalServerError, err)
			return
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
