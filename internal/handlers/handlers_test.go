package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenedURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "valid URL",
			body: "https://practicum.yandex.ru",
			want: want{
				code:        201,
				response:    "http://localhost:8080/8a9923515b446c11cef0fb86da0b29e3206fa3674412ae2de61299b820859aa2",
				contentType: "text/plain",
			},
		},
		{
			name: "valid URL 2",
			body: "https://www.google.com/",
			want: want{
				code:        201,
				response:    "http://localhost:8080/d0e196a0c25d35dd0a84593cbae0f38333aa58529936444ea26453eab28dfc86",
				contentType: "text/plain",
			},
		},
		{
			name: "empty URL",
			body: "",
			want: want{
				code:        400,
				response:    "URL is empty\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	// Initialize the router
	r := chi.NewRouter()
	r.Post("/", ShortenedURL)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(test.body))
			w := httptest.NewRecorder()

			// Use the router to serve the request
			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			if test.want.code == 201 {
				assert.True(t, strings.HasPrefix(string(resBody), "http://localhost:8080/"))
			} else {
				assert.Equal(t, test.want.response, string(resBody))
			}
		})
	}
}

func TestOriginalURL(t *testing.T) {
	type want struct {
		code        int
		location    string
		contentType string
	}
	tests := []struct {
		name string
		path string
		want want
	}{
		{
			name: "valid ID",
			path: "/8a9923515b446c11cef0fb86da0b29e3206fa3674412ae2de61299b820859aa2",
			want: want{
				code:        307,
				location:    "https://practicum.yandex.ru",
				contentType: "",
			},
		},
		{
			name: "valid ID 2",
			path: "/d0e196a0c25d35dd0a84593cbae0f38333aa58529936444ea26453eab28dfc86",
			want: want{
				code:        307,
				location:    "https://www.google.com/",
				contentType: "",
			},
		},
		{
			name: "Invalid ID",
			path: "/1234567890",
			want: want{
				code:        400,
				location:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	// Initialize the router
	r := chi.NewRouter()
	r.Get("/{id}", OriginalURL)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			w := httptest.NewRecorder()

			// Use the router to serve the request
			r.ServeHTTP(w, request)

			res := w.Result()

			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			if test.want.code == 307 {
				assert.Equal(t, test.want.location, res.Header.Get("Location"))
			} else {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
