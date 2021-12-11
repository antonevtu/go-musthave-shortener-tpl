package app

import (
	"bytes"
	"encoding/json"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/cfg"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/handlers"
	"github.com/antonevtu/go-musthave-shortener-tpl/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestJSONAPI(t *testing.T) {
	repo := repository.New()
	r := handlers.NewRouter(repo, cfg.Get())
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Create ID
	longURL := "https://yandex.ru/maps/geo/sochi/53166566/?ll=39.580041%2C43.713351&z=9.98"
	buf := testEncodeJSONLongURL(longURL)
	resp, shortURLInJSON := testRequest(t, ts.URL+"/api/shorten", "POST", buf)
	_ = shortURLInJSON
	err := resp.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Parse shortURL
	shortURL := testDecodeJSONShortURL(t, shortURLInJSON)
	u, err := url.Parse(shortURL)
	require.NoError(t, err)

	// Check redirection by existing ID
	resp, _ = testRequest(t, ts.URL+u.Path, "GET", nil)
	err = resp.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	longURLRecovered := resp.Header.Get("Location")
	assert.Equal(t, longURL, longURLRecovered)

	// Check StatusBadRequest for incorrect JSON key in request
	badJSON := `{"urlBad":"abc"}`
	resp, _ = testRequest(t, ts.URL+"/api/shorten", "POST", bytes.NewBufferString(badJSON))
	err = resp.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Check server returns correct baseURL
	baseURL := u.Scheme + "://" + u.Host
	assert.Equal(t, baseURL, cfg.Get().BaseURL)
}

func TestTextAPI(t *testing.T) {
	repo := repository.New()
	r := handlers.NewRouter(repo, cfg.Get())
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Create ID
	longURL := "https://yandex.ru/maps/geo/sochi/53166566/?ll=39.580041%2C43.713351&z=9.98"
	resp, shortURL_ := testRequest(t, ts.URL, "POST", bytes.NewBufferString(longURL))
	err := resp.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Parse shortURL
	u, err := url.Parse(shortURL_)
	require.NoError(t, err)

	// Check redirection by existing ID
	resp, _ = testRequest(t, ts.URL+u.Path, "GET", nil)
	err = resp.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	longURLRecovered := resp.Header.Get("Location")
	assert.Equal(t, longURL, longURLRecovered)

	// Check StatusBadRequest for non existed ID
	shortURL1 := ts.URL + u.Path + "xxx"
	resp, _ = testRequest(t, shortURL1, "GET", nil)
	err = resp.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Check not allowed method error
	resp, _ = testRequest(t, ts.URL+u.Path, "PUT", nil)
	err = resp.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

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

func testEncodeJSONLongURL(url string) *bytes.Buffer {
	v := struct {
		URL string `json:"url"`
	}{URL: url}
	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false) // без этой опции символ '&' будет заменён на "\u0026"
	_ = encoder.Encode(v)
	return buf
}

func testDecodeJSONShortURL(t *testing.T, js string) string {
	url := struct {
		Result string `json:"result"`
	}{}
	err := json.Unmarshal([]byte(js), &url)
	require.NoError(t, err)
	return url.Result
}
