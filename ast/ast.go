package ast

import (
	"bytes"
	"fmt"
	"monkey/token"
	"strings"
)

type Node interface {
	// TokenLiteral returns the literal value of its token
	// It's only used for debugging and testing
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	// dummy method just to type check
	statementNode()
}

type Expression interface {
	Node
	// dummy method just to type check
	expressionNode()
}

// Program is the root node of every AST produced
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// IDENTIFIER (expression)
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string      // the name of the variable (x)
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// LET statement
type LetStatement struct {
	// e.g. `let x = 5 + 5`
	Token token.Token // the token.LET token (let)
	Name  *Identifier // the name of the variable (x)
	Value Expression  // the RHS (5 + 5)
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// RETURN statement
type ReturnStatement struct {
	Token       token.Token // the token.RETURN token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// EXPRESSION statement
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// INTEGER LITERAL (expression)
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// BOOLEAN LITERAL (expression)
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

// STRING LITERAL (expression)
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

// PREFIX EXPRESSION
type PrefixExpression struct {
	Token    token.Token
	Operator string     // `-` or `!`
	Right    Expression // the expression to the right of the operator
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	// "(operator, right)"
	return "(" + pe.Operator + pe.Right.String() + ")"
}

// INFIX EXPRESSION
type InfixExpression struct {
	Token    token.Token // +, *, etc
	Left     Expression
	Operator string // +, *, etc
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	// "(left, operator, right)"
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

// REASSIGNMENT EXPRESSION
type ReassignmentExpression struct {
	Token token.Token // =
	Left  *Identifier
	Right Expression
}

func (ie *ReassignmentExpression) expressionNode()      {}
func (ie *ReassignmentExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *ReassignmentExpression) String() string {
	return ie.Left.String() + " = " + ie.Right.String()
}

// IF EXPRESSION
type IfExpression struct {
	Token       token.Token // the `if` token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	s := "if" + ie.Condition.String() + " " + ie.Consequence.String()
	if ie.Alternative != nil {
		s += "else " + ie.Alternative.String()
	}
	return s
}

// WHILE is very similar to IF
type WhileExpression struct {
	Token     token.Token // the `while` token
	Condition Expression
	Body      *BlockStatement
}

func (we *WhileExpression) expressionNode()      {}
func (we *WhileExpression) TokenLiteral() string { return we.Token.Literal }
func (we *WhileExpression) String() string {
	return fmt.Sprintf("while %s { %s }", we.Condition.String(), we.Body.String())
}

// FOR loops, Python style
type ForLoop struct {
	Token    token.Token // the `for` token
	Iterator *Identifier
	Elements []Expression // for array literals (`for i in [1,2,3])
	Ident    Expression   // identifier (`let array = ... ; for i in array`)
	Body     *BlockStatement
}

func (fl *ForLoop) expressionNode()      {}
func (fl *ForLoop) TokenLiteral() string { return fl.Token.Literal }
func (fl *ForLoop) String() string {
	return fmt.Sprintf("for %s in %s { %s }", fl.Iterator.String(), fl.Elements, fl.Body)
}

type BlockStatement struct {
	Token      token.Token // the `{` token
	Statements []Statement
}

// BLOCK STATEMENTS
func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out string
	for _, s := range bs.Statements {
		out += s.String()
	}
	return out
}

// FUNCTION LITERALS
type FunctionLiteral struct {
	Token  token.Token   // the `fn` token
	Params []*Identifier //
	Body   *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fl.Params {
		params = append(params, p.String())
	}
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())
	return out.String()
}

// CALL EXPRESSIONS
type CallExpression struct {
	Token     token.Token // the `(` token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// MAP FUNCTION
type MapFunction struct {
	Token    token.Token  // the `map` token
	Function Expression   // Identifier or FunctionLiteral
	Elements []Expression // same as in ArrayLiteral
}

func (m *MapFunction) expressionNode()      {}
func (m *MapFunction) TokenLiteral() string { return m.Token.Literal }
func (m *MapFunction) String() string       { return "map!" }

// ARRAYS
type ArrayLiteral struct {
	Token    token.Token // the [ token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// INDEX EXPRESSIONS
type IndexExpression struct {
	Token token.Token // the [ token
	Left  Expression  // identifier, array literal, function call...
	Index Expression  // so we can do things like array[1+1], array[$var]
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return "(" + ie.Left.String() + "[" + ie.Index.String() + "])"
}

// HASH TABLES
type HashLiteral struct {
	Token token.Token // the { token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
