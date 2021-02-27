package runtime

import (
	"fmt"
	"math"
)

// OpAdd implements the '+' function. It tries to determine automatically the
// type based on the first argument.
func OpAdd(values ...interface{}) interface{} {
	if len(values) < 1 {
		panic("Function '+' should take at least one argument")
	}
	var (
		totalInt   int64
		totalFloat float64
	)
	for i, v := range values {
		switch values[i].(type) {
		case int64:
			totalInt += v.(int64)
		case float64:
			totalFloat += v.(float64)
		default:
			panic(fmt.Sprintf("Unable to add values of type %T", values[0]))
		}
	}
	if totalFloat > 0 {
		return totalFloat + float64(totalInt)
	}
	return totalInt
}

// OpSub implements the '-' function. It tries to determine automatically the
// type based on the first argument.
func OpSub(values ...interface{}) interface{} {
	if len(values) < 2 {
		panic("Function '-' should take at least two argument")
	}

	switch values[0].(type) {
	case int64:
		total := values[0].(int64)
		for _, v := range values[1:] {
			total -= v.(int64)
		}
		return total
	case float64:
		total := values[0].(float64)
		for _, v := range values[1:] {
			total -= v.(float64)
		}
		return total
	default:
		panic(fmt.Sprintf("Unable to sub values of type %T", values[0]))
	}
}

// OpMul implements the '*' function. It tries to determine automatically the
// type based on the first argument.
func OpMul(values ...interface{}) interface{} {
	if len(values) < 1 {
		panic("Function '*' should take at least one argument")
	}

	switch values[0].(type) {
	case int64:
		var total int64 = 1
		for _, v := range values {
			total *= v.(int64)
		}
		return total
	case float64:
		var total = 1.0
		for _, v := range values {
			total *= v.(float64)
		}
		return total
	default:
		panic(fmt.Sprintf("Unable to add values of type %T", values[0]))
	}
}

// OpPow implements exponentiation '**' function.
// returns x**y, the base-x exponential of y
// TODO: https://github.com/rumlang/rum/issues/139
func OpPow(values ...float64) float64 {
	y := float64(1)
	if len(values) < 1 {
		panic("Function '**' should take two argument")
	} else if len(values) == 2 {
		y = values[1]
	} else if len(values) > 2 {
		// support the list of numbers, we separate it into groups of two,
		// where the second set is processed with the result of the previous
		// one (thus secessively).
		var pow float64
		for i, value := range values {
			if i == 0 {
				pow = math.Pow(value, 1)
			} else {
				pow = math.Pow(pow, value)
			}
		}
		return pow
	}
	return math.Pow(values[0], y)
}

// OpEqual implements the == comparaison operator. It can work on more than 2
// arguments - it will return true only if they have all the same value.
func OpEqual(values ...interface{}) interface{} {
	if len(values) < 2 {
		panic("Function '==' should take at least two arguments")
	}

	switch values[0].(type) {
	case int64:
		ref := values[0].(int64)
		for _, v := range values[1:] {
			if ref != v.(int64) {
				return false
			}
		}
		return true
	case float64:
		ref := values[0].(float64)
		for _, v := range values[1:] {
			if ref != v.(float64) {
				return false
			}
		}
		return true
	case string:
		ref := values[0].(string)
		for _, v := range values[1:] {
			if ref != v.(string) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf("Unable to compare values of type %T", values[0]))
	}
}

// OpNotEqual implements the != comparaison operator.
func OpNotEqual(values ...interface{}) interface{} {
	if len(values) != 2 {
		panic("Function '!=' should take exactly two arguments")
	}
	return !OpEqual(values...).(bool)
}

// OpLess implements the < comparaison operator.
func OpLess(values ...interface{}) interface{} {
	if len(values) != 2 {
		panic("Comparaison function should take two arguments")
	}

	switch values[0].(type) {
	case int64:
		ref := values[0].(int64)
		for _, v := range values[1:] {
			if ref >= v.(int64) {
				return false
			}
			ref = v.(int64)
		}
		return true
	case float64:
		ref := values[0].(float64)
		for _, v := range values[1:] {
			if ref >= v.(float64) {
				return false
			}
			ref = v.(float64)
		}
		return true
	default:
		panic(fmt.Sprintf("Unable to compare values of type %T", values[0]))
	}
}

// OpLessEqual implements the <= comparaison operator.
func OpLessEqual(values ...interface{}) interface{} {
	return OpLess(values...).(bool) || OpEqual(values...).(bool)
}

// OpGreater implements the > comparaison operator.
func OpGreater(values ...interface{}) interface{} {
	return !OpLessEqual(values...).(bool)
}

// OpGreaterEqual implements the >= comparaison operator.
func OpGreaterEqual(values ...interface{}) interface{} {
	return !OpLess(values...).(bool)
}
