package app

import (
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/handlers"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/repository"
	"log"
	"net/http"
)

const addr = "localhost:8080"

func Run() {
	var repo repository.Repository
	repo.Init()
	r := handlers.NewRouter(&repo)
	log.Fatal(http.ListenAndServe(addr, r))
}
