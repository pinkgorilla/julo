package http

import (
	"julo/internal/auth"
	httphelper "julo/internal/http"
	"julo/internal/wallet"
	"net/http"
)

func ViewWalletTransactionsHandler(wallets wallet.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response httphelper.Response
		session := auth.SessionFromContext(r.Context())
		if session == nil {
			httphelper.WriteErrorJSON(w, http.StatusUnauthorized, auth.ErrSessionNotFound)
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

		result, err := wallets.GetWalletTransactions(r.Context(), wallet.GetWalletTransactionsParam{
			WalletID: wal.ID,
		})
		if err != nil {
			httphelper.WriteErrorJSON(w, http.StatusBadRequest, err)
			return
		}

		response.Status = "success"
		response.Data = map[string]interface{}{
			"transactions": result.Transactions,
		}
		httphelper.WriteJSON(w, http.StatusOK, response)
	})
}
