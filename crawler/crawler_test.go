package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
)

func Test_Analize_Page(t *testing.T) {
	withTestServer(t, func(server *httptest.Server) {
		httpClient, err := localClient(server.URL)
		if err != nil {
			t.Fatalf("localClient error: %v", err)
		}

		opts := Options{
			URL:        server.URL + "/about",
			Depth:      10,
			Retries:    0,
			Delay:      100 * time.Millisecond,
			Timeout:    5 * time.Second,
			HTTPClient: httpClient,
		}

		ctx := context.Background()
		_, err = Analize(ctx, opts)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})
}

func Test_Analize_BadRequest(t *testing.T) {
	httpClient := successClient(http.StatusBadRequest)

	opts := Options{
		URL:        "http://example.com",
		Depth:      1,
		Retries:    0,
		Delay:      100 * time.Millisecond,
		Timeout:    5 * time.Second,
		HTTPClient: httpClient,
	}

	ctx := context.Background()
	result, err := Analize(ctx, opts)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	ja := jsonassert.New(t)

	expected := `{
		"depth": 1,
		"generated_at": "<<PRESENCE>>",
		"root_url": "http://example.com",
		"pages": [
			{
				"depth": 0,
				"error": "",
				"http_status": 400,
				"status": "ok",
				"url": "http://example.com",
				"broken_links": [],
				"discovered_at": "<<PRESENCE>>"
			}
		]
	}`

	ja.Assert(string(result), expected)
}

func Test_Analize_ServerError(t *testing.T) {
	httpClient := successClient(http.StatusInternalServerError)

	opts := Options{
		URL:        "http://example.com",
		Depth:      1,
		Retries:    0,
		Delay:      100 * time.Millisecond,
		Timeout:    5 * time.Second,
		HTTPClient: httpClient,
	}

	ctx := context.Background()
	result, err := Analize(ctx, opts)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	ja := jsonassert.New(t)

	expected := `{
		"depth": 1,
		"generated_at": "<<PRESENCE>>",
		"root_url": "http://example.com",
		"pages": [
			{
				"depth": 0,
				"error": "",
				"http_status": 500,
				"status": "ok",
				"url": "http://example.com",
				"broken_links": [],
				"discovered_at": "<<PRESENCE>>"
			}
		]
	}`

	ja.Assert(string(result), expected)
}

func Test_Analize_Timeout(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}),
	)
	defer server.Close()

	opts := Options{
		URL:        server.URL,
		Depth:      1,
		Retries:    0,
		Delay:      100 * time.Millisecond,
		Timeout:    1 * time.Microsecond,
		HTTPClient: http.DefaultClient,
	}

	ctx := context.Background()
	result, err := Analize(ctx, opts)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	ja := jsonassert.New(t)

	expected := fmt.Sprintf(`{
		"depth": 1,
		"generated_at": "<<PRESENCE>>",
		"root_url": "%[1]s",
		"pages": [
			{
				"depth": 0,
				"error": "Get \"%[1]s\": context deadline exceeded (Client.Timeout exceeded while awaiting headers)",
				"http_status": 0,
				"status": "ok",
				"url": "%[1]s",
				"broken_links": [],
				"discovered_at": "<<PRESENCE>>"
			}
		]
	}`, server.URL)

	ja.Assert(string(result), expected)
}

func Test_Analize_NetworkError(t *testing.T) {
	httpClient := failingClient()

	opts := Options{
		URL:        "http://example.com",
		Depth:      1,
		Retries:    0,
		Delay:      100 * time.Millisecond,
		Timeout:    5 * time.Second,
		HTTPClient: httpClient,
	}

	ctx := context.Background()
	result, err := Analize(ctx, opts)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	ja := jsonassert.New(t)

	expected := `{
		"depth": 1,
		"generated_at": "<<PRESENCE>>",
		"root_url": "http://example.com",
		"pages": [
			{
				"depth": 0,
				"error": "Get \"http://example.com\": connection refused: network unreachable",
				"http_status": 0,
				"status": "ok",
				"url": "http://example.com",
				"broken_links": [],
				"discovered_at": "<<PRESENCE>>"
			}
		]
	}`

	ja.Assert(string(result), expected)
}

//////////////////////////////////////////////////////////////////////////////////

// // TestAnalize_MultipleRetries tests retry mechanism on failures
// func TestAnalize_MultipleRetries(t *testing.T) {
// 	attemptCount := 0

// 	// Create a test server that succeeds on 2nd attempt
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		attemptCount++
// 		if attemptCount == 1 {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("OK"))
// 	}))
// 	defer server.Close()

// 	opts := Options{
// 		URL:        server.URL,
// 		Depth:      1,
// 		Retries:    3,
// 		Delay:      10 * time.Millisecond,
// 		Timeout:    5 * time.Second,
// 		HTTPClient: http.DefaultClient,
// 	}

// 	ctx := context.Background()
// 	result, err := Analize(ctx, opts)

// 	if err != nil {
// 		t.Fatalf("Expected no error after retries, got: %v", err)
// 	}

// 	// Server should have been called twice (first failed, second succeeded)
// 	if attemptCount != 2 {
// 		t.Errorf("Expected 2 attempts, got %d", attemptCount)
// 	}

// 	if len(result) == 0 {
// 		t.Fatal("Expected non-empty result")
// 	}
// }

// // TestAnalize_CancelContext tests handling of context cancellation
// func TestAnalize_CancelContext(t *testing.T) {
// 	// Create a test server that delays response
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		time.Sleep(10 * time.Second)
// 		w.WriteHeader(http.StatusOK)
// 	}))
// 	defer server.Close()

// 	httpClient := &http.Client{
// 		Timeout: 5 * time.Second,
// 	}

// 	opts := Options{
// 		URL:        server.URL,
// 		Depth:      1,
// 		Retries:    0,
// 		Delay:      100 * time.Millisecond,
// 		Timeout:    5 * time.Second,
// 		HTTPClient: httpClient,
// 	}

// 	// Create a context that will be cancelled
// 	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
// 	defer cancel()

// 	result, err := Analize(ctx, opts)

// 	// Should complete without error, but context may be cancelled during execution
// 	_ = result
// 	_ = err
// }

// // TestAnalize_CustomHTTPClient tests with custom HTTP client configuration
// func TestAnalize_CustomHTTPClient(t *testing.T) {
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("Custom client test"))
// 	}))
// 	defer server.Close()

// 	customClient := &http.Client{
// 		Timeout: 10 * time.Second,
// 		Transport: &http.Transport{
// 			MaxIdleConns:        10,
// 			MaxIdleConnsPerHost: 5,
// 			IdleConnTimeout:     90 * time.Second,
// 		},
// 	}

// 	opts := Options{
// 		URL:        server.URL,
// 		Depth:      1,
// 		Retries:    0,
// 		Delay:      100 * time.Millisecond,
// 		Timeout:    10 * time.Second,
// 		HTTPClient: customClient,
// 	}

// 	ctx := context.Background()
// 	result, err := Analize(ctx, opts)

// 	if err != nil {
// 		t.Fatalf("Expected no error, got: %v", err)
// 	}

// 	if len(result) == 0 {
// 		t.Fatal("Expected non-empty result")
// 	}
// }

// // TestAnalize_UserAgent tests custom user agent handling
// func TestAnalize_UserAgent(t *testing.T) {
// 	var receivedUserAgent string

// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		receivedUserAgent = r.UserAgent()
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("User agent test"))
// 	}))
// 	defer server.Close()

// 	opts := Options{
// 		URL:        server.URL,
// 		Depth:      1,
// 		Retries:    0,
// 		Delay:      100 * time.Millisecond,
// 		Timeout:    5 * time.Second,
// 		UserAgent:  "MyCrawler/1.0",
// 		HTTPClient: http.DefaultClient,
// 	}

// 	ctx := context.Background()
// 	result, err := Analize(ctx, opts)

// 	if err != nil {
// 		t.Fatalf("Expected no error, got: %v", err)
// 	}

// 	if receivedUserAgent != "MyCrawler/1.0" {
// 		t.Errorf("Expected user agent 'MyCrawler/1.0', got '%s'", receivedUserAgent)
// 	}

// 	if len(result) == 0 {
// 		t.Fatal("Expected non-empty result")
// 	}
// }

// // TestAnalize_DifferentHTTPStatusCodes tests various HTTP status codes
// func TestAnalize_DifferentHTTPStatusCodes(t *testing.T) {
// 	testCases := []struct {
// 		name       string
// 		statusCode int
// 	}{
// 		{"OK", http.StatusOK},
// 		{"Created", http.StatusCreated},
// 		{"Bad Request", http.StatusBadRequest},
// 		{"Unauthorized", http.StatusUnauthorized},
// 		{"NotFound", http.StatusNotFound},
// 		{"Method Not Allowed", http.StatusMethodNotAllowed},
// 		{"Internal Server Error", http.StatusInternalServerError},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				w.WriteHeader(tc.statusCode)
// 				w.Write([]byte("Test"))
// 			}))
// 			defer server.Close()

// 			opts := Options{
// 				URL:        server.URL,
// 				Depth:      1,
// 				Retries:    0,
// 				Delay:      100 * time.Millisecond,
// 				Timeout:    5 * time.Second,
// 				HTTPClient: http.DefaultClient,
// 			}

// 			ctx := context.Background()
// 			result, err := Analize(ctx, opts)

// 			if err != nil {
// 				t.Fatalf("Expected no error, got: %v", err)
// 			}

// 			if len(result) == 0 {
// 				t.Fatal("Expected non-empty result")
// 			}
// 		})
// 	}
// }

// // TestAnalize_DifferentDepths tests different depth values
// func TestAnalize_DifferentDepths(t *testing.T) {
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("OK"))
// 	}))
// 	defer server.Close()

// 	testCases := []struct {
// 		depth uint
// 		pages int
// 	}{
// 		{0, 1},
// 		{1, 1},
// 		{2, 1},
// 		{5, 1},
// 	}

// 	for _, tc := range testCases {
// 		t.Run("Depth"+string(rune(tc.depth+'0')), func(t *testing.T) {
// 			opts := Options{
// 				URL:        server.URL,
// 				Depth:      tc.depth,
// 				Retries:    0,
// 				Delay:      100 * time.Millisecond,
// 				Timeout:    5 * time.Second,
// 				HTTPClient: http.DefaultClient,
// 			}

// 			ctx := context.Background()
// 			result, err := Analize(ctx, opts)

// 			if err != nil {
// 				t.Fatalf("Expected no error for depth %d, got: %v", tc.depth, err)
// 			}

// 			if len(result) == 0 {
// 				t.Fatal("Expected non-empty result")
// 			}
// 		})
// 	}
// }
