package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestWeatherTool_Execute(t *testing.T) {
	tool := NewWeatherTool()

	if tool.Name() != "get_weather" {
		t.Errorf("expected name 'get_weather', got %q", tool.Name())
	}

	result, err := tool.Execute(context.Background(), `{"city":"北京"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}
	if data["city"] != "北京" {
		t.Errorf("expected city '北京', got %v", data["city"])
	}
}

func TestWeatherTool_Execute_EmptyCity(t *testing.T) {
	tool := NewWeatherTool()
	_, err := tool.Execute(context.Background(), `{"city":""}`)
	if err == nil {
		t.Fatal("expected error for empty city")
	}
}

func TestTimeTool_Execute(t *testing.T) {
	tool := NewTimeTool()

	if tool.Name() != "get_current_time" {
		t.Errorf("expected name 'get_current_time', got %q", tool.Name())
	}

	result, err := tool.Execute(context.Background(), `{"timezone":"Asia/Shanghai"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}
	if data["timezone"] != "Asia/Shanghai" {
		t.Errorf("expected timezone 'Asia/Shanghai', got %v", data["timezone"])
	}
}

func TestTimeTool_Execute_InvalidTimezone(t *testing.T) {
	tool := NewTimeTool()
	_, err := tool.Execute(context.Background(), `{"timezone":"Invalid/Zone"}`)
	if err == nil {
		t.Fatal("expected error for invalid timezone")
	}
}

func TestCalculatorTool_Execute(t *testing.T) {
	tool := NewCalculatorTool()

	if tool.Name() != "calculator" {
		t.Errorf("expected name 'calculator', got %q", tool.Name())
	}

	tests := []struct {
		expr   string
		expect float64
	}{
		{"2+3", 5},
		{"10-4", 6},
		{"3*4", 12},
		{"10/2", 5},
		{"2^10", 1024},
		{"sqrt(144)", 12},
		{"abs(-5)", 5},
	}

	for _, tc := range tests {
		args, _ := json.Marshal(map[string]string{"expression": tc.expr})
		result, err := tool.Execute(context.Background(), string(args))
		if err != nil {
			t.Errorf("expr %q: unexpected error: %v", tc.expr, err)
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(result), &data); err != nil {
			t.Errorf("expr %q: result not valid JSON: %v", tc.expr, err)
			continue
		}
		if data["result"] != tc.expect {
			t.Errorf("expr %q: expected %v, got %v", tc.expr, tc.expect, data["result"])
		}
	}
}

func TestCalculatorTool_Execute_DivByZero(t *testing.T) {
	tool := NewCalculatorTool()
	args, _ := json.Marshal(map[string]string{"expression": "1/0"})
	_, err := tool.Execute(context.Background(), string(args))
	if err == nil {
		t.Fatal("expected error for division by zero")
	}
	if !strings.Contains(err.Error(), "division by zero") {
		t.Errorf("expected 'division by zero' in error, got: %v", err)
	}
}
