package main

import (
	"fmt"
	"golang/scanner"
	"io"
	"io/ioutil"
	"os"
)

func transcan(in io.Reader, out io.Writer) error {
	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	var tok uint
	s := scanner.New("stdin", string(src))
	for tok = s.Get(); tok != scanner.EOF && tok != scanner.ERROR; tok = s.Get() {
		v := string(s.Value)
		switch tok {
		case scanner.ID, scanner.INTEGER, scanner.FLOAT, scanner.STRING:
			fmt.Fprint(out, v, " ")
		case scanner.IMAGINARY:
			fmt.Fprint(out, v, "i ")
		case scanner.RUNE:
			fmt.Fprintf(out, "'%s' ", v)
		case scanner.LINE_COMMENT:
			fmt.Fprintln(out, v)
		case scanner.BLOCK_COMMENT:
			fmt.Fprint(out, v)
		default:
			fmt.Fprint(out, scanner.TokenNames[tok], " ")
		}
	}
	if tok == scanner.ERROR {
		return s.Err
	}
	return nil
}

func main() {
	if err := transcan(os.Stdin, os.Stdout); err == nil {
		os.Exit(0)
	} else {
		fmt.Fprint(os.Stderr, "scanfilt: ", err)
		os.Exit(1)
	}
}
