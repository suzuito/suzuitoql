package filterimpl

import (
	"fmt"
	"strings"
)

type FilterText struct {
	Text string
}

func (e *FilterText) EvalFloat(v float64) (result bool, err error) {
	return e.EvalString(fmt.Sprintf("%f", v))
}

func (e *FilterText) EvalInt(v int64) (result bool, err error) {
	return e.EvalString(fmt.Sprintf("%d", v))
}

func (e *FilterText) EvalString(v string) (result bool, err error) {
	return strings.Contains(e.Text, v), nil
}

func (e *FilterText) Not(v string) (result bool, err error) {
	return !strings.Contains(e.Text, v), nil
}
