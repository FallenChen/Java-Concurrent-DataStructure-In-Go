package main

import (
	"fmt"
	"github.com/Knetic/govaluate"
)

func main() {

	expression1, _ := govaluate.NewEvaluableExpression("10 > 0")
	result1, _ := expression1.Evaluate(nil)
	fmt.Println(result1)

	expression2, _ := govaluate.NewEvaluableExpression("foo > 0")
	parameters1 := make(map[string]interface{}, 8)
	parameters1["foo"] = -1
	result2, _ := expression2.Evaluate(parameters1)
	fmt.Println(result2)

	expression3, _ := govaluate.NewEvaluableExpression("(mem_used / total_mem) * 100")
	parameters3 := make(map[string]interface{}, 8)
	parameters3["total_mem"] = 1024
	parameters3["mem_used"] = 512
	result3, _ := expression3.Evaluate(parameters3)
	fmt.Println(result3)

	functions := map[string]govaluate.ExpressionFunction{
		"strlen": func(args ...interface{}) (interface{}, error) {
			length := len(args[0].(string))
			return (float64)(length), nil
		},
	}
	expString := "strlen('someReallyLongInputString') <= 16"
	expression4, _ := govaluate.NewEvaluableExpressionWithFunctions(expString, functions)
	result4, _ := expression4.Evaluate(nil)
	fmt.Println(result4)
}
