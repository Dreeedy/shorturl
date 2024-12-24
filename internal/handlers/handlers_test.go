package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages/filestorage"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenedURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := config.NewMockConfig(ctrl)
	mockStorage := filestorage.NewMockStorage(ctrl)

	handler := NewhandlerHTTP(mockConfig, mockStorage)

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
				response:    "http://localhost:8080/", // The exact hash will be checked later.
				contentType: "text/plain",
			},
		},
		{
			name: "valid URL 2",
			body: "https://www.google.com/",
			want: want{
				code:        201,
				response:    "http://localhost:8080/", // The exact hash will be checked later.
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

	// Initialize the router.
	r := chi.NewRouter()
	r.Post("/", handler.ShortenedURL)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStorage.EXPECT().SetURL(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			mockConfig.EXPECT().GetConfig().Return(config.HTTPConfig{BaseURL: "http://localhost:8080"}).AnyTimes()

			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(test.body))
			w := httptest.NewRecorder()

			// Use the router to serve the request.
			r.ServeHTTP(w, request)

			res := w.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("Error closing response body: %v", err)
				}
			}()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			log.Printf("TestShortenedURL.resBody: %s", string(resBody))

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			if test.want.code == 201 {
				assert.True(t, strings.HasPrefix(string(resBody), "http://localhost:8080/"))
				// Check the length of the hash.
				assert.Equal(t, 8, len(strings.TrimPrefix(string(resBody), "http://localhost:8080/")))
			} else {
				assert.Equal(t, test.want.response, string(resBody))
			}
		})
	}
}

func TestOriginalURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := config.NewMockConfig(ctrl)
	mockStorage := filestorage.NewMockStorage(ctrl)

	handler := NewhandlerHTTP(mockConfig, mockStorage)

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
			path: "/8a992351", // Example short hash.
			want: want{
				code:        307,
				location:    "https://practicum.yandex.ru",
				contentType: "",
			},
		},
		{
			name: "valid ID 2",
			path: "/d0e196a0", // Example short hash.
			want: want{
				code:        307,
				location:    "https://www.google.com/",
				contentType: "",
			},
		},
		{
			name: "Invalid ID",
			path: "/1234567",
			want: want{
				code:        400,
				location:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Invalid ID 2",
			path: "/12345678",
			want: want{
				code:        400,
				location:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	// Initialize the router.
	r := chi.NewRouter()
	r.Get("/{id}", handler.OriginalURL)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id := strings.TrimPrefix(test.path, "/")
			if test.want.code == 307 {
				mockStorage.EXPECT().GetURL(id).Return(test.want.location, true)
			} else {
				mockStorage.EXPECT().GetURL(id).Return("", false)
			}

			request := httptest.NewRequest(http.MethodGet, test.path, http.NoBody)
			w := httptest.NewRecorder()

			// Use the router to serve the request.
			r.ServeHTTP(w, request)

			res := w.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("Error closing response body: %v", err)
				}
			}()

			assert.Equal(t, test.want.code, res.StatusCode)
			if test.want.code == 307 {
				assert.Equal(t, test.want.location, res.Header.Get("Location"))
			} else {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestShorten(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := config.NewMockConfig(ctrl)
	mockStorage := filestorage.NewMockStorage(ctrl)

	handler := NewhandlerHTTP(mockConfig, mockStorage)

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
			body: `{"url": "https://practicum.yandex.ru"}`,
			want: want{
				code:        201,
				response:    `{"result":"http://localhost:8080/"}`, // The exact hash will be checked later.
				contentType: "application/json",
			},
		},
		{
			name: "valid URL 2",
			body: `{"url": "https://www.google.com/"}`,
			want: want{
				code:        201,
				response:    `{"result":"http://localhost:8080/"}`, // The exact hash will be checked later.
				contentType: "application/json",
			},
		},
		{
			name: "empty URL",
			body: `{"url": ""}`,
			want: want{
				code:        400,
				response:    "URL is empty\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "invalid JSON",
			body: `{"url": "https://practicum.yandex.ru"`, // Missing closing brace.
			want: want{
				code:        400,
				response:    "Unable to read request body\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "invalid body format",
			body: `{"url": "invalid-body}`,
			want: want{
				code:        400,
				response:    "Unable to read request body\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	// Initialize the router.
	r := chi.NewRouter()
	r.Post("/shorten", handler.Shorten)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStorage.EXPECT().SetURL(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			mockConfig.EXPECT().GetConfig().Return(config.HTTPConfig{BaseURL: "http://localhost:8080"}).AnyTimes()

			request := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(test.body))
			w := httptest.NewRecorder()

			// Use the router to serve the request.
			r.ServeHTTP(w, request)

			res := w.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("Error closing response body: %v", err)
				}
			}()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			log.Printf("TestShorten.resBody: %s", string(resBody))

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			if test.want.code == 201 {
				var shortenAPIRs ShortenAPIRs
				err := json.Unmarshal(resBody, &shortenAPIRs)
				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(shortenAPIRs.Result, "http://localhost:8080/"))
				// Check the length of the hash.
				assert.Equal(t, 8, len(strings.TrimPrefix(shortenAPIRs.Result, "http://localhost:8080/")))
			} else {
				assert.Equal(t, test.want.response, string(resBody))
			}
		})
	}
}
