package cli

import (
	"bufio"
	"io"
	"strings"
)

type Scanner struct {
	scanner *bufio.Scanner
}

func NewScanner(r io.Reader) *Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	return &Scanner{scanner: scanner}
}

func (s *Scanner) Scan() (string, error) {
	if s.scanner.Scan() {
		return s.scanner.Text(), nil
	}
	if err := s.scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

func ParseCommandLine(line string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := ' '

	for _, ch := range line {
		if !inQuote && (ch == '"' || ch == '\'') {
			inQuote = true
			quoteChar = ch
		} else if inQuote && ch == quoteChar {
			inQuote = false
		} else if !inQuote && ch == ' ' {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func (s *Scanner) Autocomplete(word string, commands []string) []string {
	var suggestions []string
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, word) {
			suggestions = append(suggestions, cmd)
		}
	}
	return suggestions
}
