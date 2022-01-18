package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(repo Repositorier, baseURL string) chi.Router {
	// Определяем роутер chi
	r := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// архивирование запроса/ответа gzip
	r.Use(gzipResponseHandle)
	r.Use(gzipRequestHandle)

	// создадим суброутер
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlerShortenURL(repo, baseURL))
		r.Post("/api/shorten", handlerShortenURLJSONAPI(repo, baseURL))
		r.Get("/{id}", handlerExpandURL(repo))
		r.Get("/user/urls", handlerUserHistory(repo, baseURL))
		r.Get("/ping", handlerPingDB(repo))
		r.Post("/api/shorten/batch", handlerShortenURLAPIBatch(repo, baseURL))
	})
	return r
}
