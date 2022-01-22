package handlers

import (
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/cfg"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(repo Repositorier, cfgApp cfg.Config) chi.Router {
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
		r.Post("/", handlerShortenURL(repo, cfgApp))
		r.Post("/api/shorten", handlerShortenURLJSONAPI(repo, cfgApp))
		r.Get("/{id}", handlerExpandURL(repo, cfgApp))
		r.Get("/user/urls", handlerUserHistory(repo, cfgApp))
		r.Get("/ping", handlerPingDB(repo))
		r.Post("/api/shorten/batch", handlerShortenURLAPIBatch(repo, cfgApp))
		r.Delete("/api/user/urls", handlerDelete(repo, cfgApp))
	})
	return r
}
