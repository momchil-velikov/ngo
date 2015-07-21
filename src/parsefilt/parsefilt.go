package main

import (
	"fmt"
	"golang/ast"
	"golang/parser"
	"io/ioutil"
	"os"
)

func main() {
	if in, err := ioutil.ReadAll(os.Stdin); err == nil {
		if f, err := parser.Parse("stdin", string(in)); err == nil {
			ctx := new(ast.FormatContext).Init()
			f.Format(ctx)
			ctx.Flush(os.Stdout)
		} else {
			fmt.Fprintln(os.Stderr, "parsefilt:", err)
		}
	} else {
		fmt.Fprintln(os.Stderr, "parsefilt:", err)
	}
}
