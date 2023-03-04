package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"julo/internal/account"
	"julo/internal/auth"
	authhttp "julo/internal/auth/http"
	"julo/internal/wallet"
	wallethttp "julo/internal/wallet/http"

	"github.com/go-chi/chi"
)

func main() {
	router := chi.NewRouter()

	accounts := account.NewService(account.NewInMemoryRepository())
	initializer := auth.NewInitializer(accounts)
	wallets := wallet.NewService(wallet.NewInMemoryRepository())

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

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		// service connections
		log.Println("Server listening on localhost:8080 ...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	<-ctx.Done()
	log.Println("timeout of 5 seconds.")
	log.Println("Server exiting")
}
