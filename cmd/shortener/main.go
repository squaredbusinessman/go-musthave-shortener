package main

import (
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type shortener struct {
	mu      sync.RWMutex
	storage map[string]string
}

func (s *shortener) PostingURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "text/plain") {
			http.Error(w, "Content-Type not supported", http.StatusUnsupportedMediaType)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		urlStr := string(body)

		if urlStr == "" {
			http.Error(w, "Empty URL", http.StatusBadRequest)
			return
		}

		mockID := "EwHXdJfB"

		s.mu.Lock()
		s.storage[mockID] = urlStr
		s.mu.Unlock()

		shortURL := "http://localhost:8080/" + mockID
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			log.Printf("Failed to write post-response: %v", err)
		}
	}
}

func (s *shortener) GetURLFromID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ID := r.URL.Query().Get("id")
		s.mu.RLock()
		url, ok := s.storage[ID]
		s.mu.RUnlock()
		if !ok {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusTemporaryRedirect)
		_, err := io.WriteString(w, url)
		if err != nil {
			log.Printf("Failed to write get-response: %v", err)
		}

	}
}

func main() {
	s := &shortener{
		storage: make(map[string]string),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.PostingURL())
	mux.HandleFunc("/{id}", s.GetURLFromID())

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
