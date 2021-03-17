package ecssd

import (
	"fmt"
	"strings"
)

type multiError struct {
	errors []error
}

func newMulti() *multiError {
	return &multiError{}
}

func (m *multiError) Append(err error) {
	if err == nil {
		return
	}
	m.errors = append(m.errors, err)
}

func (m *multiError) Errors() []error {
	return m.errors
}

func (m *multiError) Error() string {
	if len(m.errors) == 0 {
		return "no error. forgot to use ErrorOrNil"
	}
	var msgs []string
	for _, e := range m.errors {
		msgs = append(msgs, e.Error())
	}
	return fmt.Sprintf("%d errors: %s", len(m.errors), strings.Join(msgs, "; "))
}

func (m *multiError) ErrorOrNil() error {
	if len(m.errors) == 0 {
		return nil
	}
	return m
}
