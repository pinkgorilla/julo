package http

import (
	"julo/internal/account"
	"julo/internal/auth"
	httphelper "julo/internal/http"
	"net/http"
)

func InitHandler(initializer auth.Initializer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response httphelper.Response

		customerXID := r.FormValue("customer_xid")
		if customerXID == "" {
			response.Status = "fail"
			response.Data = map[string]interface{}{"error": map[string]interface{}{"customer_xid": "missing data for required field."}}
			httphelper.WriteJSON(w, http.StatusBadRequest, response)
			return
		}

		result, err := initializer.Init(r.Context(), auth.InitParam{
			CustomerXID: customerXID,
		})
		if err != nil && err != account.ErrAccountAlreadyExists {
			httphelper.WriteErrorJSON(w, http.StatusBadRequest, err)
			return
		} else if err != nil {
			httphelper.WriteErrorJSON(w, http.StatusInternalServerError, err)
			return
		}

		response.Status = "success"
		response.Data = struct {
			Token string `json:"token"`
		}{
			Token: result.Session.Token,
		}
		httphelper.WriteJSON(w, http.StatusOK, response)
	})
}
