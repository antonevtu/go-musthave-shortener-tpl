package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const addr = "localhost:8080"
const idLen = 5

type Storage map[string]string

var storageLock sync.Mutex
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func Run() {
	storage := make(Storage, 100)
	r := NewRouter(storage)
	log.Fatal(http.ListenAndServe(addr, r))
}

func NewRouter(base Storage) chi.Router {
	// Определяем роутер chi
	r := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// создадим суброутер, который будет содержать две функции
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlerShortenURL(base))
		r.Get("/{id}", handlerRecoverURL(base))
	})
	return r
}

func handlerShortenURL(base Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		urlString := string(body)

		id := shortenURL(urlString, base)
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

func handlerRecoverURL(base Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storageLock.Lock()
		defer storageLock.Unlock()
		id := chi.URLParam(r, "id")
		longURL, ok := base[id]
		if !ok {
			http.Error(w, "А nonexistent ID was requested", http.StatusBadRequest)
		}
		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func shortenURL(urlString string, base Storage) (id string) {
	storageLock.Lock()
	defer storageLock.Unlock()
	for {
		id = randStringRunes(idLen)
		if _, ok := base[id]; !ok {
			base[id] = urlString
			break
		}
	}
	return id
}

func randStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
