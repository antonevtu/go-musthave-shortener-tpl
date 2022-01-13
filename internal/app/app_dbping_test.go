package app

import (
	"context"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/db"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDBPing(t *testing.T) {
	var cfgApp = config(t)
	err := db.Pool.New(context.Background(), cfgApp.DatabaseDSN)
	assert.Equal(t, err, nil)
	r := handlers.NewRouter(&db.Pool, cfgApp)
	ts := httptest.NewServer(r)
	defer ts.Close()

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/ping", nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	defer resp.Body.Close()

}
