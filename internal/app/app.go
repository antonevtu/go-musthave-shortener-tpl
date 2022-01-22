package app

import (
	"context"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/cfg"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/db"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/handlers"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/repository"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	var cfgApp, err = cfg.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// database
	dbPool, err := db.New(ctx, cfgApp.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	// repository
	repo, err := repository.New(cfgApp.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	// delete thread
	go deleteLoop(ctx, &dbPool, cfgApp.ToDeleteChan)

	//r := handlers.NewRouter(repo, cfgApp)
	r := handlers.NewRouter(&dbPool, cfgApp)
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

func deleteLoop(ctx context.Context, repo handlers.Repositorier, input chan cfg.ToDeleteItem) {
	for {
		select {
		case item := <-input:
			_ = repo.SetDeleted(ctx, item)
		case <-ctx.Done():
			return
		}
	}
}
