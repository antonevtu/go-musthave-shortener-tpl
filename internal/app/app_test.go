package app

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, url, method string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	base := make(Storage, 100)
	r := NewRouter(base)
	ts := httptest.NewServer(r)
	defer ts.Close()

	longURL := "https://yandex.ru/maps/geo/sochi/53166566/?ll=39.580041%2C43.713351&z=9.98"
	resp, shortURL := testRequest(t, ts.URL, "POST", bytes.NewBufferString(longURL))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	shortURL1 := shortURL + "xxx"
	resp, _ = testRequest(t, shortURL1, "GET", nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, _ = testRequest(t, shortURL, "PUT", nil)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	resp.Body.Close()
}
