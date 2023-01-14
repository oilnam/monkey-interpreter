package ast

import "monkey/token"

type Node interface {
	// TokenLiteral returns the literal value of its token
	// It's only used for debugging and testing
	TokenLiteral() string
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

type LetStatement struct {
	// e.g. `let x = 5 + 5`
	Token token.Token // the token.LET token (let)
	Name  *Identifier // the name of the variable (x)
	Value Expression  // the RHS (5 + 5)
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

type Identifier struct {
	Token token.Token // the token.IDENT token
	value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
