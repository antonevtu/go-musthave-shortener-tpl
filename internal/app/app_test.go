package app

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerMethodNotAllowed(t *testing.T) {
	tests := []struct {
		name           string
		request        string
		url            string
		base           map[string]string
		requestMethod  string
		wantStatusCode int
	}{
		{
			name:           "Check existing ID",
			request:        "http:localhost:8080/",
			url:            "https://yandex.ru/maps/geo/sochi/53166566/?ll=39.580041%2C43.713351&z=9.98",
			base:           make(map[string]string),
			requestMethod:  http.MethodPut,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// добавление ID методом POST
			request := httptest.NewRequest(tt.requestMethod, tt.request, bytes.NewBufferString(tt.url))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler(tt.base))
			h.ServeHTTP(w, request)
			result := w.Result()

			require.Equal(t, tt.wantStatusCode, result.StatusCode)
		})
	}
}

func TestHandlerExistingId(t *testing.T) {
	type want struct {
		contentType    string
		postStatusCode int
		location       string
		getStatusCode  int
	}
	tests := []struct {
		name     string
		request  string
		url      string
		base     map[string]string
		shortURL string
		want     want
	}{
		{
			name:     "Check existing ID",
			request:  "http:localhost:8080/",
			url:      "https://yandex.ru/maps/geo/sochi/53166566/?ll=39.580041%2C43.713351&z=9.98",
			base:     make(map[string]string),
			shortURL: "correct",
			want: want{
				contentType:    "text/plain; charset=utf-8",
				postStatusCode: http.StatusCreated,
				location:       "https://yandex.ru/maps/geo/sochi/53166566/?ll=39.580041%2C43.713351&z=9.98",
				getStatusCode:  http.StatusTemporaryRedirect,
			},
		},

		{
			name:     "Check existing ID for empty path",
			request:  "http:localhost:8080/",
			url:      "",
			base:     make(map[string]string),
			shortURL: "correct",
			want: want{
				contentType:    "text/plain; charset=utf-8",
				postStatusCode: http.StatusCreated,
				location:       "",
				getStatusCode:  http.StatusTemporaryRedirect,
			},
		},

		{
			name:     "Check existing ID for random path",
			request:  "http:localhost:8080/",
			url:      "kmdlgjmxa",
			base:     make(map[string]string),
			shortURL: "correct",
			want: want{
				contentType:    "text/plain; charset=utf-8",
				postStatusCode: http.StatusCreated,
				location:       "kmdlgjmxa",
				getStatusCode:  http.StatusTemporaryRedirect,
			},
		},

		{
			name:     "Check non-existent ID",
			request:  "http:localhost:8080/",
			url:      "https://yandex.ru/maps/geo/sochi/53166566/?ll=39.580041%2C43.713351&z=9.98",
			base:     make(map[string]string),
			shortURL: "fslm3",
			want: want{
				contentType:    "text/plain; charset=utf-8",
				postStatusCode: http.StatusCreated,
				location:       "",
				getStatusCode:  http.StatusBadRequest,
			},
		},

		{
			name:     "Check empty ID",
			request:  "http:localhost:8080/",
			url:      "https://yandex.ru/maps/geo/sochi/53166566/?ll=39.580041%2C43.713351&z=9.98",
			base:     make(map[string]string),
			shortURL: "",
			want: want{
				contentType:    "text/plain; charset=utf-8",
				postStatusCode: http.StatusCreated,
				location:       "",
				getStatusCode:  http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// добавление ID методом POST
			request := httptest.NewRequest(http.MethodPost, tt.request, bytes.NewBufferString(tt.url))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(handler(tt.base))
			h.ServeHTTP(w, request)
			result := w.Result()

			require.Equal(t, tt.want.postStatusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			shortURL, err := ioutil.ReadAll(result.Body)
			assert.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			var shortURLForGET string
			switch tt.shortURL {
			case "correct":
				shortURLForGET = string(shortURL)
			default:
				shortURLForGET = tt.request + tt.shortURL
			}

			// восстановление URL методом Get
			request = httptest.NewRequest(http.MethodGet, shortURLForGET, nil)
			w = httptest.NewRecorder()
			h = http.HandlerFunc(handler(tt.base))
			h.ServeHTTP(w, request)
			result = w.Result()

			assert.Equal(t, tt.want.getStatusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
