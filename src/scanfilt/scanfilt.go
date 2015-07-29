package main

import (
	"bytes"
	"fmt"
	"golang/scanner"
	"io"
	"io/ioutil"
	"os"
	"unicode/utf8"
)

func sanitizeLineComment(s string) []byte {
	buf := bytes.Buffer{}
	n := len(s)
	i, j := 0, 0
	for j < n {
		r, sz := utf8.DecodeRuneInString(s[j:])
		if r != '*' || j+1 >= n {
			j += sz
			continue
		}
		r, sz = utf8.DecodeRuneInString(s[j+1:])
		if r == '/' {
			buf.WriteString(s[i : j+1])
			buf.WriteByte(' ')
			buf.WriteByte('/')
			i = j + 2
			j = i
		} else {
			j += sz
		}
	}
	if i < j {
		buf.WriteString(s[i:j])
	}
	return buf.Bytes()
}

func sanitizeBlockComment(s string) []byte {
	buf := bytes.Buffer{}
	n := len(s)
	i, j := 0, 0
	for j < n {
		r, sz := utf8.DecodeRuneInString(s[j:])
		if r == '\n' {
			buf.WriteString(s[i:j])
			buf.WriteByte(' ')
			j += sz
			i = j
		} else {
			j += sz
		}
	}
	if i < j {
		buf.WriteString(s[i:j])
	}
	return buf.Bytes()
}

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
			if len(v) < 2 {
				panic("invalid comment token from scanner")
			}
			io.WriteString(out, "/*")
			out.Write(sanitizeLineComment(v[2:]))
			io.WriteString(out, "*/")
		case scanner.BLOCK_COMMENT:
			out.Write(sanitizeBlockComment(v))
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
