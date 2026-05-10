package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aipivot/pkg/llm"
)

// mockTool 用于测试的 mock 工具
type mockTool struct {
	name   string
	result string
}

func (t *mockTool) Name() string        { return t.name }
func (t *mockTool) Description() string  { return "mock tool for testing" }
func (t *mockTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Type: "function",
		Function: llm.FunctionDefinition{
			Name:        t.name,
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}
func (t *mockTool) Execute(_ context.Context, _ string) (string, error) {
	return t.result, nil
}

func TestAgent_Run_NoTools(t *testing.T) {
	// 模拟 LLM 直接返回文本回复
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := llm.ChatResponse{
			ID:    "test-1",
			Model: "gpt-4o",
			Choices: []llm.ChatChoice{
				{
					Index:        0,
					Message:      llm.ChatMessage{Role: "assistant", Content: "Hello!"},
					FinishReason: "stop",
				},
			},
			Usage: llm.ChatUsage{TotalTokens: 10},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-key", 30)
	registry := NewRegistry()
	ag := NewAgent(client, registry, 5)

	result, err := ag.Run(context.Background(), &RunRequest{
		Model:       "gpt-4o",
		Messages:    []llm.ChatMessage{{Role: "user", Content: "Hi"}},
		MaxTokens:   100,
		Temperature: 0.7,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Content != "Hello!" {
		t.Errorf("expected content 'Hello!', got %q", result.Content)
	}
	if result.TotalRound != 1 {
		t.Errorf("expected 1 round, got %d", result.TotalRound)
	}
	if len(result.ToolUses) != 0 {
		t.Errorf("expected no tool uses, got %d", len(result.ToolUses))
	}
}

func TestAgent_Run_WithToolCall(t *testing.T) {
	callCount := 0
	// 模拟 LLM：第一次返回 tool_calls，第二次返回文本
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp llm.ChatResponse

		if callCount == 1 {
			// 第一轮：LLM 决定调用工具
			resp = llm.ChatResponse{
				ID:    "test-1",
				Model: "gpt-4o",
				Choices: []llm.ChatChoice{
					{
						Index: 0,
						Message: llm.ChatMessage{
							Role: "assistant",
							ToolCalls: []llm.ToolCall{
								{
									ID:   "call_123",
									Type: "function",
									Function: llm.FunctionCall{
										Name:      "mock_tool",
										Arguments: "{}",
									},
								},
							},
						},
						FinishReason: "tool_calls",
					},
				},
				Usage: llm.ChatUsage{TotalTokens: 5},
			}
		} else {
			// 第二轮：LLM 使用工具结果生成最终回复
			resp = llm.ChatResponse{
				ID:    "test-2",
				Model: "gpt-4o",
				Choices: []llm.ChatChoice{
					{
						Index:        0,
						Message:      llm.ChatMessage{Role: "assistant", Content: "Tool says: mock result"},
						FinishReason: "stop",
					},
				},
				Usage: llm.ChatUsage{TotalTokens: 15},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-key", 30)
	registry := NewRegistry()
	registry.Register(&mockTool{name: "mock_tool", result: "mock result"})
	ag := NewAgent(client, registry, 5)

	result, err := ag.Run(context.Background(), &RunRequest{
		Model:       "gpt-4o",
		Messages:    []llm.ChatMessage{{Role: "user", Content: "use the tool"}},
		MaxTokens:   100,
		Temperature: 0.7,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(result.Content, "Tool says") {
		t.Errorf("expected content about tool result, got %q", result.Content)
	}
	if result.TotalRound != 2 {
		t.Errorf("expected 2 rounds, got %d", result.TotalRound)
	}
	if len(result.ToolUses) != 1 {
		t.Fatalf("expected 1 tool use, got %d", len(result.ToolUses))
	}
	if result.ToolUses[0].Name != "mock_tool" {
		t.Errorf("expected tool name 'mock_tool', got %q", result.ToolUses[0].Name)
	}
	if result.ToolUses[0].Result != "mock result" {
		t.Errorf("expected tool result 'mock result', got %q", result.ToolUses[0].Result)
	}
}

func TestAgent_Run_MaxRoundsExceeded(t *testing.T) {
	// LLM 始终返回 tool_calls，测试 maxRounds 限制
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := llm.ChatResponse{
			ID:    "test-loop",
			Model: "gpt-4o",
			Choices: []llm.ChatChoice{
				{
					Index: 0,
					Message: llm.ChatMessage{
						Role: "assistant",
						ToolCalls: []llm.ToolCall{
							{
								ID:   "call_loop",
								Type: "function",
								Function: llm.FunctionCall{
									Name:      "mock_tool",
									Arguments: "{}",
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := llm.NewClient(server.URL, "test-key", 30)
	registry := NewRegistry()
	registry.Register(&mockTool{name: "mock_tool", result: "loop"})
	ag := NewAgent(client, registry, 3)

	_, err := ag.Run(context.Background(), &RunRequest{
		Model:       "gpt-4o",
		Messages:    []llm.ChatMessage{{Role: "user", Content: "loop"}},
		MaxTokens:   100,
		Temperature: 0.7,
	})
	if err == nil {
		t.Fatal("expected error for max rounds exceeded")
	}
	if !strings.Contains(err.Error(), "max rounds") {
		t.Errorf("expected 'max rounds' in error, got: %v", err)
	}
}

func TestRegistry_ExecuteUnknownTool(t *testing.T) {
	registry := NewRegistry()
	_, err := registry.Execute(context.Background(), "nonexistent", "{}")
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}
