package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"aipivot/pkg/llm"
)

// CalculatorTool 简单数学计算工具，支持四则运算和常用数学函数。
type CalculatorTool struct{}

type calcArgs struct {
	Expression string `json:"expression"` // 数学表达式
}

type calcResult struct {
	Expression string  `json:"expression"`
	Result     float64 `json:"result"`
}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (t *CalculatorTool) Name() string        { return "calculator" }
func (t *CalculatorTool) Description() string  { return "执行数学计算，支持加减乘除、幂运算和常用数学函数" }

func (t *CalculatorTool) Definition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Type: "function",
		Function: llm.FunctionDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"expression": map[string]interface{}{
						"type":        "string",
						"description": "数学表达式，支持: +, -, *, /, ^(幂), sqrt(x), abs(x), round(x)。例如: '2 + 3 * 4', 'sqrt(144)', '2 ^ 10'",
					},
				},
				"required": []string{"expression"},
			},
		},
	}
}

func (t *CalculatorTool) Execute(_ context.Context, arguments string) (string, error) {
	var args calcArgs
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("parse calc args: %w", err)
	}

	expr := strings.TrimSpace(args.Expression)
	if expr == "" {
		return "", fmt.Errorf("expression is required")
	}

	result, err := evalExpression(expr)
	if err != nil {
		return "", fmt.Errorf("calculate %q: %w", expr, err)
	}

	data, _ := json.Marshal(calcResult{
		Expression: expr,
		Result:     result,
	})
	return string(data), nil
}

// evalExpression 简单表达式求值，支持常用数学函数。
// MVP 阶段仅处理单函数调用和二元运算，复杂表达式由 LLM 拆解为多步调用。
func evalExpression(expr string) (float64, error) {
	expr = strings.TrimSpace(expr)

	// 处理函数调用: sqrt(x), abs(x), round(x)
	for _, fn := range []struct {
		name string
		eval func(float64) float64
	}{
		{"sqrt", math.Sqrt},
		{"abs", math.Abs},
		{"round", math.Round},
	} {
		if strings.HasPrefix(expr, fn.name+"(") && strings.HasSuffix(expr, ")") {
			inner := expr[len(fn.name)+1 : len(expr)-1]
			val, err := strconv.ParseFloat(strings.TrimSpace(inner), 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number in %s(): %w", fn.name, err)
			}
			return fn.eval(val), nil
		}
	}

	// 处理二元运算（从右向左扫描低优先级运算符，保证运算顺序）
	// 优先级：+- 低于 */ 低于 ^
	for _, ops := range []string{"+-", "*/", "^"} {
		// 从右向左扫描，确保左结合
		depth := 0
		for i := len(expr) - 1; i >= 0; i-- {
			ch := expr[i]
			if ch == ')' {
				depth++
			} else if ch == '(' {
				depth--
			} else if depth == 0 && strings.ContainsRune(ops, rune(ch)) && i > 0 {
				left, err := evalExpression(expr[:i])
				if err != nil {
					return 0, err
				}
				right, err := evalExpression(expr[i+1:])
				if err != nil {
					return 0, err
				}
				switch ch {
				case '+':
					return left + right, nil
				case '-':
					return left - right, nil
				case '*':
					return left * right, nil
				case '/':
					if right == 0 {
						return 0, fmt.Errorf("division by zero")
					}
					return left / right, nil
				case '^':
					return math.Pow(left, right), nil
				}
			}
		}
	}

	// 处理括号
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		return evalExpression(expr[1 : len(expr)-1])
	}

	// 纯数字
	val, err := strconv.ParseFloat(expr, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as number", expr)
	}
	return val, nil
}
