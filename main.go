package main

import (
	"fmt"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/repl"
	"os"
	"os/user"
)

func main() {

	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	if len(os.Args) == 1 {
		fmt.Printf("Hello %s !\n", u.Username)
		repl.Start(os.Stdin, os.Stdout)
	}

	if len(os.Args) == 2 {
		data, err := os.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
		l := lexer.New(string(data))
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, e := range p.Errors() {
				fmt.Println("Parse error: ", e)
			}
			return
		}

		env := object.NewEnvironment()
		_ = evaluator.Eval(program, env)
	}
	return
}
