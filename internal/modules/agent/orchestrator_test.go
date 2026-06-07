package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"aipivot/pkg/llm"
)

func TestOrchestrator_Run_MultipleWorkers(t *testing.T) {
	var mu sync.Mutex
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req llm.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		mu.Lock()
		callCount++
		mu.Unlock()

		content := flattenMessages(req.Messages)
		resp := llm.ChatResponse{
			ID:    "test",
			Model: "gpt-4o",
			Usage: llm.ChatUsage{PromptTokens: 1, CompletionTokens: 2, TotalTokens: 3},
		}

		switch {
		case strings.Contains(content, "AIPivot task orchestrator"):
			resp.Choices = []llm.ChatChoice{{
				Message: llm.ChatMessage{Role: "assistant", Content: `{"tasks":[{"id":"task-1","role":"billing analyst","objective":"Analyze pricing risk"},{"id":"task-2","role":"support analyst","objective":"Analyze support impact"}]}`},
			}}
		case strings.Contains(content, "Worker results:"):
			if !strings.Contains(content, "pricing finding") || !strings.Contains(content, "support finding") {
				t.Fatalf("synthesis missing worker findings: %s", content)
			}
			resp.Choices = []llm.ChatChoice{{
				Message: llm.ChatMessage{Role: "assistant", Content: "final synthesized answer"},
			}}
		case strings.Contains(content, "Analyze pricing risk"):
			resp.Choices = []llm.ChatChoice{{
				Message: llm.ChatMessage{Role: "assistant", Content: "pricing finding"},
			}}
		case strings.Contains(content, "Analyze support impact"):
			resp.Choices = []llm.ChatChoice{{
				Message: llm.ChatMessage{Role: "assistant", Content: "support finding"},
			}}
		default:
			t.Fatalf("unexpected request messages: %s", content)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-key", 30)
	worker := NewAgent(client, NewRegistry(), 3)
	orchestrator := NewOrchestrator(client, worker, 3)

	result, err := orchestrator.Run(context.Background(), &RunRequest{
		Model:       "gpt-4o",
		Messages:    []llm.ChatMessage{{Role: "user", Content: "Compare pricing and support tradeoffs"}},
		MaxTokens:   500,
		Temperature: 0.2,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Content != "final synthesized answer" {
		t.Fatalf("unexpected content: %q", result.Content)
	}
	if len(result.WorkerResults) != 2 {
		t.Fatalf("expected 2 worker results, got %d", len(result.WorkerResults))
	}
	if result.Usage == nil || result.Usage.TotalTokens != 12 {
		t.Fatalf("expected planner+workers+synth usage total 12, got %#v", result.Usage)
	}

	mu.Lock()
	defer mu.Unlock()
	if callCount != 4 {
		t.Fatalf("expected 4 LLM calls, got %d", callCount)
	}
}

func TestOrchestrator_Run_SingleTaskFallback(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var req llm.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		content := flattenMessages(req.Messages)
		resp := llm.ChatResponse{
			ID:    "test",
			Model: "gpt-4o",
			Usage: llm.ChatUsage{TotalTokens: 3},
		}
		if strings.Contains(content, "AIPivot task orchestrator") {
			resp.Choices = []llm.ChatChoice{{
				Message: llm.ChatMessage{Role: "assistant", Content: `{"tasks":[{"id":"task-1","role":"generalist","objective":"Answer directly"}]}`},
			}}
		} else {
			resp.Choices = []llm.ChatChoice{{
				Message: llm.ChatMessage{Role: "assistant", Content: "direct worker answer"},
			}}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-key", 30)
	worker := NewAgent(client, NewRegistry(), 3)
	orchestrator := NewOrchestrator(client, worker, 3)

	result, err := orchestrator.Run(context.Background(), &RunRequest{
		Model:    "gpt-4o",
		Messages: []llm.ChatMessage{{Role: "user", Content: "Say hello"}},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Content != "direct worker answer" {
		t.Fatalf("unexpected fallback content: %q", result.Content)
	}
	if len(result.WorkerResults) != 0 {
		t.Fatalf("expected no worker result records in fallback, got %d", len(result.WorkerResults))
	}
	if callCount != 2 {
		t.Fatalf("expected planner + direct worker calls, got %d", callCount)
	}
}

func flattenMessages(messages []llm.ChatMessage) string {
	var b strings.Builder
	for _, msg := range messages {
		b.WriteString(msg.Role)
		b.WriteString(":")
		b.WriteString(msg.Content)
		b.WriteString("\n")
	}
	return b.String()
}
