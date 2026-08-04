package fasthttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	fasthttpclient "github.com/valyala/fasthttp"
)

func TestDoer_GET(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.Header.Get("X-Test"); got != "yes" {
			t.Errorf("X-Test = %q, want \"yes\"", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	d := New()
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("X-Test", "yes")

	resp, err := d.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if resp.Status != "200 OK" {
		t.Errorf("Status = %q, want %q", resp.Status, "200 OK")
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("body = %q", body)
	}
}

func TestDoer_POSTWithBody(t *testing.T) {
	t.Parallel()
	want := []byte(`{"hello":"world"}`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ := io.ReadAll(r.Body)
		if !bytes.Equal(got, want) {
			t.Errorf("server received body = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	d := New()
	req, err := http.NewRequest(http.MethodPost, srv.URL, bytes.NewReader(want))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := d.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want 201", resp.StatusCode)
	}
}

func TestDoer_ContextCancellation(t *testing.T) {
	t.Parallel()
	// Slow server: holds the request for longer than the context allows.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		_, _ = io.WriteString(w, "late")
	}))
	defer srv.Close()

	d := New()
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	resp, err := d.Do(ctx, req)
	elapsed := time.Since(start)

	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
	if !strings.Contains(err.Error(), "context") && !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context-related error, got: %v", err)
	}
	if elapsed > 250*time.Millisecond {
		t.Errorf("Do took %v; expected to return promptly after ctx deadline", elapsed)
	}
}

type blockingClient struct {
	started     chan struct{}
	finish      chan struct{}
	observedURI chan string
}

func (c *blockingClient) Do(req *fasthttpclient.Request, _ *fasthttpclient.Response) error {
	close(c.started)
	<-c.finish
	c.observedURI <- req.URI().String()
	return nil
}

func (c *blockingClient) DoDeadline(
	_ *fasthttpclient.Request,
	_ *fasthttpclient.Response,
	_ time.Time,
) error {
	panic("unexpected DoDeadline call")
}

func TestDoer_ContextCancellationWithoutDeadlineKeepsRequestAlive(t *testing.T) {
	t.Parallel()

	client := &blockingClient{
		started:     make(chan struct{}),
		finish:      make(chan struct{}),
		observedURI: make(chan string, 1),
	}
	defer func() {
		select {
		case <-client.finish:
		default:
			close(client.finish)
		}
	}()
	d := &Doer{client: client}
	req, err := http.NewRequest(http.MethodGet, "http://example.com/v1", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		resp, doErr := d.Do(ctx, req)
		if resp != nil {
			_ = resp.Body.Close()
		}
		result <- doErr
	}()

	<-client.started
	cancel()

	select {
	case doErr := <-result:
		if !errors.Is(doErr, context.Canceled) {
			t.Errorf("Do error = %v, want context.Canceled", doErr)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("Do did not return promptly after cancellation")
	}

	close(client.finish)
	select {
	case got := <-client.observedURI:
		if got != req.URL.String() {
			t.Errorf("in-flight request URI = %q, want %q", got, req.URL.String())
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("in-flight request did not finish")
	}
}

type deadlineClient struct {
	deadlineReached <-chan struct{}
}

func (c *deadlineClient) Do(_ *fasthttpclient.Request, _ *fasthttpclient.Response) error {
	panic("unexpected Do call")
}

func (c *deadlineClient) DoDeadline(
	_ *fasthttpclient.Request,
	_ *fasthttpclient.Response,
	_ time.Time,
) error {
	<-c.deadlineReached
	return fasthttpclient.ErrTimeout
}

func TestExecuteWithContext_DeadlineKeepsCallerOwnership(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	d := &Doer{client: &deadlineClient{deadlineReached: ctx.Done()}}
	req := fasthttpclient.AcquireRequest()
	defer fasthttpclient.ReleaseRequest(req)
	resp := fasthttpclient.AcquireResponse()
	defer fasthttpclient.ReleaseResponse(resp)

	cleanupAsync, err := d.executeWithContext(ctx, req, resp)

	if cleanupAsync {
		t.Error("deadline path transferred resource ownership to asynchronous cleanup")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("executeWithContext error = %v, want context.DeadlineExceeded", err)
	}
}

func TestDoer_NonOKStatus(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadRequest)
	}))
	defer srv.Close()

	d := New()
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	resp, err := d.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", resp.StatusCode)
	}
}
