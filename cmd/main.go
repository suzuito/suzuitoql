package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/suzuito/suzuitoql"
	"github.com/suzuito/suzuitoql/evalimpl"
)

func main() {
	// main2()
	main3()
}

// func main2() {
// 	body, err := ioutil.ReadFile("c.txt")
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "%+v\n", err)
// 		os.Exit(1)
// 	}
// 	filter, err := suzuitoql.GenerateFilterFromString(string(body))
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "%+v\n", err)
// 		os.Exit(1)
// 	}
// 	fmt.Printf("====\n")
// 	result, err := filter.Eval(&EvaluatorImpl{})
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "%+v\n", err)
// 		os.Exit(1)
// 	}
// 	fmt.Println(result)
// }

func main3() {
	all, err := ioutil.ReadFile("data/1.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	rows := strings.Split(string(all), "\n")

	filter, err := suzuitoql.GenerateFilterFromString(`
	("ゴーシュ" && "われわれは下手")
	||
	("ゴーシュ" && Not("ねずみ"))
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	evaluator := evalimpl.EvaluatorText{}
	for _, row := range rows {
		evaluator.Init(row)
		result, err := filter.Eval(&evaluator)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			os.Exit(1)
		}
		if !result {
			continue
		}
		fmt.Printf("> %s\n", row)
	}
}
