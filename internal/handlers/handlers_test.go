package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storage"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenedURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfig := config.NewMockConfig(ctrl)
	mockStorage := storage.NewMockStorage(ctrl)

	handler := NewMyHandler(mockConfig, mockStorage)

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
			mockStorage.EXPECT().Exists(gomock.Any()).Return(false).AnyTimes()
			mockStorage.EXPECT().SetURL(gomock.Any(), gomock.Any()).AnyTimes()
			mockConfig.EXPECT().GetConfig().Return(config.MyConfig{BaseURL: "http://localhost:8080"}).AnyTimes()

			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(test.body))
			w := httptest.NewRecorder()

			// Use the router to serve the request.
			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			log.Printf("TestShortenedURL.resBody: %s", string(resBody))

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			if test.want.code == 201 {
				assert.True(t, strings.HasPrefix(string(resBody), "http://localhost:8080/"))
				assert.Equal(t, 8, len(strings.TrimPrefix(string(resBody), "http://localhost:8080/"))) // Check the length of the hash.
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
	mockStorage := storage.NewMockStorage(ctrl)

	handler := NewMyHandler(mockConfig, mockStorage)

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

			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			w := httptest.NewRecorder()

			// Use the router to serve the request.
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
