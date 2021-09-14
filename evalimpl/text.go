package evalimpl

import (
	"fmt"
	"strings"
)

type EvaluatorText struct {
	text string
}

func (e *EvaluatorText) Init(s string) {
	e.text = s
}

func (e *EvaluatorText) EvalFloat(v float64) (result bool, err error) {
	return e.EvalString(fmt.Sprintf("%f", v))
}

func (e *EvaluatorText) EvalInt(v int64) (result bool, err error) {
	return e.EvalString(fmt.Sprintf("%d", v))
}

func (e *EvaluatorText) EvalString(v string) (result bool, err error) {
	return strings.Contains(e.text, v), nil
}

func (e *EvaluatorText) Not(v string) (result bool, err error) {
	return !strings.Contains(e.text, v), nil
}
