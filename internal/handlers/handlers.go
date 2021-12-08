package handlers

import (
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"net/http"
)

func NewRouter(repo repository.Repositorier) chi.Router {
	// Определяем роутер chi
	r := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// создадим суброутер, который будет содержать две функции
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlerStoreURL(repo))
		r.Get("/{id}", handlerLoadURL(repo))
	})
	return r
}

func handlerStoreURL(repo repository.Repositorier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		urlString := string(body)

		id, err := repo.Store(urlString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		shortURL := "http://" + r.Host + "/" + id

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handlerLoadURL(repo repository.Repositorier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		longURL, err := repo.Load(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
