package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

type Repositorier interface {
	AddEntity(ctx context.Context, entity db.Entity) error
	SelectByLongURL(ctx context.Context, longURL string) (string, error)
	SelectByIDURL(ctx context.Context, shortURL string) (string, error)
	SelectByUser(ctx context.Context, userID string) ([]db.Entity, error)
	Flush(ctx context.Context, userID string, input db.BatchInput) error
	Ping(ctx context.Context) error
}

type requestURL struct {
	URL string `json:"url"`
}
type responseURL struct {
	Result string `json:"result"`
}
type responseUserHistory []item
type item struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

const timeout = 600

func handlerShortenURLJSONAPI(repo Repositorier, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		longURL := requestURL{}
		err = json.Unmarshal(body, &longURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if longURL.URL == "" {
			http.Error(w, `no key "url" or empty request`, http.StatusBadRequest)
			return
		}

		id := uuid.NewString() // ID короткого URL

		// запрос в БД на сохранение URL. Замена id на существующий в случае дублирования longURL
		var statusCode = http.StatusCreated
		ctx, cancel := context.WithTimeout(r.Context(), timeout*time.Second)
		defer cancel()
		err = repo.AddEntity(ctx, db.Entity{UserID: userID.String(), ID: id, URL: longURL.URL})
		if errors.Is(err, db.UniqueViolationError) {
			id, err = repo.SelectByLongURL(ctx, longURL.URL)
			statusCode = http.StatusConflict
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Ответ на запрос
		response := responseURL{Result: baseURL + "/" + id}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		setCookie(w, userID)
		w.WriteHeader(statusCode)
		_, err = w.Write(jsonResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handlerShortenURL(repo Repositorier, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		longURL := string(body)

		id := uuid.NewString()

		// запрос в БД на сохранение URL. Замена id на существующий в случае дублирования longURL
		var statusCode = http.StatusCreated
		ctx, cancel := context.WithTimeout(r.Context(), timeout*time.Second)
		defer cancel()
		err = repo.AddEntity(ctx, db.Entity{UserID: userID.String(), ID: id, URL: longURL})
		if errors.Is(err, db.UniqueViolationError) {
			id, err = repo.SelectByLongURL(ctx, longURL)
			statusCode = http.StatusConflict
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		shortURL := baseURL + "/" + id

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		setCookie(w, userID)
		w.WriteHeader(statusCode)
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handlerExpandURL(repo Repositorier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		ctx, cancel := context.WithTimeout(r.Context(), timeout*time.Second)
		defer cancel()
		longURL, err := repo.SelectByIDURL(ctx, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", longURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func handlerUserHistory(repo Repositorier, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout*time.Second)
		defer cancel()
		selection, err := repo.SelectByUser(ctx, userID.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		setCookie(w, userID)
		w.Header().Set("Content-Type", "application/json")

		if len(selection) > 0 {
			history := make(responseUserHistory, len(selection))
			for i, v := range selection {
				history[i] = item{
					ShortURL:    baseURL + "/" + v.ID,
					OriginalURL: v.URL,
				}
			}
			js, err := json.Marshal(history)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			w.WriteHeader(http.StatusOK)
			_, err = w.Write(js)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handlerPingDB(repo Repositorier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := repo.Ping(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

type batchOutput []batchOutputItem
type batchOutputItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func handlerShortenURLAPIBatch(repo Repositorier, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		input := db.BatchInput{}
		err = json.Unmarshal(body, &input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// generate ID's for short URL's
		for i := range input {
			input[i].ShortPath = uuid.NewString()
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout*time.Second)
		defer cancel()
		err = repo.Flush(ctx, userID.String(), input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		output := make(batchOutput, len(input))
		for i := range input {
			output[i].ShortURL = baseURL + "/" + input[i].ShortPath
			output[i].CorrelationID = input[i].CorrelationID
		}

		jsonResponse, err := json.Marshal(output)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		setCookie(w, userID)
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(jsonResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
