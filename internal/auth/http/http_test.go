package http_test

import (
	"encoding/json"
	"fmt"
	"julo/internal/account"
	"julo/internal/auth"
	authhttp "julo/internal/auth/http"
	httphelper "julo/internal/http"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
)

func TestInit(t *testing.T) {
	accounts := account.NewService(account.NewInMemoryRepository())
	initializer := auth.NewInitializer(accounts)
	inithandler := authhttp.InitHandler(initializer)

	t.Run("init for first time, should success", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", nil)
		xid := uuid.NewString()
		form := url.Values{}
		form.Add("customer_xid", xid)
		req.Form = form
		inithandler.ServeHTTP(rec, req)

		if rec.Result().StatusCode != http.StatusOK {
			t.Fatal("status")
		}
		var response httphelper.Response
		err := json.NewDecoder(rec.Result().Body).Decode(&response)
		if err != nil {
			t.Fatal(err)
		}

		if response.Status != "success" {
			t.Fatalf("expecting response.Status %s, got %s", "success", response.Status)
		}
		data := response.Data.(map[string]interface{})
		vtoken, ok := data["token"]
		if !ok {
			t.Fatal("token not found")
		}
		token := vtoken.(string)

		t.Run("call init with the same xid, should failed", func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", nil)
			form := url.Values{}
			form.Add("customer_xid", xid)
			req.Form = form
			inithandler.ServeHTTP(rec, req)

			if rec.Result().StatusCode != http.StatusBadRequest {
				t.Fatal("status")
			}
		})

		t.Run("test authorized endpoint", func(t *testing.T) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				session := auth.SessionFromContext(r.Context())
				if session == nil {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			})
			server := httptest.NewServer(authhttp.Middleware(h))
			defer server.Close()

			t.Run("call with valid token, should success", func(t *testing.T) {
				req, err := http.NewRequest("GET", server.URL, nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))

				res, err := server.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}

				if res.StatusCode != http.StatusOK {
					t.Fatalf("expecting status %v, got %v", http.StatusOK, res.StatusCode)
				}
			})

			t.Run("call with invalid token, should success", func(t *testing.T) {
				req, err := http.NewRequest("GET", server.URL, nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("Authorization", fmt.Sprintf("Token %s", "invalid-token"))

				res, err := server.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}

				if res.StatusCode != http.StatusUnauthorized {
					t.Fatalf("expecting status %v, got %v", http.StatusOK, res.StatusCode)
				}
			})
		})
	})

	t.Run("init for first time with empty xid, should fail", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", nil)
		xid := ""
		form := url.Values{}
		form.Add("customer_xid", xid)
		req.Form = form
		inithandler.ServeHTTP(rec, req)

		if rec.Result().StatusCode != http.StatusBadRequest {
			t.Fatal("status")
		}
	})
}
