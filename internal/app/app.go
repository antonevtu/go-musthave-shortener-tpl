package app

import (
	"context"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/cfg"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/db"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/handlers"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/pool"
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

	// file&map repository
	fileRepo, err := repository.New(cfgApp.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fileRepo.Close()

	repo := &dbPool
	//repo := fileRepo

	// repository pool for delete items (set flag "deleted")
	deleterPool := pool.New(ctx, repo)
	defer deleterPool.Close()
	cfgApp.DeleterChan = deleterPool.Input

	//r := handlers.NewRouter(repo, cfgApp)
	r := handlers.NewRouter(repo, cfgApp)
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

	select {
	case <-signalChan:
		log.Println("os.Interrupt - shutting down...")
	case err := <-deleterPool.ErrCh:
		log.Println(err)
	}
	cancel()

	gracefulCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err = httpServer.Shutdown(gracefulCtx); err != nil {
		log.Printf("shutdown error: %v\n", err)
	} else {
		log.Printf("web server gracefully stopped\n")
	}
}
