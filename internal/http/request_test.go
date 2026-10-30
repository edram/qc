package http

import (
	"context"
	stdhttp "net/http"
	"reflect"
	"testing"
)

func TestRequestHeadersAreScopedAndAvailableToMiddleware(t *testing.T) {
	client := New(
		WithBaseURL("https://example.com"),
		WithHTTPClient(&stdhttp.Client{Transport: roundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
			return &stdhttp.Response{StatusCode: stdhttp.StatusOK, Request: request}, nil
		})}),
	)
	var observed []string
	client.AddMiddleware(func(ctx *Context, next Next) error {
		observed = append(observed, ctx.Request.Referer())
		return next()
	})
	scoped := client.WithContext(context.Background())
	for _, tt := range []struct {
		name string
		do   func() (*stdhttp.Response, error)
		want string
	}{
		{"client header", func() (*stdhttp.Response, error) {
			return client.Request(stdhttp.MethodGet, "/items", nil, WithRequestHeader("Referer", "https://example.com/first"))
		}, "https://example.com/first"},
		{"scoped header", func() (*stdhttp.Response, error) {
			return scoped.Request(stdhttp.MethodGet, "/items", nil, WithRequestHeader("Referer", "https://example.com/second"))
		}, "https://example.com/second"},
		{"scoped without header", func() (*stdhttp.Response, error) { return scoped.Get("/items") }, ""},
		{"client without header", func() (*stdhttp.Response, error) { return client.Get("/items") }, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response, err := tt.do()
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if got := response.Request.Referer(); got != tt.want {
				t.Fatalf("Referer = %q, want %q", got, tt.want)
			}
		})
	}
	if want := []string{"https://example.com/first", "https://example.com/second", "", ""}; !reflect.DeepEqual(observed, want) {
		t.Fatalf("middleware Referers = %q, want %q", observed, want)
	}
}
