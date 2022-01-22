package handlers

import (
	"encoding/json"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/cfg"
	"io"
	"net/http"
)

type shortIDList []string

func handlerDelete(repo Repositorier, cfgApp cfg.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDBytes, err := getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		userID := userIDBytes.String()

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

		//sequentialDelete(repo, userID.String(), shortIDs)

		for _, shortID := range shortIDs {
			cfgApp.ToDeleteChan <- cfg.ToDeleteItem{UserID: userID, ShortID: shortID}
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

//func sequentialDelete(repo Repositorier, userID string, shortIDs shortIDList) {
//	repo.SetDeletedBatch(context.Background(), userID, shortIDs)
//}
