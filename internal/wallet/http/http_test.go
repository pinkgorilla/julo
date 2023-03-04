package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"julo/internal/account"
	"julo/internal/auth"
	authhttp "julo/internal/auth/http"
	httphelper "julo/internal/http"
	"julo/internal/wallet"
	wallethttp "julo/internal/wallet/http"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func TestWallet(t *testing.T) {
	accounts := account.NewService(account.NewInMemoryRepository())
	initializer := auth.NewInitializer(accounts)
	wallets := wallet.NewService(wallet.NewInMemoryRepository())

	router := chi.NewRouter()
	router.Mount("/api/v1", router.Group(func(r chi.Router) {
		r.Post("/init", authhttp.InitHandler(initializer).ServeHTTP)
		r.Mount("/wallet", r.Group(func(r chi.Router) {
			r.Use(authhttp.Middleware)
			r.Get("/", wallethttp.ViewWalletBalanceHandler(wallets).ServeHTTP)
			r.Post("/", wallethttp.EnableWalletHandler(wallets).ServeHTTP)
			r.Patch("/", wallethttp.DisableWalletHandler(wallets).ServeHTTP)
			r.Post("/deposits", wallethttp.DepositWalletHandler(wallets).ServeHTTP)
			r.Post("/withdrawals", wallethttp.WithdrawWalletHandler(wallets).ServeHTTP)
			r.Get("/transactions", wallethttp.ViewWalletTransactionsHandler(wallets).ServeHTTP)
		}))
	}))

	server := httptest.NewServer(router)
	defer server.Close()
	baseUrl := server.URL

	t.Run("init account", func(t *testing.T) {
		xid := uuid.NewString()
		form := url.Values{}
		form.Set("customer_xid", xid)

		req := buildAuthenticatedRequest(t, http.MethodPost, baseUrl+"/api/v1/init", "", bytes.NewBufferString(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		var response httphelper.Response
		err = json.NewDecoder(res.Body).Decode(&response)
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

		t.Run("get balance when wallet not enabled, should fail", func(t *testing.T) {
			req := buildAuthenticatedRequest(t, http.MethodGet, baseUrl+"/api/v1/wallet", token, nil)
			res, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("expecting status %v, got %v", http.StatusBadRequest, res.StatusCode)
			}
		})

		t.Run("enable wallet, should success", func(t *testing.T) {
			req := buildAuthenticatedRequest(t, http.MethodPost, baseUrl+"/api/v1/wallet", token, nil)
			res, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode != http.StatusOK {
				t.Fatalf("expecting status %v, got %v", http.StatusOK, res.StatusCode)
			}

			t.Run("get balance after wallet enabled, should success", func(t *testing.T) {
				req := buildAuthenticatedRequest(t, http.MethodGet, baseUrl+"/api/v1/wallet", token, nil)
				res, err := server.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}

				if res.StatusCode != http.StatusOK {
					t.Fatalf("expecting status %v, got %v", http.StatusOK, res.StatusCode)
				}
			})

			t.Run("deposit wallet after wallet enabled, should success", func(t *testing.T) {
				refid := uuid.NewString()
				amount := 100000
				form := url.Values{}
				form.Set("reference_id", refid)
				form.Set("amount", fmt.Sprint(amount))
				req := buildAuthenticatedRequest(t, http.MethodPost, baseUrl+"/api/v1/wallet/deposits", token, bytes.NewBufferString(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				res, err := server.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}

				if res.StatusCode != http.StatusOK {
					t.Fatalf("expecting status %v, got %v", http.StatusOK, res.StatusCode)
				}

				t.Run("withdraw wallet after deposit, should success", func(t *testing.T) {
					refid := uuid.NewString()
					amount := 50000
					form := url.Values{}
					form.Set("reference_id", refid)
					form.Set("amount", fmt.Sprint(amount))
					req := buildAuthenticatedRequest(t, http.MethodPost, baseUrl+"/api/v1/wallet/withdrawals", token, bytes.NewBufferString(form.Encode()))
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

					res, err := server.Client().Do(req)
					if err != nil {
						t.Fatal(err)
					}

					if res.StatusCode != http.StatusOK {
						t.Fatalf("expecting status %v, got %v", http.StatusOK, res.StatusCode)
					}

					t.Run("get balance after withdrawal, should have 50000", func(t *testing.T) {
						req := buildAuthenticatedRequest(t, http.MethodGet, baseUrl+"/api/v1/wallet", token, nil)
						res, err := server.Client().Do(req)
						if err != nil {
							t.Fatal(err)
						}

						if res.StatusCode != http.StatusOK {
							t.Fatalf("expecting status %v, got %v", http.StatusOK, res.StatusCode)
						}

						var response httphelper.Response
						err = json.NewDecoder(res.Body).Decode(&response)
						if err != nil {
							t.Fatal(err)
						}

						if response.Status != "success" {
							t.Fatalf("expecting response.Status %s, got %s", "success", response.Status)
						}

						data := response.Data.(map[string]interface{})
						vwallet, ok := data["wallet"]
						if !ok {
							t.Fatal("wallet not found")
						}
						walletdata, ok := vwallet.(map[string]interface{})
						if !ok {
							t.Fatal("balance not found")
						}
						balance, err := strconv.ParseInt(fmt.Sprint(walletdata["balance"]), 10, 32)
						if err != nil {
							t.Fatal(err)
						}
						if balance != 50000 {
							t.Fatal("should 5000")
						}
					})
				})
				t.Run("2nd withdrawal with big amount, should fail", func(t *testing.T) {
					refid := uuid.NewString()
					amount := 60000
					form := url.Values{}
					form.Set("reference_id", refid)
					form.Set("amount", fmt.Sprint(amount))
					req := buildAuthenticatedRequest(t, http.MethodPost, baseUrl+"/api/v1/wallet/withdrawals", token, bytes.NewBufferString(form.Encode()))
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

					res, err := server.Client().Do(req)
					if err != nil {
						t.Fatal(err)
					}

					if res.StatusCode != http.StatusBadRequest {
						t.Fatalf("expecting status %v, got %v", http.StatusBadRequest, res.StatusCode)
					}

				})

				t.Run("get transactions, should success", func(t *testing.T) {
					req := buildAuthenticatedRequest(t, http.MethodGet, baseUrl+"/api/v1/wallet/transactions", token, nil)
					res, err := server.Client().Do(req)
					if err != nil {
						t.Fatal(err)
					}

					if res.StatusCode != http.StatusOK {
						t.Fatalf("expecting status %v, got %v", http.StatusOK, res.StatusCode)
					}
				})
			})

			t.Run("disable wallet when it's enabled, should success", func(t *testing.T) {
				form := url.Values{}
				form.Set("is_disabled", "true")
				req := buildAuthenticatedRequest(t, http.MethodPatch, baseUrl+"/api/v1/wallet", token, bytes.NewBufferString(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				res, err := server.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}

				if res.StatusCode != http.StatusOK {
					t.Fatalf("expecting status %v, got %v", http.StatusOK, res.StatusCode)
				}

				t.Run("get balance when wallet disabled, should failed", func(t *testing.T) {
					req := buildAuthenticatedRequest(t, http.MethodGet, baseUrl+"/api/v1/wallet", token, nil)
					res, err := server.Client().Do(req)
					if err != nil {
						t.Fatal(err)
					}

					if res.StatusCode != http.StatusBadRequest {
						t.Fatalf("expecting status %v, got %v", http.StatusBadRequest, res.StatusCode)
					}
				})
			})
		})
	})
}

func buildAuthenticatedRequest(t *testing.T, method string, url string, token string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
	return req
}
