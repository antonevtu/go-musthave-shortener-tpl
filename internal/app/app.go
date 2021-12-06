package app

import (
	"errors"
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

type Repository struct {
	Storage     storageT
	storageLock sync.Mutex
}
type storageT map[string]string

type Repositorier interface {
	Store(url string) (string, error)
	Load(shortURL string) (string, error)
}

func Run() {
	var repository Repository
	repository.Init()
	r := NewRouter(&repository)
	log.Fatal(http.ListenAndServe(addr, r))
}

func (r *Repository) Init() {
	r.Storage = make(storageT, 100)
}

func NewRouter(repository Repositorier) chi.Router {
	// Определяем роутер chi
	r := chi.NewRouter()

	// зададим встроенные middleware, чтобы улучшить стабильность приложения
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// создадим суброутер, который будет содержать две функции
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlerStoreURL(repository))
		r.Get("/{id}", handlerLoadURL(repository))
	})
	return r
}

func handlerStoreURL(repository Repositorier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		urlString := string(body)

		id, err := repository.Store(urlString)
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

func handlerLoadURL(repository Repositorier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		longURL, err := repository.Load(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (r *Repository) Load(id string) (string, error) {
	r.storageLock.Lock()
	defer r.storageLock.Unlock()
	longURL, ok := r.Storage[id]
	if ok {
		return longURL, nil
	} else {
		return longURL, errors.New("А nonexistent ID was requested")
	}
}

func (r *Repository) Store(url string) (string, error) {
	const idLen = 5
	const attemptsNumber = 10
	r.storageLock.Lock()
	defer r.storageLock.Unlock()
	for i := 0; i < attemptsNumber; i++ {
		id := randStringRunes(idLen)
		if _, ok := r.Storage[id]; !ok {
			r.Storage[id] = url
			return id, nil
		}
	}
	return "", errors.New("Can't generate random ID")
}

func randStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
