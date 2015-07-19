package main

import (
	"bufio"
	"fmt"
	"golang/parser"
	"io/ioutil"
	"os"
)

func main() {
	in, err := ioutil.ReadAll(bufio.NewReader(os.Stdin))
	if err != nil {
		fmt.Fprintln(os.Stderr, "parsefilt:", err)
	}
	f, err := parser.Parse("stdin", string(in))
	if err != nil {
		fmt.Fprintln(os.Stderr, "parsefilt:", err)
	}
	out := f.Format()
	w := bufio.NewWriter(os.Stdout)
	w.WriteString(out)
	w.Flush()
}
