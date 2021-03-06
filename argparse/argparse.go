package argparse

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var (
	ErrInvalidSyntax = errors.New("invalid syntax")
)

type Parser interface {
	Parse(string) (interface{}, int, error)
}

type EmptyParser interface {
	Empty() (interface{}, error)
}

func countLeftSpaces(str string) (count int) {
	for _, r := range str {
		if !unicode.IsSpace(r) {
			break
		}
		count++
	}
	return
}

func spaceEndIndex(str string) int {
	if i := strings.IndexFunc(str, unicode.IsSpace); i != -1 {
		return i
	}
	return len(str)
}

type ErrDetails struct {
	Err error
	Str string
	Pos int
}

func (e ErrDetails) Error() string {
	if det, ok := e.Err.(ErrDetails); ok {
		return det.Error()
	}
	msg := e.Err.Error()
	str, pos := shorten(e.Str, e.Pos)
	return fmt.Sprintf("%s\n  in \"%s\"\n%*s^", msg, str, pos+6, " ")
}

var etc = ".." // var for tests

func shorten(str string, pos int) (string, int) {
	const nl = '\n'
	for i, r := range str {
		if r == nl && i < pos {
			i++
			str = etc + str[i:]
			pos = len(etc) - i + pos
			break
		}
	}
	for i, r := range str {
		if r == nl && i >= pos {
			str = str[:i] + etc
			break
		}
	}
	return str, pos
}

func OriginalErr(err error) error {
	for {
		argErr, ok := err.(ErrDetails)
		if !ok {
			break
		}
		err = argErr.Err
	}
	return err
}
