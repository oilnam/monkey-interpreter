package parser

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"monkey/ast"
	"monkey/lexer"
	"testing"
)

//func TestLetStatements(t *testing.T) {
//	input := `
//		let x = 5;
//		let y = 10;
//		let foobar = 838383;`
//
//	l := lexer.New(input)
//	p := New(l)
//
//	program := p.ParseProgram()
//	checkParserErrors(t, p)
//
//	if program == nil {
//		t.Fatalf("parse program returned nil")
//	}
//
//	assert.Len(t, program.Statements, 3)
//
//	tests := []struct {
//		expectedIdentifier string
//	}{
//		{"x"},
//		{"y"},
//		{"foobar"},
//	}
//
//	for i, tt := range tests {
//		stmt := program.Statements[i]
//		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
//			return
//		}
//	}
//}

func TestNewLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Len(t, program.Statements, 1)

		stmt := program.Statements[0]
		// test `let x`
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
		// test the value
		val := stmt.(*ast.LetStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func TestNewReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5", 5},
		{"return true", true},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Len(t, program.Statements, 1)

		returnStmt, ok := program.Statements[0].(*ast.ReturnStatement)
		assert.True(t, ok)
		assert.Equal(t, "return", returnStmt.TokenLiteral())

		// test the value
		if !testLiteralExpression(t, returnStmt.ReturnValue, tt.expectedValue) {
			return
		}
	}
}

//func TestOldReturnStatements(t *testing.T) {
//	input := `
//	return 5;
//	return 10;
//	return 993322;`
//
//	l := lexer.New(input)
//	p := New(l)
//
//	program := p.ParseProgram()
//	checkParserErrors(t, p)
//
//	assert.Len(t, program.Statements, 3)
//
//	for _, stmt := range program.Statements {
//		returnStmt, ok := stmt.(*ast.ReturnStatement)
//		if !ok {
//			t.Errorf("stmt not *ast.returnStatement. got=%T", stmt)
//			continue
//		}
//		if returnStmt.TokenLiteral() != "return" {
//			t.Errorf("returnStmt.TokenLiteral not 'return', got %q",
//				returnStmt.TokenLiteral())
//		}
//	}
//}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("s.Name not '%s'. got=%s", name, letStmt.Name)
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Len(t, program.Statements, 1)

	// the first (and only) statement is an ExpressionStatement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	// the expression of the statement is an Identifier (foobar)
	ident, ok := stmt.Expression.(*ast.Identifier)
	assert.True(t, ok)

	assert.Equal(t, "foobar", ident.Value)
	assert.Equal(t, "foobar", ident.TokenLiteral())
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Len(t, program.Statements, 1)

	// the first (and only) statement is an ExpressionStatement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	// the expression of the statement is an IntegerLiteral
	ident, ok := stmt.Expression.(*ast.IntegerLiteral)
	assert.True(t, ok)
	assert.Equal(t, int64(5), ident.Value)
	assert.Equal(t, "5", ident.TokenLiteral())
}

func TestBooleanExpression(t *testing.T) {
	input := "true;"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Len(t, program.Statements, 1)

	// the first (and only) statement is an ExpressionStatement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	// the expression of the statement is an IntegerLiteral
	ident, ok := stmt.Expression.(*ast.Boolean)
	assert.True(t, ok)
	assert.Equal(t, true, ident.Value)
	assert.Equal(t, "true", ident.TokenLiteral())
}

func TestIfExpression(t *testing.T) {
	input := "if (x < y) { x }"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Len(t, program.Statements, 1)

	// the first (and only) statement is an ExpressionStatement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	// the expression of the statement is an IfExpression
	exp, ok := stmt.Expression.(*ast.IfExpression)
	assert.True(t, ok)

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	// got 1 conseq (x) of type ExpressionSt
	assert.Len(t, exp.Consequence.Statements, 1)
	cons, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	if !testIdentifier(t, cons.Expression, "x") {
		return
	}

	assert.Nil(t, exp.Alternative)
}

func TestIfElseExpression(t *testing.T) {
	input := "if (x < y) { x } else { y }"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Len(t, program.Statements, 1)

	// the first (and only) statement is an ExpressionStatement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	// the expression of the statement is an IfExpression
	exp, ok := stmt.Expression.(*ast.IfExpression)
	assert.True(t, ok)

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	// got 1 conseq (x) of type ExpressionSt
	assert.Len(t, exp.Consequence.Statements, 1)
	cons, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	if !testIdentifier(t, cons.Expression, "x") {
		return
	}

	assert.Len(t, exp.Alternative.Statements, 1)
	alt, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	if !testIdentifier(t, alt.Expression, "y") {
		return
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `fn(x, y) { x + y; }`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Len(t, program.Statements, 1)
	// the first (and only) statement is an ExpressionStatement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	// the expression of the statement is a function literal
	fn, ok := stmt.Expression.(*ast.FunctionLiteral)
	assert.True(t, ok)

	// got 2 params, x and y
	assert.Len(t, fn.Params, 2)
	testLiteralExpression(t, fn.Params[0], "x")
	testLiteralExpression(t, fn.Params[1], "y")

	// and one body
	assert.Len(t, fn.Body.Statements, 1)

	// the body stmt in an Expression
	bodyStmt, ok := fn.Body.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	// test the expression
	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestCallExpressionParsing(t *testing.T) {
	input := `add(1, 2 * 3, 4 + 5);`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assert.Len(t, program.Statements, 1)
	// the first (and only) statement is an ExpressionStatement
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	// the expression of the statement is a call expression
	exp, ok := stmt.Expression.(*ast.CallExpression)
	assert.True(t, ok)

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	assert.Len(t, exp.Arguments, 3)
	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		function := stmt.Expression.(*ast.FunctionLiteral)
		if len(function.Params) != len(tt.expectedParams) {
			t.Errorf("length parameters wrong. want %d, got=%d\n",
				len(tt.expectedParams), len(function.Params))
		}
		for i, ident := range tt.expectedParams {
			testLiteralExpression(t, function.Params[i], ident)
		}
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Len(t, program.Statements, 1)
		// the first (and only) statement is an ExpressionStatement
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok)

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		assert.True(t, ok)

		assert.Equal(t, tt.operator, exp.Operator)
		if !testLiteralExpression(t, exp.Right, tt.value) {
			return
		}
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}
	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value,
			integ.TokenLiteral())
		return false
	}
	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}
	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}
	if bo.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("bo.TokenLiteral not %t. got=%s",
			value, bo.TokenLiteral())
		return false
	}
	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}
	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value,
			ident.TokenLiteral())
		return false
	}
	return true
}

func testLiteralExpression(t *testing.T,
	exp ast.Expression,
	expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.OperatorExpression. got=%T(%s)", exp, exp)
		return false
	}
	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}
	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}
	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}
	return true
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}
	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		assert.Len(t, program.Statements, 1)
		// the first (and only) statement is an ExpressionStatement
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok)

		// the Expression on the ExpStatement is an InfixExpression
		exp, ok := stmt.Expression.(*ast.InfixExpression)
		assert.True(t, ok)

		// test the left side
		if !testLiteralExpression(t, exp.Left, tt.leftValue) {
			return
		}

		// test the operator
		assert.Equal(t, tt.operator, exp.Operator)

		// test the right side
		if !testLiteralExpression(t, exp.Right, tt.rightValue) {
			return
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"true",
			"true",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestLalala(t *testing.T) {
	input := `-1 + 2;`
	l := lexer.New(input)
	p := New(l)
	_ = p.ParseProgram()
	checkParserErrors(t, p)
}
