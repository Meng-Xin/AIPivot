package flow

import (
	"fmt"
	"strconv"
	"strings"
)

// 表达式语法（一期刻意极简，零第三方依赖）：
//   - 单比较：confidence < 0.6 / status == "ok" / message != ""
//   - 裸标识符：truthy 判定（bool 直接取值，数值 !=0，字符串 !=""）
// 操作数支持：数字字面量、单/双引号字符串字面量、黑板变量名。
// 解析或求值失败返回 error，由调用方 fail-soft（视为 true + warning），不中断流程。

var exprOperators = []string{"==", "!=", "<=", ">=", "<", ">"}

// EvalExpression 对黑板上下文求值单个比较表达式。
func EvalExpression(expr string, bb *Blackboard) (bool, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false, fmt.Errorf("表达式为空")
	}

	for _, op := range exprOperators {
		if idx := strings.Index(expr, op); idx > 0 {
			left := strings.TrimSpace(expr[:idx])
			right := strings.TrimSpace(expr[idx+len(op):])
			if left == "" || right == "" {
				return false, fmt.Errorf("表达式操作数缺失: %q", expr)
			}
			return compareOperands(left, right, op, bb)
		}
	}

	// 裸标识符：truthy 判定
	v, err := resolveOperand(expr, bb)
	if err != nil {
		return false, err
	}
	return truthy(v), nil
}

// compareOperands 比较 左 op 右。数值语境优先（两侧都可解析为数值），
// 否则按字符串语境比较（==/!= 全支持，< <= > >= 按字典序）。
func compareOperands(left, right, op string, bb *Blackboard) (bool, error) {
	lv, err := resolveOperand(left, bb)
	if err != nil {
		return false, err
	}
	rv, err := resolveOperand(right, bb)
	if err != nil {
		return false, err
	}

	if lf, lok := toFloat(lv); lok {
		if rf, rok := toFloat(rv); rok {
			return compareOrdered(lf, rf, op), nil
		}
	}

	ls, lok := lv.(string)
	rs, rok := rv.(string)
	if !lok || !rok {
		return false, fmt.Errorf("类型不可比较: %v %s %v", lv, op, rv)
	}
	switch op {
	case "==":
		return ls == rs, nil
	case "!=":
		return ls != rs, nil
	default:
		return compareOrdered(ls, rs, op), nil
	}
}

func compareOrdered[T float64 | string](a, b T, op string) bool {
	switch op {
	case "==":
		return a == b
	case "!=":
		return a != b
	case "<":
		return a < b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case ">=":
		return a >= b
	}
	return false
}

// resolveOperand 解析操作数：引号字面量 / 数字字面量 / 黑板变量。
func resolveOperand(operand string, bb *Blackboard) (any, error) {
	if len(operand) >= 2 {
		if (operand[0] == '"' && operand[len(operand)-1] == '"') ||
			(operand[0] == '\'' && operand[len(operand)-1] == '\'') {
			return operand[1 : len(operand)-1], nil
		}
	}
	if f, err := strconv.ParseFloat(operand, 64); err == nil {
		return f, nil
	}
	if v, ok := bb.Lookup(operand); ok {
		return v, nil
	}
	return nil, fmt.Errorf("未知变量: %s", operand)
}

func toFloat(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func truthy(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case string:
		return val != ""
	default:
		return false
	}
}
