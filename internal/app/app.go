package app

import (
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const idLen = 5

type baseT map[string]string

var base baseT
var baseLock sync.Mutex
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func Run() {

	base = make(baseT, 100)
	initRand()
	http.HandleFunc("/", handler)
	addr := "localhost:8080"
	log.Println("Запуск сервера:", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		urlString := string(body)
		shortUrl, err := shortenUrl(urlString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(shortUrl))
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
		longUrl := base[id]
		if longUrl == "" {
			http.Error(w, "А nonexistent ID was requested", http.StatusBadRequest)
		}
		w.Header().Set("Location", longUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)

	default:
		http.Error(w, "Only POST or GET requests are allowed!", http.StatusMethodNotAllowed)
	}
}

func shortenUrl(urlString string) (shortUrl string, err error) {
	if urlString == "" {
		return "", errors.New("empty URL string")
	}

	var id string
	baseLock.Lock()
	defer baseLock.Unlock()
	for {
		id = randStringRunes(idLen)
		if base[id] == "" {
			base[id] = urlString
			break
		}
	}

	shortUrl = "http://localhost:8080/" + id
	return shortUrl, err
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
