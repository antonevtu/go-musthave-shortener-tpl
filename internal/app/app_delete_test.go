package app

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/cfg"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/db"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/handlers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type shortIDList []string

func TestDBDeleteBatch(t *testing.T) {
	cfgApp := cfg.Config{
		ServerAddress:   *ServerAddress,
		BaseURL:         *BaseURL,
		FileStoragePath: *FileStoragePath,
		DatabaseDSN:     *DatabaseDSN,
		CtxTimeout:      *CtxTimeout,
	}

	dbPool, err := db.New(context.Background(), *DatabaseDSN)
	assert.Equal(t, err, nil)
	r := handlers.NewRouter(&dbPool, cfgApp)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// запись в БД
	batch := make(batchInput, 3)
	batch[0] = batchInputItem{CorrelationID: "0", OriginalURL: "https://yandex.ru/" + uuid.NewString()}
	batch[1] = batchInputItem{CorrelationID: "1", OriginalURL: "https://yandex.ru/" + uuid.NewString()}
	batch[2] = batchInputItem{CorrelationID: "2", OriginalURL: "https://yandex.ru/" + uuid.NewString()}
	jsonBatch, err := json.Marshal(batch)
	require.NoError(t, err)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten/batch", bytes.NewBuffer(jsonBatch))
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusCreated)
	cookies := resp.Cookies()
	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	// извлечение shortIDs
	var batchOut batchOutput
	err = json.Unmarshal(respBody, &batchOut)
	require.NoError(t, err)
	shortIDs := make(shortIDList, 0, len(batch))
	for _, shortURL := range batchOut {
		u, err := url.Parse(shortURL.ShortURL)
		require.NoError(t, err)
		shortIDs = append(shortIDs, u.Path[1:])
	}

	// Запрос на удаление
	resp1, _ := testGZipRequestCookie(t, ts.URL+"/api/user/urls", http.MethodDelete, testEncodeJSONDeleteList(shortIDs), cookies)
	err = resp1.Body.Close()
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp1.StatusCode)
}

func testEncodeJSONDeleteList(s shortIDList) *bytes.Buffer {
	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false) // без этой опции символ '&' будет заменён на "\u0026"
	_ = encoder.Encode(s)
	return buf
}
