package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

type Parser struct {
	l              *lexer.Lexer
	curToken       token.Token
	peekToken      token.Token
	errors         []string
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	// register PREFIX parse functions
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseInteger)
	p.registerPrefix(token.STRING, p.parseString)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)
	p.registerPrefix(token.MAP, p.parseMapFunction)
	p.registerPrefix(token.WHILE, p.parseWhileExpression)
	p.registerPrefix(token.FOR, p.parseForLoop)

	// register INFIX parse functions
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	// read two tokens so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// advances both curToken and peekToken, skipping comments
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
	for {
		if p.curToken.Type != token.COMMENT {
			break
		} else {
			p.curToken = p.peekToken
			p.peekToken = p.l.NextToken()
		}
	}
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{} // the root node of every AST

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		// since the only two real statements are `let` and `return`,
		// everything else is dealt with as an expression
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	// after `let`, next token is an identifier (variable)
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// create identifier based on it
	stmt.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	// after `let $xxx`, next token is `=`; error if not
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()                          // move past =
	stmt.Value = p.parseExpression(LOWEST) // parse exp

	// now token is on exp, as in `let xxx = exp;`
	// if the next token is `;` move one up so at the next iteration
	// of parseProgram we skip the `;`
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	// skip semicolon if any
	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// Parsing Expressions

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(expression ast.Expression) ast.Expression
)

// add prefixParseFn for token type
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// add infixParseFn for token type
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// parses a whole exp statement, such as 1 + 2 + 3
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	// so that expression have optional `;`
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	// check if we have a prefix parsing function associated with
	// the current token type; the first element of an exp is always one of
	// IDENT, INT, BANG, MINUS
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.errors = append(p.errors, fmt.Sprintf("no prefix parse function found for %s", p.curToken.Type))
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// try to find an infix parse func for the next token
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

// get precedence for peek token (next token)
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// get precedence for current token
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseIdentifier() ast.Expression {
	// normal case, just an identifier
	if !p.peekTokenIs(token.ASSIGN) {
		return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}
	// we are reassigning a value to an identifier, e.g. `x = 5`
	// current token is `x`
	left := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken() // move to `=`
	tk := p.curToken

	// move to next token and parse it as an expression
	p.nextToken()
	right := p.parseExpression(LOWEST)

	return &ast.ReassignmentExpression{
		Token: tk,
		Left:  left,
		Right: right,
	}
}

func (p *Parser) parseInteger() ast.Expression {
	val, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("cannot parse %s as integer", p.curToken.Literal))
	}

	return &ast.IntegerLiteral{Token: p.curToken, Value: val}
}

func (p *Parser) parseString() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	exp := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()                         // got ! or -, now read the next token
	exp.Right = p.parseExpression(PREFIX) // with PREFIX precedence

	return exp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	exp := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)
	return exp
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	exp := &ast.IfExpression{Token: p.curToken}
	// curToken is `if`; expect ( and move on curToken
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken() // curToken is `(`; move it the exp
	exp.Condition = p.parseExpression(LOWEST)

	// expect ) and move on curToken
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// expect { and move on curToken
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// parse the whole { ... } block
	exp.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken() // move to `else`

		// expect { and move on curToken
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		exp.Alternative = p.parseBlockStatement()
	}

	return exp
}

func (p *Parser) parseWhileExpression() ast.Expression {
	exp := &ast.WhileExpression{Token: p.curToken}
	// curToken is `while`; expect ( and move on curToken
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken() // curToken is `(`; move to the exp
	exp.Condition = p.parseExpression(LOWEST)

	// expect ) and move on curToken
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// expect { and move on curToken
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// parse the whole { ... } block
	exp.Body = p.parseBlockStatement()
	return exp
}

func (p *Parser) parseForLoop() ast.Expression {
	exp := &ast.ForLoop{Token: p.curToken}
	// cur token is `for`; expect an identifier and move on curToken
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	exp.Iterator = p.parseIdentifier().(*ast.Identifier) // parse the iterator

	if !p.expectPeek(token.IN) { // curToken is `in`
		return nil
	}
	p.nextToken() // curToken is either `[` or an identifier

	if p.curTokenIs(token.LBRACKET) { // parse array literal
		exp.Elements = p.parseExpressionList(token.RBRACKET)
	} else { // parse identifier
		exp.Ident = p.parseIdentifier()
	}

	p.nextToken() // curToken is `{`
	exp.Body = p.parseBlockStatement()

	return exp
}

// example: given the block `{ x; let y = x; }`, it will return a BlockStatement
// object with two statements: `x` (an expression) and `let y = x` (a statement)
// (also check my test TestIfWithTwoStatements)
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}

	p.nextToken() // move after {

	// go on until you find } or EOF
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionExpression() ast.Expression {
	exp := &ast.FunctionLiteral{Token: p.curToken}

	// expect ( and move on it
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken() // move past (

	// go on until ) or EOF
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		// skip commas
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			continue
		}
		// create identifier and add it to params
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		exp.Params = append(exp.Params, ident)
		p.nextToken()
	}

	// expect { and move on curToken
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// parse the whole { ... } block
	exp.Body = p.parseBlockStatement()

	return exp
}

// both Call Expressions and Array literals are instances of parsing
// an expression list: only the parenthesis are different
// (arg1, arg2, ...) vs [elem1, elem2, ...]
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	args := []ast.Expression{}

	// end of args, we found the end token: `)` or `]` or whatever
	if p.peekTokenIs(end) {
		p.nextToken()
		return args
	}

	p.nextToken()                                  // move past `(` or `[`
	args = append(args, p.parseExpression(LOWEST)) // parse exp

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()                                  // move to the comma
		p.nextToken()                                  // move to the next exp
		args = append(args, p.parseExpression(LOWEST)) // parse exp
	}

	// no more commas, so we want the end-token next
	if !p.expectPeek(end) {
		return nil
	}
	return args
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken, Pairs: map[ast.Expression]ast.Expression{}}
	// tokens like: { exp : exp , exp : exp }

	p.nextToken() // move past {

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) { // cur token is key; this moves past it
			return nil
		}
		p.nextToken() // move past the :

		val := p.parseExpression(LOWEST)
		p.nextToken() // move past the value

		hash.Pairs[key] = val // add to the pairs

		if p.curTokenIs(token.COMMA) { // move past ,
			p.nextToken()
		}
	}

	return hash
}

func (p *Parser) parseMapFunction() ast.Expression {
	mf := &ast.MapFunction{Token: p.curToken}
	// tokens like: map(fn(x) { x * 2}, [1,2,3])

	// current token is `map`
	p.expectPeek(token.LPAREN) // expect next to be (, and move to it
	p.nextToken()              // move past (

	// I first used parseFunctionExpression(), but that would only work for function literals
	// and not identifiers. Then I remembered that in Monkey they are both expressions!
	mf.Function = p.parseExpression(LOWEST)

	p.expectPeek(token.COMMA)    // move cur to ,
	p.expectPeek(token.LBRACKET) // move cur to [

	mf.Elements = p.parseExpressionList(token.RBRACKET) // now cur token is ]

	p.expectPeek(token.RPAREN) // move cur to )

	return mf
}
