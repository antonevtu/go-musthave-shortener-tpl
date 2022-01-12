package app

import (
	"context"
	"fmt"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/cfg"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/db"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/handlers"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/repository"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	cfgApp, err := cfg.New()
	if err != nil {
		log.Fatal(err)
	}

	// database
	db.DbPool, err = pgxpool.Connect(context.Background(), cfgApp.DatabaseDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.DbPool.Close()

	// repository
	repo, err := repository.New(cfgApp.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	r := handlers.NewRouter(repo, cfgApp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServer := &http.Server{
		Addr:        cfgApp.ServerAddress,
		Handler:     r,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	// Run server
	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	<-signalChan
	cancel()
	log.Print("os.Interrupt - shutting down...\n")

	gracefullCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err = httpServer.Shutdown(gracefullCtx); err != nil {
		log.Printf("shutdown error: %v\n", err)
	} else {
		log.Printf("gracefully stopped\n")
	}
}
