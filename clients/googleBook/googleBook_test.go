package googleBook

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBooks(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"kind": "books#volumes", "totalItems": 1, "items": []}`))
	}))
	defer mockServer.Close()

	client := New(mockServer.URL)
	_, err := client.Books("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBookImage(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("image data"))
	}))
	defer mockServer.Close()

	client := New(mockServer.URL)
	imageLinks := ImageLinks{
		Medium: mockServer.URL,
	}

	image, err := client.BookImage(imageLinks)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(image) != "image data" {
		t.Fatalf("expected 'image data', got %s", string(image))
	}
}
