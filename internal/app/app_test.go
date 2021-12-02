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
	client := &http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)

	resp, err := client.Do(req)
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

	// Create ID
	longURL := "https://yandex.ru/maps/geo/sochi/53166566/?ll=39.580041%2C43.713351&z=9.98"
	resp, shortURL := testRequest(t, ts.URL, "POST", bytes.NewBufferString(longURL))
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Check redirection by existing ID
	resp, _ = testRequest(t, shortURL, "GET", nil)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	longURLRecovered := resp.Header.Get("Location")
	assert.Equal(t, longURL, longURLRecovered)

	// Check StatusBadRequest for non existed ID
	shortURL1 := shortURL + "xxx"
	resp, _ = testRequest(t, shortURL1, "GET", nil)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Check not allowed method error
	resp, _ = testRequest(t, shortURL, "PUT", nil)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	_ = resp.Body.Close()
}
