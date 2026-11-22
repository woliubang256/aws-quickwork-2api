package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"quickwork-openai/internal/config"
)

// fakeUpstream starts a WebSocket server that replays frames and drains
// whatever the client sends, so Upstream.Chat can be exercised offline.
func fakeUpstream(t *testing.T, frames ...string) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()
		for _, frame := range frames {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(frame)); err != nil {
				return
			}
		}
		time.Sleep(200 * time.Millisecond)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func sessionFor(srv *httptest.Server) *Session {
	return &Session{
		JWT:      "test-jwt",
		WSHost:   strings.Replace(srv.URL, "http://", "ws://", 1),
		WSPath:   "/ws",
		Provider: "social",
	}
}

func TestChatSurfacesReasoningDeltas(t *testing.T) {
	srv := fakeUpstream(t,
		`{"type":"thinking_delta","payload":{"delta":"step 1 "}}`,
		`{"type":"thinking_delta","payload":"step 2"}`,
		`{"type":"text_delta","payload":{"delta":"answer"}}`,
		`{"type":"turn_completed"}`,
	)

	cfg := config.Default()
	up := NewUpstream(&cfg)

	var events []StreamEvent
	res, err := up.Chat(context.Background(), sessionFor(srv), ChatInput{Prompt: "hi", Model: "smart"}, func(ev StreamEvent) error {
		events = append(events, ev)
		return nil
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if res.Reasoning != "step 1 step 2" {
		t.Fatalf("Reasoning = %q", res.Reasoning)
	}
	if res.Text != "answer" {
		t.Fatalf("Text = %q", res.Text)
	}

	var reasoning strings.Builder
	var text strings.Builder
	for _, ev := range events {
		switch ev.Type {
		case "reasoning":
			reasoning.WriteString(ev.Text)
		case "text":
			text.WriteString(ev.Text)
		}
	}
	if reasoning.String() != "step 1 step 2" {
		t.Fatalf("streamed reasoning = %q", reasoning.String())
	}
	if text.String() != "answer" {
		t.Fatalf("streamed text = %q", text.String())
	}
}

func TestChatReasoningCompleteFallback(t *testing.T) {
	srv := fakeUpstream(t,
		`{"type":"thinking_complete","payload":{"thinking":"full thought"}}`,
		`{"type":"text_delta","payload":{"delta":"ok"}}`,
		`{"type":"turn_completed"}`,
	)

	cfg := config.Default()
	up := NewUpstream(&cfg)

	var streamed []string
	res, err := up.Chat(context.Background(), sessionFor(srv), ChatInput{Prompt: "hi"}, func(ev StreamEvent) error {
		if ev.Type == "reasoning" {
			streamed = append(streamed, ev.Text)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if res.Reasoning != "full thought" {
		t.Fatalf("Reasoning = %q", res.Reasoning)
	}
	if len(streamed) != 1 || streamed[0] != "full thought" {
		t.Fatalf("streamed reasoning = %v", streamed)
	}
}

func TestUsageMapReportsReasoningTokens(t *testing.T) {
	m := usageMap(Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15, ReasoningTokens: 3})
	details, ok := m["completion_tokens_details"].(map[string]any)
	if !ok {
		t.Fatalf("missing completion_tokens_details: %v", m)
	}
	if details["reasoning_tokens"] != 3 {
		t.Fatalf("reasoning_tokens = %v", details["reasoning_tokens"])
	}

	plain := usageMap(Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2})
	if _, ok := plain["completion_tokens_details"]; ok {
		t.Fatal("completion_tokens_details must be omitted when there are no reasoning tokens")
	}
}
