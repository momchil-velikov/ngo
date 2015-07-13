package main

import (
	"fmt"
	"golang/scanner"
	"io"
	"io/ioutil"
	"os"
	"unicode"
	"unicode/utf8"
)

func escapeRune(in rune) string {
	return fmt.Sprintf("\\U%08x", in)
}

func escapeByte(in byte) string {
	return fmt.Sprintf("\\x%02x", in)
}

func escape(in string, str bool) string {
	var out []byte
	for off, r := range in {
		if r == '\\' {
			out = append(out, "\\\\"...)
		} else if r == '"' && str {
			out = append(out, "\\\""...)
		} else if r == '\'' && !str {
			out = append(out, "\\'"...)
		} else if r == utf8.RuneError {
			_, sz := utf8.DecodeRuneInString(in[off:])
			if sz == 1 {
				out = append(out, escapeByte(in[off])...)
			} else {
				out = append(out, escapeRune(r)...)
			}
		} else if unicode.IsPrint(r) {
			out = append(out, in[off:off+utf8.RuneLen(r)]...)
		} else {
			out = append(out, escapeRune(r)...)
		}
	}

	return string(out)
}

func transcan(in io.Reader, out io.Writer) {
	src, err := ioutil.ReadAll(in)
	if err != nil {
		fmt.Fprint(os.Stderr, "scanfilt: ", err)
	}

	var tok uint
	s := scanner.New("stdin", string(src))
	for tok = s.Get(); tok != scanner.EOF && tok != scanner.ERROR; tok = s.Get() {
		switch tok {
		case scanner.ID, scanner.INTEGER, scanner.FLOAT:
			fmt.Fprint(out, s.Value, " ")
		case scanner.IMAGINARY:
			fmt.Fprint(out, s.Value, "i ")
		case scanner.STRING:
			fmt.Fprintf(out, "\"%s\" ", escape(s.Value, true))
		case scanner.RUNE:
			fmt.Fprintf(out, "'%s' ", escape(s.Value, false))
		// case ';':
		//     fmt.Fprint(out, "\n")
		default:
			fmt.Fprint(out, scanner.TokenNames[tok], " ")
		}
	}

	if tok == scanner.ERROR {
		fmt.Fprint(os.Stderr, "\nscanfilt: ", s.Err, "\n")
	}
}

func main() {
	transcan(os.Stdin, os.Stdout)
}
