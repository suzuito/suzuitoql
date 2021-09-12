package suzuitoql

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/xerrors"
)

var newlineRegexp = regexp.MustCompile(`\r?\n`)

func GenerateFilterFromString(expr string) (*Filter, error) {
	norm := newlineRegexp.ReplaceAllString(expr, "")
	b, err := format.Source([]byte(norm))
	if err != nil {
		return nil, err
	}
	root, err := parser.ParseExpr(string(b))
	if err != nil {
		return nil, xerrors.Errorf("Cannot ParseExpr %s : %w", expr, err)
	}
	return GenerateFilter(b, root)
}

func GenerateFilter(source []byte, root ast.Expr) (*Filter, error) {
	visitor := visitorExpression2{
		Err:   nil,
		nodes: []ast.Node{},
		Stack: []ast.Node{},
	}
	ast.Walk(&visitor, root)
	elems, err := newElements(source, visitor.Stack)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	// fmt.Println("====")
	// for _, elem := range *elems {
	// fmt.Printf("%s\n", elem.String())
	// }
	return &Filter{
		elems: elems,
	}, visitor.Err
}

type visitorExpression2 struct {
	Err    error
	nodes  []ast.Node
	Stack  []ast.Node
	Source []byte
}

func (v *visitorExpression2) Visit(current ast.Node) ast.Visitor {
	if current != nil {
		// Process after entered
		v.nodes = append(v.nodes, current)
		if err := v.onEnter(current); err != nil {
			v.Err = err
			return nil
		}
		return v
	}
	// Process before exited
	current = v.nodes[len(v.nodes)-1]
	v.nodes = v.nodes[:len(v.nodes)-1]
	if err := v.onExit(current); err != nil {
		v.Err = err
		return nil
	}
	return nil
}

func (v *visitorExpression2) onEnter(current ast.Node) error {
	return nil
}

func (v *visitorExpression2) onExit(current ast.Node) error {
	switch n := current.(type) {
	case *ast.BinaryExpr:
		if n.Op != token.LAND && n.Op != token.LOR {
			return xerrors.Errorf("Unsupported BinaryExpr: %s", n.Op)
		}
		v.Stack = append(v.Stack, current)
		// fmt.Printf("BinaryExpr %+v\n", n)
	case *ast.BasicLit:
		// fmt.Printf("Lit : %s %s\n", n.Kind, n.Value)
		if n.Kind != token.STRING && n.Kind != token.INT && n.Kind != token.FLOAT {
			return xerrors.Errorf("Unsupported BasecLit : %s %s", n.Kind, n.Value)
		}
		v.Stack = append(v.Stack, current)
		// fmt.Printf("BasicLit %+v\n", n)
	case *ast.UnaryExpr:
		if n.Op != token.SUB {
			return xerrors.Errorf("Unsupported UnaryExpr : %s %s", n.Op)
		}
		v.Stack = append(v.Stack, current)
		// fmt.Printf("UnaryExpr %+v\n", n)
	case *ast.CallExpr:
		// fmt.Printf("Call: %s %v\n", n.Fun, n.Args)
		v.Stack = append(v.Stack, current)
		// fmt.Printf("CallExpr %+v\n", n)
	case *ast.Ident:
		if n.String() == "true" || n.String() == "false" {
			v.Stack = append(v.Stack, current)
		}
		// fmt.Printf("Ident %+v\n", n)
	default:
		// fmt.Printf("%s %+v\n", reflect.TypeOf(n), n)
	}
	return nil
}

type elementType string

const (
	elementTypeOpBinAnd  elementType = "and"
	elementTypeOpBinOr   elementType = "or"
	elementTypeOpMinus   elementType = "-"
	elementTypeOpFunc    elementType = "func"
	elementTypeLitString elementType = "string"
	elementTypeLitInt    elementType = "int"
	elementTypeLitFloat  elementType = "float"
	elementTypeLitBool   elementType = "bool"
)

type element struct {
	Type        elementType
	FuncName    string
	FuncArgs    int
	ValueString string
	ValueInt    int64
	ValueFloat  float64
	ValueBool   bool
}

func (e *element) String() string {
	switch e.Type {
	case elementTypeOpBinAnd:
		return string(e.Type)
	case elementTypeOpBinOr:
		return string(e.Type)
	case elementTypeOpMinus:
		return string(e.Type)
	case elementTypeOpFunc:
		return fmt.Sprintf("%s(%d)", e.FuncName, e.FuncArgs)
	case elementTypeLitString:
		return e.ValueString
	case elementTypeLitInt:
		return fmt.Sprintf("%d", e.ValueInt)
	case elementTypeLitFloat:
		return fmt.Sprintf("%f", e.ValueFloat)
	case elementTypeLitBool:
		return fmt.Sprintf("%v", e.ValueBool)
	}
	return fmt.Sprintf("%+v", *e)
}

func (e *element) Value() (reflect.Value, error) {
	var v reflect.Value
	switch e.Type {
	case elementTypeLitFloat:
		v = reflect.ValueOf(e.ValueFloat)
	case elementTypeLitBool:
		v = reflect.ValueOf(e.ValueBool)
	case elementTypeLitInt:
		v = reflect.ValueOf(e.ValueInt)
	case elementTypeLitString:
		v = reflect.ValueOf(e.ValueString)
	default:
		return v, xerrors.Errorf("Cannot new value of type '%s'", e.Type)
	}
	return v, nil
}

type elements []element

func (e *elements) String() string {
	r := []string{}
	for _, v := range *e {
		r = append(r, v.String())
	}
	return strings.Join(r, ",")
}

func newElements(source []byte, nodes []ast.Node) (*elements, error) {
	r := elements{}
	for _, node := range nodes {
		e, err := newElement(source, node)
		if err != nil {
			return nil, xerrors.Errorf(": %w", err)
		}
		r = append(r, *e)
	}
	return &r, nil
}

func newElementByValue(v reflect.Value) (*element, error) {
	switch v.Type().Name() {
	case "string":
		return &element{
			Type:        elementTypeLitString,
			ValueString: v.String(),
		}, nil
	case "int":
		return &element{
			Type:     elementTypeLitInt,
			ValueInt: v.Int(),
		}, nil
	case "int64":
		return &element{
			Type:     elementTypeLitInt,
			ValueInt: v.Int(),
		}, nil
	case "float":
		return &element{
			Type:       elementTypeLitFloat,
			ValueFloat: v.Float(),
		}, nil
	case "float64":
		return &element{
			Type:       elementTypeLitFloat,
			ValueFloat: v.Float(),
		}, nil
	case "bool":
		return &element{
			Type:      elementTypeLitBool,
			ValueBool: v.Bool(),
		}, nil
	}
	return nil, xerrors.Errorf("Cannot element from Value(%s)", v.Type().Name())
}

func newElement(source []byte, node ast.Node) (*element, error) {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		if n.Op == token.LAND {
			return &element{
				Type: elementTypeOpBinAnd,
			}, nil
		}
		if n.Op == token.LOR {
			return &element{
				Type: elementTypeOpBinOr,
			}, nil
		}
		return nil, xerrors.Errorf("Unsupported BinaryExpr: %s", n.Op)
	case *ast.BasicLit:
		if n.Kind == token.STRING {
			return &element{
				Type:        elementTypeLitString,
				ValueString: strings.TrimSuffix(strings.TrimPrefix(n.Value, `"`), `"`),
			}, nil
		}
		if n.Kind == token.INT {
			v, err := strconv.ParseInt(n.Value, 10, 64)
			if err != nil {
				return nil, xerrors.Errorf("Cannot convert str to int64 : %w", err)
			}
			return &element{
				Type:     elementTypeLitInt,
				ValueInt: v,
			}, nil
		}
		if n.Kind == token.FLOAT {
			v, err := strconv.ParseFloat(n.Value, 64)
			if err != nil {
				return nil, xerrors.Errorf("Cannot convert str to float64 : %w", err)
			}
			return &element{
				Type:       elementTypeLitFloat,
				ValueFloat: v,
			}, nil
		}
		return nil, xerrors.Errorf("Unsupported BasecLit : %s %s", n.Kind, n.Value)
	case *ast.UnaryExpr:
		if n.Op == token.SUB {
			return &element{
				Type: elementTypeOpMinus,
			}, nil
		}
		return nil, xerrors.Errorf("Unsupported UnaryExpr : %s %s", n.Op)
	case *ast.CallExpr:
		return &element{
			Type:     elementTypeOpFunc,
			FuncName: string(source[n.Fun.Pos()-1 : n.Fun.End()-1]),
			FuncArgs: len(n.Args),
		}, nil
	case *ast.Ident:
		if n.String() == "true" {
			return &element{
				Type:      elementTypeLitBool,
				ValueBool: true,
			}, nil
		}
		if n.String() == "false" {
			return &element{
				Type:      elementTypeLitBool,
				ValueBool: false,
			}, nil
		}
		return nil, xerrors.Errorf("Unsupported UnaryExpr : %+v", n)
	}
	return nil, xerrors.Errorf("Unsupported %s : %+v", reflect.TypeOf(node), node)
}

type Filter struct {
	elems *elements
}

func (f *Filter) Eval(
	evaluator Evaluator,
) (bool, error) {
	stack := elements{}
	// fmt.Println("Start")
	for i := range *f.elems {
		elem := (*f.elems)[i]
		// fmt.Println("> ----")
		// aaa := (*f.elems)[i:]
		// fmt.Println(aaa.String())
		// fmt.Println(stack.String())
		switch elem.Type {
		case elementTypeLitString:
			stack = append(stack, elem)
		case elementTypeLitInt:
			stack = append(stack, elem)
		case elementTypeLitFloat:
			stack = append(stack, elem)
		case elementTypeLitBool:
			stack = append(stack, elem)
		case elementTypeOpBinAnd, elementTypeOpBinOr:
			if len(stack) < 2 {
				return false, xerrors.Errorf("Stack must be larger than 2 for %s op", elem.Type)
			}
			args := elements{
				stack[len(stack)-2],
				stack[len(stack)-1],
			}
			stack = stack[:len(stack)-2]
			// var err error
			var result *element
			switch elem.Type {
			case elementTypeOpBinAnd:
				bresult, err := evalAnd(&args[0], &args[1], evaluator)
				if err != nil {
					return false, xerrors.Errorf(": %w", err)
				}
				result = &element{
					Type:      elementTypeLitBool,
					ValueBool: bresult,
				}
			case elementTypeOpBinOr:
				bresult, err := evalOr(&args[0], &args[1], evaluator)
				if err != nil {
					return false, xerrors.Errorf(": %w", err)
				}
				result = &element{
					Type:      elementTypeLitBool,
					ValueBool: bresult,
				}
			default:
				return false, xerrors.Errorf("Unsupport op %+v", elem.Type)
			}
			stack = append(stack, *result)
		case elementTypeOpMinus:
			if len(stack) < 1 {
				return false, xerrors.Errorf("Stack must be larger than 1 for %s op", elem.Type)
			}
			args := elements{
				stack[len(stack)-1],
			}
			stack = stack[:len(stack)-1]
			if args[0].Type != elementTypeLitInt && args[0].Type != elementTypeLitFloat {
				return false, xerrors.Errorf("Cannot apply minus for %+v", args[0])
			}
			stack = append(stack, element{
				Type:       args[0].Type,
				ValueInt:   -args[0].ValueInt,
				ValueFloat: -args[0].ValueFloat,
			})
		case elementTypeOpFunc:
			if len(stack) < elem.FuncArgs {
				return false, xerrors.Errorf("Stack must be larger than %d for function", elem.FuncArgs)
			}
			args := elements{}
			for i := 0; i < elem.FuncArgs; i++ {
				args = append(args, stack[len(stack)-elem.FuncArgs+i])
			}
			stack = stack[:len(stack)-elem.FuncArgs]
			result, err := evalFunc(evaluator, elem.FuncName, args...)
			if err != nil {
				return false, xerrors.Errorf(": %w", err)
			}
			stack = append(stack, *result)
		}
	}
	return stack[0].ValueBool, nil
}

func evalAnd(
	a *element,
	b *element,
	evaluator Evaluator,
) (bool, error) {
	aResult, err := evalElement(a, evaluator)
	if err != nil {
		return false, xerrors.Errorf("Cannot evalElement : %w", err)
	}
	bResult, err := evalElement(b, evaluator)
	if err != nil {
		return false, xerrors.Errorf("Cannot evalElement : %w", err)
	}
	return aResult && bResult, nil
}

func evalOr(
	a *element,
	b *element,
	evaluator Evaluator,
) (bool, error) {
	aResult, err := evalElement(a, evaluator)
	if err != nil {
		return false, xerrors.Errorf("Cannot evalElement : %w", err)
	}
	bResult, err := evalElement(b, evaluator)
	if err != nil {
		return false, xerrors.Errorf("Cannot evalElement : %w", err)
	}
	return aResult || bResult, nil
}

func evalElement(
	v *element,
	evaluator Evaluator,
) (bool, error) {
	switch v.Type {
	case elementTypeLitBool:
		return v.ValueBool, nil
	case elementTypeLitFloat:
		return evaluator.EvalFloat(v.ValueFloat)
	case elementTypeLitInt:
		return evaluator.EvalInt(v.ValueInt)
	case elementTypeLitString:
		return evaluator.EvalString(v.ValueString)
	}
	return false, xerrors.Errorf("Cannot eval %s", v.Type)
}

func evalFunc(evaluator Evaluator, funcName string, args ...element) (result *element, err error) {
	et := reflect.TypeOf(evaluator)
	method, exists := et.MethodByName(funcName)
	if !exists {
		return nil, xerrors.Errorf("Method is not found '%s'", funcName)
	}
	values := []reflect.Value{
		reflect.ValueOf(evaluator),
	}
	for i, arg := range args {
		v, err := arg.Value()
		if err != nil {
			return nil, xerrors.Errorf(
				"Arg %d of function %s is not value",
				i,
				funcName,
			)
		}
		// fmt.Printf("%s\n", v)
		values = append(values, v)
	}
	// fmt.Printf("%+v\n", method)
	results := method.Func.Call(values)
	if len(results) != 2 {
		return nil, xerrors.Errorf(
			"Number of function %s's returned value must be 2 : %d",
			len(results),
		)
	}
	valueResult := results[0]
	valueErr := results[1]
	if !valueErr.IsNil() {
		// FIXME
		return nil, xerrors.Errorf("Not impl")
	}
	return newElementByValue(valueResult)
}

type Evaluator interface {
	EvalFloat(v float64) (result bool, err error)
	EvalInt(v int64) (result bool, err error)
	EvalString(v string) (result bool, err error)
}
