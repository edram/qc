package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestMiddlewareRunsBeforeAndAfterNext(t *testing.T) {
	var calls []string
	transport := roundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
		calls = append(calls, "transport")
		if got := request.URL.String(); got != "https://example.com/resources" {
			t.Fatalf("URL = %q, want %q", got, "https://example.com/resources")
		}
		if got := request.Header.Get("X-Request"); got != "modified" {
			t.Fatalf("X-Request = %q, want %q", got, "modified")
		}
		return &stdhttp.Response{Header: make(stdhttp.Header), Request: request}, nil
	})
	client := New(
		WithBaseURL("https://example.com"),
		WithHTTPClient(&stdhttp.Client{Transport: transport}),
	)
	client.AddMiddleware(func(ctx *Context, next Next) error {
		calls = append(calls, "first:before")
		ctx.Request.Header.Set("X-Request", "modified")
		if err := next(); err != nil {
			return err
		}
		calls = append(calls, "first:after")
		ctx.Response.Header.Set("X-Response", "modified")
		return nil
	})
	client.AddMiddleware(func(ctx *Context, next Next) error {
		calls = append(calls, "second:before")
		if err := next(); err != nil {
			return err
		}
		calls = append(calls, "second:after")
		return nil
	})
	response, err := client.Get("/resources")
	if err != nil {
		t.Fatal(err)
	}
	if got := response.Header.Get("X-Response"); got != "modified" {
		t.Fatalf("X-Response = %q, want %q", got, "modified")
	}
	want := []string{"first:before", "second:before", "transport", "second:after", "first:after"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestBaseURLUsesStandardURLResolution(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		endpoint string
		want     string
	}{
		{name: "page-relative with file base", baseURL: "https://example.com/base", endpoint: "users", want: "https://example.com/users"},
		{name: "page-relative with directory base", baseURL: "https://example.com/base/", endpoint: "users", want: "https://example.com/base/users"},
		{name: "origin-relative", baseURL: "https://example.com/base/", endpoint: "/users", want: "https://example.com/users"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(
				WithBaseURL(tt.baseURL),
				WithHTTPClient(&stdhttp.Client{Transport: roundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
					if got := request.URL.String(); got != tt.want {
						t.Fatalf("URL = %q, want %q", got, tt.want)
					}
					return &stdhttp.Response{Header: make(stdhttp.Header), Request: request}, nil
				})}),
			)
			response, err := client.Get(tt.endpoint)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
		})
	}
}

func TestWithContextScopesCancellation(t *testing.T) {
	client := New(WithHTTPClient(&stdhttp.Client{Transport: roundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
		if err := request.Context().Err(); err != nil {
			return nil, err
		}
		return &stdhttp.Response{Header: make(stdhttp.Header), Body: stdhttp.NoBody, Request: request}, nil
	})}))
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	response, err := client.WithContext(cancelled).Get("https://example.com")
	if response != nil {
		response.Body.Close()
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WithContext().Get() error = %v, want %v", err, context.Canceled)
	}

	response, err = client.Get("https://example.com")
	if err != nil {
		t.Fatalf("Get() after scoped cancellation: %v", err)
	}
	response.Body.Close()
}

func TestPrefixAppliesBeforeBaseURL(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{name: "page-relative prefix", prefix: "api", want: "https://example.com/base/api/users"},
		{name: "origin-relative prefix", prefix: "/api", want: "https://example.com/api/users"},
		{name: "absolute prefix", prefix: "https://api.example.com/v1", want: "https://api.example.com/v1/users"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(
				WithBaseURL("https://example.com/base/"),
				WithPrefix(tt.prefix),
				WithHTTPClient(&stdhttp.Client{Transport: roundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
					if got := request.URL.String(); got != tt.want {
						t.Fatalf("URL = %q, want %q", got, tt.want)
					}
					return &stdhttp.Response{Header: make(stdhttp.Header), Request: request}, nil
				})}),
			)
			response, err := client.Get("/users")
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
		})
	}
}

func TestMiddlewareCanStopChain(t *testing.T) {
	wantErr := errors.New("middleware failed")
	transportCalled := false
	client := New(WithHTTPClient(&stdhttp.Client{Transport: roundTripFunc(func(*stdhttp.Request) (*stdhttp.Response, error) {
		transportCalled = true
		return nil, nil
	})}))
	client.AddMiddleware(func(*Context, Next) error { return wantErr })
	request, err := stdhttp.NewRequest(stdhttp.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Do(request)

	if !errors.Is(err, wantErr) {
		t.Fatalf("Do() error = %v, want %v", err, wantErr)
	}
	if transportCalled {
		t.Fatal("transport was called after middleware stopped the chain")
	}
}

func TestMiddlewareNextCanOnlyBeCalledOnce(t *testing.T) {
	transportCalls := 0
	client := New(WithHTTPClient(&stdhttp.Client{Transport: roundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
		transportCalls++
		return &stdhttp.Response{Header: make(stdhttp.Header), Request: request}, nil
	})}))
	client.AddMiddleware(func(_ *Context, next Next) error {
		if err := next(); err != nil {
			return err
		}
		return next()
	})
	request, err := stdhttp.NewRequest(stdhttp.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Do(request)

	if !errors.Is(err, ErrNextCalledMoreThanOnce) {
		t.Fatalf("Do() error = %v, want %v", err, ErrNextCalledMoreThanOnce)
	}
	if transportCalls != 1 {
		t.Fatalf("transport calls = %d, want 1", transportCalls)
	}
}

func TestAddMiddlewarePanicsAfterRequest(t *testing.T) {
	client := New(WithHTTPClient(&stdhttp.Client{Transport: roundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
		return &stdhttp.Response{Header: make(stdhttp.Header), Request: request}, nil
	})}))
	request, err := stdhttp.NewRequest(stdhttp.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Do(request); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("AddMiddleware() did not panic after a request")
		}
	}()
	client.AddMiddleware(func(*Context, Next) error { return nil })
}

func TestConcurrentFirstRequests(t *testing.T) {
	client := New(WithHTTPClient(&stdhttp.Client{Transport: roundTripFunc(func(request *stdhttp.Request) (*stdhttp.Response, error) {
		return &stdhttp.Response{Header: make(stdhttp.Header), Request: request}, nil
	})}))
	client.AddMiddleware(func(_ *Context, next Next) error { return next() })

	const requestCount = 32
	start := make(chan struct{})
	results := make(chan error, requestCount)
	for range requestCount {
		go func() {
			<-start
			response, err := client.Get("https://example.com")
			if response != nil {
				response.Body.Close()
			}
			results <- err
		}()
	}
	close(start)

	for range requestCount {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
}

func TestAddMiddlewarePanicsForNil(t *testing.T) {
	client := New()

	defer func() {
		if recover() == nil {
			t.Fatal("AddMiddleware(nil) did not panic")
		}
	}()
	client.AddMiddleware(nil)
}

func TestNewUsesHTTPDefaultsWithoutOptions(t *testing.T) {
	client := New()
	if got := client.httpClient.Timeout; got != 0 {
		t.Fatalf("Timeout = %s, want 0", got)
	}
	if client.httpClient.CheckRedirect != nil {
		t.Fatal("CheckRedirect was customized")
	}
}

func TestWithoutRedirects(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(response stdhttp.ResponseWriter, request *stdhttp.Request) {
		if request.URL.Path == "/redirected" {
			response.WriteHeader(stdhttp.StatusOK)
			return
		}
		stdhttp.Redirect(response, request, "/redirected", stdhttp.StatusFound)
	}))
	defer server.Close()

	client := New(WithoutRedirects())
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusFound {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusFound)
	}
}

func TestOptionsApplyToProvidedHTTPClientRegardlessOfOrder(t *testing.T) {
	httpClient := &stdhttp.Client{Timeout: time.Minute}
	client := New(
		WithTimeout(5*time.Second),
		WithoutRedirects(),
		WithHTTPClient(httpClient),
	)

	if got := client.httpClient.Timeout; got != 5*time.Second {
		t.Fatalf("Timeout = %s, want %s", got, 5*time.Second)
	}
	request, err := stdhttp.NewRequest(stdhttp.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.httpClient.CheckRedirect(request, nil); !errors.Is(err, stdhttp.ErrUseLastResponse) {
		t.Fatalf("CheckRedirect() error = %v, want %v", err, stdhttp.ErrUseLastResponse)
	}
	if got := httpClient.Timeout; got != time.Minute {
		t.Fatalf("provided client Timeout = %s, want %s", got, time.Minute)
	}
}

func TestWithHTTPClientPreservesConfiguration(t *testing.T) {
	httpClient := &stdhttp.Client{}
	client := New(WithHTTPClient(httpClient))

	if client.httpClient == httpClient {
		t.Fatal("WithHTTPClient did not copy the provided client")
	}
	if client.httpClient.Timeout != 0 {
		t.Fatalf("Timeout = %s, want 0", client.httpClient.Timeout)
	}
	if client.httpClient.CheckRedirect != nil {
		t.Fatal("CheckRedirect was overridden")
	}
}

type roundTripFunc func(*stdhttp.Request) (*stdhttp.Response, error)

func (f roundTripFunc) RoundTrip(request *stdhttp.Request) (*stdhttp.Response, error) {
	return f(request)
}
