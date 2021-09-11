package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/suzuito/suzuitoql"
)

func main() {
	// main1()
	main2()
}

func main1() {
	expr := ""
	flag.StringVar(&expr, "expr", "", "")
	flag.Parse()

	if expr == "" {
		fmt.Fprintf(os.Stderr, "Required expr\n")
		flag.Usage()
		os.Exit(1)
	}

	_, err := suzuitoql.GenerateFilterFromString(expr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func main2() {
	body, err := ioutil.ReadFile("b.txt")
	// body, err := ioutil.ReadFile("c.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	filter, err := suzuitoql.GenerateFilterFromString(string(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	fmt.Printf("====\nResult %+v\n", filter)
}
