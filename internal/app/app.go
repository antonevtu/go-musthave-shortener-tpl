package app

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const addr = "localhost:8080"
const idLen = 5

type baseT map[string]string

var baseLock sync.Mutex
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func Run() {
	base := make(baseT, 100)
	initRand()
	http.HandleFunc("/", handler(base))
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handler(base baseT) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			urlString := string(body)

			shortURL, err := shortenURL(urlString, base)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusCreated)
			_, err = w.Write([]byte(shortURL))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		case "GET":
			path := r.URL.Path
			if len(path) < idLen+1 {
				http.Error(w, "The ID is missing", http.StatusBadRequest)
				return
			}
			id := path[1:]
			longURL, ok := base[id]
			if !ok {
				http.Error(w, "А nonexistent ID was requested", http.StatusBadRequest)
			}
			w.Header().Set("Location", longURL)
			w.WriteHeader(http.StatusTemporaryRedirect)

		default:
			http.Error(w, "Only POST or GET requests are allowed!", http.StatusMethodNotAllowed)
		}
	}
}

func shortenURL(urlString string, base baseT) (shortURL string, err error) {
	var id string
	baseLock.Lock()
	defer baseLock.Unlock()
	for {
		id = randStringRunes(idLen)
		if _, ok := base[id]; !ok {
			base[id] = urlString
			break
		}
	}
	shortURL = "http://" + addr + "/" + id
	return shortURL, err
}

func initRand() {
	rand.Seed(time.Now().UnixNano())
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
