package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aipivot/pkg/llm"
)

// TimeTool 获取当前日期时间工具。
type TimeTool struct{}

type timeArgs struct {
	Timezone string `json:"timezone"` // 可选，如 "Asia/Shanghai"
}

type timeResult struct {
	Datetime  string `json:"datetime"`
	Timezone  string `json:"timezone"`
	Timestamp int64  `json:"timestamp"`
	Weekday   string `json:"weekday"`
}

func NewTimeTool() *TimeTool {
	return &TimeTool{}
}

func (t *TimeTool) Name() string        { return "get_current_time" }
func (t *TimeTool) Description() string  { return "获取当前日期和时间信息" }

func (t *TimeTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Type: "function",
		Function: llm.FunctionDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"timezone": map[string]interface{}{
						"type":        "string",
						"description": "时区标识，如 Asia/Shanghai、America/New_York，默认 Asia/Shanghai",
					},
				},
			},
		},
	}
}

func (t *TimeTool) Execute(_ context.Context, arguments string) (string, error) {
	var args timeArgs
	if arguments != "" && arguments != "{}" {
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("parse time args: %w", err)
		}
	}

	tz := "Asia/Shanghai"
	if args.Timezone != "" {
		tz = args.Timezone
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", fmt.Errorf("invalid timezone %q: %w", tz, err)
	}

	now := time.Now().In(loc)

	weekdayNames := map[time.Weekday]string{
		time.Monday: "星期一", time.Tuesday: "星期二", time.Wednesday: "星期三",
		time.Thursday: "星期四", time.Friday: "星期五", time.Saturday: "星期六", time.Sunday: "星期日",
	}

	result := timeResult{
		Datetime:  now.Format("2006-01-02 15:04:05"),
		Timezone:  tz,
		Timestamp: now.Unix(),
		Weekday:   weekdayNames[now.Weekday()],
	}

	data, _ := json.Marshal(result)
	return string(data), nil
}
