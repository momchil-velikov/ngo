package parser

import "fmt"

type ErrorList []error

func (lst ErrorList) Error() string {
    var s []byte = nil

    for _, e := range lst {
        s = append(s, '\n')
        s = append(s, e.Error()...)
    }

    return string(s)
}

type parse_error struct {
    Name      string
    Line, Pos int
    Msg       string
}

func (e parse_error) Error() string {
    return fmt.Sprintf("parse error: %s:%d:%d: %s", e.Name, e.Line, e.Pos, e.Msg)
}
