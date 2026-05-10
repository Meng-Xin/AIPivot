package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"aipivot/pkg/llm"
)

// WeatherTool 天气查询工具（Mock 实现，MVP 阶段验证 Function Calling 链路）。
type WeatherTool struct{}

type weatherArgs struct {
	City string `json:"city"`
}

type weatherResult struct {
	City        string `json:"city"`
	Temperature int    `json:"temperature"`
	Condition   string `json:"condition"`
	Humidity    int    `json:"humidity"`
	Wind        string `json:"wind"`
}

func NewWeatherTool() *WeatherTool {
	return &WeatherTool{}
}

func (t *WeatherTool) Name() string        { return "get_weather" }
func (t *WeatherTool) Description() string  { return "获取指定城市的当前天气信息" }

func (t *WeatherTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Type: "function",
		Function: llm.FunctionDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"city": map[string]interface{}{
						"type":        "string",
						"description": "城市名称，如 北京、上海、New York",
					},
				},
				"required": []string{"city"},
			},
		},
	}
}

func (t *WeatherTool) Execute(_ context.Context, arguments string) (string, error) {
	var args weatherArgs
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("parse weather args: %w", err)
	}
	if args.City == "" {
		return "", fmt.Errorf("city is required")
	}

	// Mock 天气数据
	conditions := []string{"晴", "多云", "阴", "小雨", "大雨", "雷阵雨", "雪"}
	winds := []string{"微风", "东北风 3 级", "西南风 4 级", "北风 2 级"}

	result := weatherResult{
		City:        args.City,
		Temperature: rand.Intn(35) + 5, // 5~39°C
		Condition:   conditions[rand.Intn(len(conditions))],
		Humidity:    rand.Intn(60) + 30, // 30~89%
		Wind:        winds[rand.Intn(len(winds))],
	}

	data, _ := json.Marshal(result)
	return string(data), nil
}
