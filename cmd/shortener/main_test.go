package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestPostingURL(t *testing.T) {
	const (
		mockID      = "EwHXdJfB"
		shortURL    = "http://localhost:8080/" + mockID
		originalURL = "http://k0jywbhmyihtgr.ru/puqzlytten/kyluppw2"
	)

	tests := []struct {
		name        string
		method      string
		body        string
		contentType string
		wantStatus  int
		wantBody    string
		wantStored  bool
	}{
		{
			name:        "success",
			method:      http.MethodPost,
			body:        originalURL,
			contentType: "text/plain",
			wantStatus:  http.StatusCreated,
			wantBody:    shortURL,
			wantStored:  true,
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "Method not allowed\n",
		},
		{
			name:        "unsupported content type",
			method:      http.MethodPost,
			body:        originalURL,
			contentType: "application/json",
			wantStatus:  http.StatusUnsupportedMediaType,
			wantBody:    "Content-Type not supported\n",
		},
		{
			name:        "empty body",
			method:      http.MethodPost,
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
			wantBody:    "Empty URL\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &shortener{storage: make(map[string]string)}
			handler := s.PostingURL()

			var bodyReader *strings.Reader
			if tt.body == "" {
				bodyReader = strings.NewReader("")
			} else {
				bodyReader = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/", bodyReader)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			resp := httptest.NewRecorder()
			handler(resp, req)

			if resp.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.Code, tt.wantStatus)
			}

			if gotBody := resp.Body.String(); gotBody != tt.wantBody {
				t.Fatalf("body = %q, want %q", gotBody, tt.wantBody)
			}

			value, ok := s.storage[mockID]
			if tt.wantStored {
				if !ok {
					t.Fatal("expected value stored but map empty")
				}
				if value != tt.body {
					t.Fatalf("stored value = %q, want %q", value, tt.body)
				}
			} else if ok {
				t.Fatalf("unexpected stored value: %q", value)
			}
		})
	}
}

func TestGetURLFromID(t *testing.T) {
	const (
		mockID    = "EwHXdJfB"
		sampleURL = "http://example.com/alpha"
	)

	withRouteParam := func(r *http.Request, key, value string) *http.Request {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add(key, value)
		ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
		return r.WithContext(ctx)
	}

	tests := []struct {
		name         string
		method       string
		path         string
		storage      map[string]string
		paramValue   *string
		wantStatus   int
		wantBody     string
		wantLocation string
	}{
		{
			name:         "redirects existing",
			method:       http.MethodGet,
			path:         "/" + mockID,
			storage:      map[string]string{mockID: sampleURL},
			paramValue:   strPtr(mockID),
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: sampleURL,
		},
		{
			name:       "unknown id",
			method:     http.MethodGet,
			path:       "/unknown",
			storage:    map[string]string{mockID: sampleURL},
			paramValue: strPtr("unknown"),
			wantStatus: http.StatusNotFound,
			wantBody:   "URL not found\n",
		},
		{
			name:       "empty id",
			method:     http.MethodGet,
			path:       "/",
			storage:    map[string]string{mockID: sampleURL},
			wantStatus: http.StatusNotFound,
			wantBody:   "URL not found\n",
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			path:       "/" + mockID,
			storage:    map[string]string{mockID: sampleURL},
			paramValue: strPtr(mockID),
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "Method not allowed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &shortener{storage: make(map[string]string)}
			for k, v := range tt.storage {
				s.storage[k] = v
			}

			handler := s.GetURLFromID()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.paramValue != nil {
				req = withRouteParam(req, "id", *tt.paramValue)
			}
			resp := httptest.NewRecorder()

			handler(resp, req)

			if resp.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.Code, tt.wantStatus)
			}

			if tt.wantBody != "" {
				if got := resp.Body.String(); got != tt.wantBody {
					t.Fatalf("body = %q, want %q", got, tt.wantBody)
				}
			}

			if tt.wantLocation != "" {
				if got := resp.Header().Get("Location"); got != tt.wantLocation {
					t.Fatalf("location header = %q, want %q", got, tt.wantLocation)
				}
			}
		})
	}
}

func strPtr(v string) *string {
	return &v
}
