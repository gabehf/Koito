package queue

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRequestQueue_Enqueue_SingleRequest(t *testing.T) {
	// Arrange: test server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	q := NewRequestQueue(1000, 1000) // high limits so rate limiter won't slow test
	defer q.Shutdown()

	// Act: enqueue one job
	resCh := q.Enqueue(func(client *http.Client, done chan<- RequestResult) {
		resp, err := client.Get(srv.URL)
		if err != nil {
			done <- RequestResult{Err: err}
			return
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		done <- RequestResult{Body: b, Err: err}
	})

	// Assert: receive result with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	select {
	case res := <-resCh:
		if res.Err != nil {
			t.Fatalf("expected nil error, got %v", res.Err)
		}
		if string(res.Body) != "ok" {
			t.Fatalf("expected body %q, got %q", "ok", string(res.Body))
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for queued request result")
	}
}

func TestRequestQueue_Enqueue_MultipleRequests(t *testing.T) {
	var hits int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	q := NewRequestQueue(1000, 1000)
	defer q.Shutdown()

	const n = 10
	resultChans := make([]<-chan RequestResult, 0, n)

	for i := 0; i < n; i++ {
		ch := q.Enqueue(func(client *http.Client, done chan<- RequestResult) {
			resp, err := client.Get(srv.URL)
			if err != nil {
				done <- RequestResult{Err: err}
				return
			}
			defer resp.Body.Close()
			b, err := io.ReadAll(resp.Body)
			done <- RequestResult{Body: b, Err: err}
		})
		resultChans = append(resultChans, ch)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for i, ch := range resultChans {
		select {
		case res := <-ch:
			if res.Err != nil {
				t.Fatalf("job %d: unexpected error: %v", i, res.Err)
			}
			if string(res.Body) != "ok" {
				t.Fatalf("job %d: expected body %q, got %q", i, "ok", string(res.Body))
			}
		case <-ctx.Done():
			t.Fatalf("timed out waiting for job %d result", i)
		}
	}

	if got := atomic.LoadInt32(&hits); got != n {
		t.Fatalf("expected server hits %d, got %d", n, got)
	}
}
