package aiqicha

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientGetAndPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Write([]byte("get"))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if r.Method != http.MethodPost || string(body) != "request body" {
			t.Fatalf("request = %s %q, want POST %q", r.Method, body, "request body")
		}
		w.Write([]byte("post"))
	}))
	defer server.Close()

	client := New(Options{BaseURL: server.URL})
	for _, tt := range []struct {
		name string
		do   func() (*http.Response, error)
		want string
	}{
		{name: "GET", do: func() (*http.Response, error) { return client.Get(context.Background(), "/companies") }, want: "get"},
		{name: "POST", do: func() (*http.Response, error) {
			return client.Post(context.Background(), "/companies", strings.NewReader("request body"))
		}, want: "post"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response, err := tt.do()
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(body) != tt.want {
				t.Fatalf("response body = %q, want %q", body, tt.want)
			}
		})
	}
}
