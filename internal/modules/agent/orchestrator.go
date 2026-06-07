package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"aipivot/pkg/llm"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultMaxWorkers = 3
	maxPlanTasks      = 5
)

// WorkerTask is a single task assigned by the orchestrator to a worker agent.
type WorkerTask struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Objective string `json:"objective"`
}

// WorkerResult records the output of one worker agent run.
type WorkerResult struct {
	ID        string
	Role      string
	Objective string
	Content   string
	Usage     *llm.ChatUsage
	ToolUses  []ToolUseRecord
}

// Orchestrator coordinates a planner, multiple worker agents, and a final
// synthesis step. Workers reuse the existing ReAct Agent so tool calling keeps
// the same execution path and audit records.
type Orchestrator struct {
	llmClient  *llm.Client
	worker     *Agent
	maxWorkers int
}

type planResponse struct {
	Tasks []WorkerTask `json:"tasks"`
}

func NewOrchestrator(llmClient *llm.Client, worker *Agent, maxWorkers int) *Orchestrator {
	if maxWorkers <= 0 {
		maxWorkers = defaultMaxWorkers
	}
	if maxWorkers > maxPlanTasks {
		maxWorkers = maxPlanTasks
	}
	return &Orchestrator{
		llmClient:  llmClient,
		worker:     worker,
		maxWorkers: maxWorkers,
	}
}

func (o *Orchestrator) Run(ctx context.Context, req *RunRequest) (*RunResult, error) {
	if o == nil || o.worker == nil {
		return nil, fmt.Errorf("orchestrator: worker agent not configured")
	}

	tasks, planUsage, err := o.plan(ctx, req)
	if err != nil {
		logx.WithContext(ctx).Errorf("orchestrator plan failed, fallback to worker: %v", err)
		return o.worker.Run(ctx, req)
	}
	if len(tasks) < 2 {
		return o.worker.Run(ctx, req)
	}

	results, err := o.runWorkers(ctx, req, tasks)
	if err != nil {
		return nil, err
	}

	final, synthUsage, model, err := o.synthesize(ctx, req, results)
	if err != nil {
		return nil, err
	}

	usage := mergeUsage(planUsage, synthUsage)
	var toolUses []ToolUseRecord
	for _, result := range results {
		toolUses = append(toolUses, result.ToolUses...)
		usage = mergeUsage(usage, result.Usage)
	}

	return &RunResult{
		Content:       final,
		Model:         model,
		Usage:         usage,
		ToolUses:      toolUses,
		WorkerResults: results,
		TotalRound:    len(results) + 2,
	}, nil
}

func (o *Orchestrator) plan(ctx context.Context, req *RunRequest) ([]WorkerTask, *llm.ChatUsage, error) {
	prompt := fmt.Sprintf(`You are the AIPivot task orchestrator.
Decide whether the user request needs multiple specialist workers.
Return only compact JSON with this schema:
{"tasks":[{"id":"task-1","role":"specialist role","objective":"specific worker objective"}]}
Use 1 task for simple requests. Use at most %d tasks.`, o.maxWorkers)

	messages := append([]llm.ChatMessage{{Role: "system", Content: prompt}}, req.Messages...)
	resp, err := o.llmClient.ChatCompletion(ctx, &llm.ChatRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   600,
		Temperature: 0,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("planner LLM call: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, &resp.Usage, fmt.Errorf("planner returned empty choices")
	}

	var plan planResponse
	content := stripJSONFence(resp.Choices[0].Message.Content)
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, &resp.Usage, fmt.Errorf("parse planner JSON: %w", err)
	}

	tasks := sanitizeTasks(plan.Tasks, o.maxWorkers)
	return tasks, &resp.Usage, nil
}

func (o *Orchestrator) runWorkers(ctx context.Context, req *RunRequest, tasks []WorkerTask) ([]WorkerResult, error) {
	results := make([]WorkerResult, len(tasks))
	errs := make(chan error, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		i, task := i, task
		wg.Add(1)
		go func() {
			defer wg.Done()
			messages := buildWorkerMessages(req.Messages, task)
			result, err := o.worker.Run(ctx, &RunRequest{
				Model:       req.Model,
				Messages:    messages,
				MaxTokens:   req.MaxTokens,
				Temperature: req.Temperature,
				ExtraTools:  req.ExtraTools,
			})
			if err != nil {
				errs <- fmt.Errorf("worker %s run: %w", task.ID, err)
				return
			}
			results[i] = WorkerResult{
				ID:        task.ID,
				Role:      task.Role,
				Objective: task.Objective,
				Content:   result.Content,
				Usage:     result.Usage,
				ToolUses:  result.ToolUses,
			}
		}()
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (o *Orchestrator) synthesize(ctx context.Context, req *RunRequest, results []WorkerResult) (string, *llm.ChatUsage, string, error) {
	var report strings.Builder
	report.WriteString("Worker results:\n")
	for _, result := range results {
		report.WriteString(fmt.Sprintf("\n[%s] %s\nObjective: %s\nResult:\n%s\n", result.ID, result.Role, result.Objective, result.Content))
	}

	messages := append([]llm.ChatMessage{
		{
			Role:    "system",
			Content: "You are the AIPivot orchestrator synthesizer. Combine worker findings into one concise, accurate answer for the end user. Do not mention internal worker mechanics unless the user asks.",
		},
	}, req.Messages...)
	messages = append(messages, llm.ChatMessage{Role: "user", Content: report.String()})

	resp, err := o.llmClient.ChatCompletion(ctx, &llm.ChatRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	})
	if err != nil {
		return "", nil, "", fmt.Errorf("synthesizer LLM call: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", &resp.Usage, resp.Model, fmt.Errorf("synthesizer returned empty choices")
	}
	return resp.Choices[0].Message.Content, &resp.Usage, resp.Model, nil
}

func buildWorkerMessages(base []llm.ChatMessage, task WorkerTask) []llm.ChatMessage {
	system := fmt.Sprintf("You are a %s worker agent. Focus only on this objective: %s. Use tools when they help, then return findings for the orchestrator.", task.Role, task.Objective)
	messages := make([]llm.ChatMessage, 0, len(base)+2)
	messages = append(messages, llm.ChatMessage{Role: "system", Content: system})
	messages = append(messages, base...)
	messages = append(messages, llm.ChatMessage{Role: "user", Content: "Worker objective: " + task.Objective})
	return messages
}

func sanitizeTasks(tasks []WorkerTask, limit int) []WorkerTask {
	if limit <= 0 {
		limit = defaultMaxWorkers
	}
	if len(tasks) > limit {
		tasks = tasks[:limit]
	}

	result := make([]WorkerTask, 0, len(tasks))
	for i, task := range tasks {
		task.ID = strings.TrimSpace(task.ID)
		task.Role = strings.TrimSpace(task.Role)
		task.Objective = strings.TrimSpace(task.Objective)
		if task.Objective == "" {
			continue
		}
		if task.ID == "" {
			task.ID = fmt.Sprintf("task-%d", i+1)
		}
		if task.Role == "" {
			task.Role = "generalist"
		}
		result = append(result, task)
	}
	return result
}

func stripJSONFence(content string) string {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "```") {
		return content
	}
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}

func mergeUsage(a, b *llm.ChatUsage) *llm.ChatUsage {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return &llm.ChatUsage{
		PromptTokens:     a.PromptTokens + b.PromptTokens,
		CompletionTokens: a.CompletionTokens + b.CompletionTokens,
		TotalTokens:      a.TotalTokens + b.TotalTokens,
	}
}
