package engine

import (
	"fmt"
	"strconv"
	"strings"
)

// ErrGuardFailed indicates a guard condition evaluated to false.
// Kept as a simple error type; callers may inspect via errors.Is if needed.
type ErrGuardFailed struct {
	Message string
}

func (e *ErrGuardFailed) Error() string { return e.Message }

// getContextValue resolves a dot-path key (e.g. "payment.status") from a flat
// context map. It first checks the literal flat key; if absent, it traverses
// nested maps by splitting on ".".
func getContextValue(ctx map[string]any, key string) (any, bool) {
	// 1. flat key
	if v, ok := ctx[key]; ok {
		return v, true
	}
	// 2. nested traversal
	parts := strings.Split(key, ".")
	var cur any = ctx
	for _, part := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[part]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

// EvaluateGuardGroup evaluates a single guard group (AND/OR) against context.
func EvaluateGuardGroup(group GuardGroup, ctx map[string]any) (bool, error) {
	if len(group.Conditions) == 0 {
		return true, nil
	}
	result := group.Logic == "AND" // start true for AND, false for OR
	for i, cond := range group.Conditions {
		ok, err := evaluateCondition(cond, ctx)
		if err != nil {
			return false, err
		}
		if group.Logic == "AND" {
			if !ok {
				return false, nil
			}
		} else {
			if ok {
				return true, nil
			}
			// OR: keep going; last one decides
			if i == len(group.Conditions)-1 {
				result = ok
			}
		}
	}
	return result, nil
}

// EvaluateGuards evaluates all guard groups; every group must pass (AND across groups).
func EvaluateGuards(guards []GuardGroup, ctx map[string]any) (bool, error) {
	for _, g := range guards {
		ok, err := EvaluateGuardGroup(g, ctx)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, &ErrGuardFailed{Message: "guard group failed"}
		}
	}
	return true, nil
}

func evaluateCondition(cond GuardCondition, ctx map[string]any) (bool, error) {
	switch cond.Operator {
	case OpExists:
		_, ok := getContextValue(ctx, cond.Field)
		return ok, nil
	case OpNotExists:
		_, ok := getContextValue(ctx, cond.Field)
		return !ok, nil
	}

	actual, ok := getContextValue(ctx, cond.Field)
	if !ok {
		// field missing; non-EXISTS operators treat missing as false
		return false, nil
	}

	switch cond.Operator {
	case OpEq:
		return compareEq(actual, cond.Value)
	case OpNeq:
		eq, err := compareEq(actual, cond.Value)
		if err != nil {
			return false, err
		}
		return !eq, nil
	case OpGt, OpGte, OpLt, OpLte:
		return compareOrdered(actual, cond.Value, cond.Operator)
	case OpIn:
		return contains(cond.Value, actual)
	case OpNotIn:
		ok, err := contains(cond.Value, actual)
		if err != nil {
			return false, err
		}
		return !ok, nil
	default:
		return false, fmt.Errorf("unsupported guard operator: %s", cond.Operator)
	}
}

func compareEq(a, b any) (bool, error) {
	// numeric comparison fallback (e.g. string "100" vs number 100)
	if an, ok := toFloat(a); ok {
		if bn, ok2 := toFloat(b); ok2 {
			return an == bn, nil
		}
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b), nil
}

func compareOrdered(a, b any, op GuardOperator) (bool, error) {
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		switch op {
		case OpGt:
			return af > bf, nil
		case OpGte:
			return af >= bf, nil
		case OpLt:
			return af < bf, nil
		case OpLte:
			return af <= bf, nil
		}
	}
	// fallback to string ordering
	as, bs := fmt.Sprintf("%v", a), fmt.Sprintf("%v", b)
	switch op {
	case OpGt:
		return as > bs, nil
	case OpGte:
		return as >= bs, nil
	case OpLt:
		return as < bs, nil
	case OpLte:
		return as <= bs, nil
	}
	return false, nil
}

func contains(container, needle any) (bool, error) {
	switch c := container.(type) {
	case []any:
		for _, item := range c {
			eq, err := compareEq(item, needle)
			if err != nil {
				continue
			}
			if eq {
				return true, nil
			}
		}
		return false, nil
	case []string:
		ns := fmt.Sprintf("%v", needle)
		for _, item := range c {
			if item == ns {
				return true, nil
			}
		}
		return false, nil
	case string:
		// treat as comma/space separated list
		parts := strings.FieldsFunc(c, func(r rune) bool {
			return r == ',' || r == ' '
		})
		ns := fmt.Sprintf("%v", needle)
		for _, p := range parts {
			if p == ns {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unsupported IN container type")
	}
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	}
	return 0, false
}
