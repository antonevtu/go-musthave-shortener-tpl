package handlers

import (
	"context"
	"encoding/json"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/cfg"
	"io"
	"net/http"
)

type shortIDList []string

func handlerDelete(repo Repositorier, cfgApp cfg.Config) http.HandlerFunc {
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

		var shortIDs shortIDList
		err = json.Unmarshal(body, &shortIDs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_ = userID

		sequentalDelete(repo, userID.String(), shortIDs)

		w.WriteHeader(http.StatusAccepted)
		return
	}
}

func sequentalDelete(repo Repositorier, userID string, shortIDs shortIDList) {
	repo.SetDeletedBatch(context.Background(), userID, shortIDs)
}
