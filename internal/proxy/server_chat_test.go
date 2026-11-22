package proxy

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"quickwork-openai/internal/accounts"
	"quickwork-openai/internal/config"
	"quickwork-openai/internal/store"
)

// newTestServer wires a real store + account manager against the fake
// WebSocket upstream, so handler behaviour is covered end to end.
func newTestServer(t *testing.T, upstreamURL string) *Server {
	t.Helper()

	st, err := store.Open(filepath.Join(t.TempDir(), "quickwork.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	acct := &store.Account{
		Email:       "test@example.com",
		Provider:    "social",
		WSHost:      strings.Replace(upstreamURL, "http://", "ws://", 1),
		WSPath:      "/ws",
		AccessToken: "test-jwt",
		ExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
	}
	if err := st.Save(context.Background(), acct); err != nil {
		t.Fatalf("save account: %v", err)
	}

	cfg := config.Default()
	srv := NewServer(&cfg, accounts.New(st, &cfg))
	return srv
}

func TestChatHandlerReturnsReasoningContent(t *testing.T) {
	up := fakeUpstream(t,
		`{"type":"thinking_delta","payload":{"delta":"think "}}`,
		`{"type":"thinking_delta","payload":{"delta":"hard"}}`,
		`{"type":"text_delta","payload":{"delta":"hello"}}`,
		`{"type":"turn_completed"}`,
	)
	handler := newTestServer(t, up.URL).Handler()

	body := `{"model":"smart","messages":[{"role":"user","content":"hi"}]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Choices []struct {
			Message map[string]any `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v (%s)", err, rec.Body.String())
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("choices = %d", len(resp.Choices))
	}
	msg := resp.Choices[0].Message
	if msg["reasoning_content"] != "think hard" {
		t.Fatalf("reasoning_content = %v", msg["reasoning_content"])
	}
	if msg["content"] != "hello" {
		t.Fatalf("content = %v", msg["content"])
	}
}

func TestChatHandlerStreamsReasoningContent(t *testing.T) {
	up := fakeUpstream(t,
		`{"type":"thinking_delta","payload":{"delta":"think "}}`,
		`{"type":"text_delta","payload":{"delta":"hello"}}`,
		`{"type":"turn_completed"}`,
	)
	handler := newTestServer(t, up.URL).Handler()

	body := `{"model":"smart","stream":true,"messages":[{"role":"user","content":"hi"}]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	out := rec.Body.String()
	for _, want := range []string{
		`"reasoning_content":"think "`,
		`"content":"hello"`,
		"data: [DONE]",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stream missing %q:\n%s", want, out)
		}
	}
}
