package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

// Global objects
var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		if node.Value {
			return TRUE
		} else {
			return FALSE
		}
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		right := Eval(node.Right)
		left := Eval(node.Left)
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalStatements(node.Statements)
	case *ast.IfExpression:
		return evalIfExpression(node)
	}
	return NULL
}

// NOTE: this seems to be evaluating *all* statements
// (so I guess at some point we'll add side effects to it too)
// but only returns the LAST statement, even if we have more than one.
func evalStatements(stmts []ast.Statement) object.Object {
	var result object.Object
	for _, s := range stmts {
		result = Eval(s)
	}
	return result
}

func evalPrefixExpression(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return evalBangOperatorExp(right)
	case "-":
		return evalMinusOperatorExp(right)
	default:
		return NULL
	}
}

func evalBangOperatorExp(exp object.Object) object.Object {
	//fmt.Printf("got exp %v of type %T\n", exp, exp)
	switch exp {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		// anything else, like an int, is `true`, so !anything => false
		return FALSE
	}
}

func evalMinusOperatorExp(exp object.Object) object.Object {
	if exp.Type() != object.INTEGER_OBJ {
		return NULL
	}
	value := exp.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(op string, left, right object.Object) object.Object {

	if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
		l := left.(*object.Boolean)
		r := right.(*object.Boolean)
		switch op {
		case "==":
			return &object.Boolean{Value: l.Value == r.Value}
		case "!=":
			return &object.Boolean{Value: l.Value != r.Value}
		default:
			return NULL
		}
	}

	l, ok1 := left.(*object.Integer)
	r, ok2 := right.(*object.Integer)
	if ok1 && ok2 {
		switch op {
		case "+":
			return &object.Integer{Value: l.Value + r.Value}
		case "-":
			return &object.Integer{Value: l.Value - r.Value}
		case "*":
			return &object.Integer{Value: l.Value * r.Value}
		case "/":
			return &object.Integer{Value: l.Value / r.Value}
		case "<":
			return &object.Boolean{Value: l.Value < r.Value}
		case ">":
			return &object.Boolean{Value: l.Value > r.Value}
		case "==":
			return &object.Boolean{Value: l.Value == r.Value}
		case "!=":
			return &object.Boolean{Value: l.Value != r.Value}
		default:
			return NULL
		}
	}
	return NULL
}

// My own implementation, because the one in the book (see below)
// breaks the tests.
func evalIfExpression(node *ast.IfExpression) object.Object {
	cond := Eval(node.Condition)

	if cond.Type() == object.BOOLEAN_OBJ {
		if cond.(*object.Boolean).Value { // if true
			if node.Consequence != nil {
				return Eval(node.Consequence)
			}
		} else { // bool is false
			if node.Alternative != nil {
				return Eval(node.Alternative)
			}
		}
	}
	if cond.Type() == object.INTEGER_OBJ {
		if node.Consequence != nil {
			return Eval(node.Consequence)
		}
	}
	return NULL
}

// The following is the implementation suggested in the book,
// but for some strange reason it doesn't work, so I kept my own.
//func evalIfExpression(ie *ast.IfExpression) object.Object {
//	condition := Eval(ie.Condition)
//	fmt.Printf("got cond %v of type %T\n", condition, condition)
//	if isTruthy(condition) {
//		return Eval(ie.Consequence)
//	} else if ie.Alternative != nil {
//		return Eval(ie.Alternative)
//	} else {
//		return NULL
//	}
//}
//
//// this always goes to default when evaluating an expression
//// like (1 > 2), instead of matching FALSE, and I have no idea why
//func isTruthy(obj object.Object) bool {
//	switch obj {
//	case NULL:
//		fmt.Println("1")
//		return false
//	case TRUE:
//		fmt.Println("2")
//		return true
//	case FALSE:
//		fmt.Println("3")
//		return false
//	default:
//		fmt.Println("4")
//		return true
//	}
//}
