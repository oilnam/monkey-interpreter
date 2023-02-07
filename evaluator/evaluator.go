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
	}
	return nil
}

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
			return nil
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
			return nil
		}
	}
	return nil
}
