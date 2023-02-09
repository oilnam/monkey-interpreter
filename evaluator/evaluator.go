package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
)

// Global objects
var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program: // THIS is the entry point for a program
		return evalProgram(node, env)
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val) // bind the variable name to its val
	// Expressions
	case *ast.Identifier:
		return evalIdentifier(node, env) // eval identifier (a variable)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		if node.Value {
			return TRUE
		} else {
			return FALSE
		}
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	}
	return NULL
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, s := range program.Statements {
		result = Eval(s, env)
		// we unwrap and return the first Return we find
		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
		// we also immediately return errors
		if result.Type() == object.ERROR_OBJ {
			return result
		}
	}
	return result
}

// here every call to evalBlockSt returns the moment it finds a
// Return OR an Error, so that the first one is always returned
// since every call to evalBlockSt always returns
func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, s := range block.Statements {
		result = Eval(s, env)
		if result != nil &&
			(result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ) {
			return result
		}
	}
	return result
}

// old implementation: this doesn't work bc
// say we have `if (true) { return 10 } return 1`
// evalStatements has 2 statements:
// 1) if true return 10
// 2) return 1
// but evaluating (1) calls evalStatement again!
// so in this inner loop result is set to 10
// but then evaluating (2) (outer loop) sets result to 1!
//
// However, if I have `9; return 1; 9` the statements are all
// at the same level, and we return immediately at 1
//
//func evalStatements(stmts []ast.Statement) object.Object {
//	fmt.Println("(call)")
//	for i, s := range stmts {
//		fmt.Printf("%d -> %s\n", i, s.String())
//	}
//
//	var result object.Object
//	for _, s := range stmts {
//		fmt.Println("	evaluating: ", s.String())
//		result = Eval(s)
//		if returnValue, ok := result.(*object.ReturnValue); ok {
//			return returnValue.Value
//		}
//	}
//	return result
//}

func evalPrefixExpression(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return evalBangOperatorExp(right)
	case "-":
		return evalMinusOperatorExp(right)
	default:
		return newError("unknown operator: %s%s", op, right.Type())
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
		return newError("unknown operator: -%s", exp.Type())
	}
	value := exp.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(op string, left, right object.Object) object.Object {
	// both sides of an infix exp must be of the same type
	if left.Type() != right.Type() {
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	}

	// handle bools
	if left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ {
		l := left.(*object.Boolean)
		r := right.(*object.Boolean)
		switch op {
		case "==":
			return &object.Boolean{Value: l.Value == r.Value}
		case "!=":
			return &object.Boolean{Value: l.Value != r.Value}
		default:
			return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
		}
	}

	// handle integers
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
			return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())

		}
	}
	// everything else: type not supported
	return newError("unsupported type: %s", left.Type())
}

// My own implementation, because the one in the book (see below)
// breaks the tests.
func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object {
	cond := Eval(node.Condition, env)
	if isError(cond) {
		return cond
	}
	if cond.Type() == object.BOOLEAN_OBJ {
		if cond.(*object.Boolean).Value { // if true
			if node.Consequence != nil {
				return Eval(node.Consequence, env)
			}
		} else { // bool is false
			if node.Alternative != nil {
				return Eval(node.Alternative, env)
			}
		}
	}
	if cond.Type() == object.INTEGER_OBJ {
		if node.Consequence != nil {
			return Eval(node.Consequence, env)
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

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value) // get the obj associated to this identifier
	if !ok {
		return newError("identifier not found: " + node.Value)
	}
	return val
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
