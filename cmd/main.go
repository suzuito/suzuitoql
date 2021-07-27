package main

import (
	"flag"
	"fmt"
	"go/parser"
	"os"
)

func main() {
	expr := ""
	flag.StringVar(&expr, "expr", "", "")
	flag.Parse()

	if expr == "" {
		fmt.Fprintf(os.Stderr, "Required expr\n")
		flag.Usage()
		os.Exit(1)
	}

	root, err := parser.ParseExpr(expr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot ParseExpr %s : %+v\n", expr, err)
	}
	fmt.Println(root)
}
