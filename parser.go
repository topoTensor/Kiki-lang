package main

import (
	"fmt"
	"math"
)

type Command struct {
	depth      int
	typ        string
	left_side  []Token
	right_side []Token
	line       int
}

type Parser struct {
	Iterator[Token]
	commands []Command
}

func new_parser(itokens []Token) Parser {
	return Parser{Iterator[Token]{-1, itokens[0], itokens}, []Command{}}
}
func (p *Parser) restart() {
	p.index = -1
	p.current = p.array[0]
}
func (p *Parser) advance() Token {
	p.index++
	p.current = p.array[p.index]
	return p.current
}
func (p *Parser) peek() Token {
	return p.array[p.index+1]
}
func (p *Parser) back() Token {
	return p.array[p.index-1]
}
func (p *Parser) addCommand(c Command) {
	p.commands = append(p.commands, c)
}

// Advances until the 'end' Token, returning the tokens before it
func (p *Parser) findEnd(end Token) []Token {
	ret := []Token{}
	for !(p.peek().typ == end.typ && p.peek().value == end.value) {
		if p.index+2 >= len(p.array) {
			throw_error("can't find the token "+fmt.Sprintf("%v", TO_LONGER[end.value.(string)]), p.current.line, p.current.column)
		}
		ret = append(ret, p.advance())
	}
	p.advance()
	return ret
}

// called if the sign after indentifier is
func (p Parser) sign_error() {
	for _, t := range p.array {
		if t.line == p.peek().line {
			print_commentf("%v ", t.value)
		} else if t.line > p.peek().line {
			break
		}
	}
	throw_error("wrong token ("+fmt.Sprintf("%v", p.peek().value)+") after identifier ("+fmt.Sprintf("%v", p.current.value)+")", p.peek().line, p.peek().column)
}

// Replaces the dot token with a float number, using two numbers before and after the dot
func (p *Parser) replace_dot() {
	newtokens := p.array[:p.index-1]
	x := p.peek().value.(float64)
	number := p.back().value.(float64) + (x / (math.Pow(10, math.Floor(math.Log10(x)+1))))
	newtokens = append(newtokens, Token{NUMBER, number, p.back().line, p.back().column})
	newtokens = append(newtokens, p.array[p.index+2:]...)
	p.array = newtokens
}

// replaces the minus token with the number infront
func (p *Parser) replace_minus() {
	newtokens := p.array[:p.index]
	number := -p.peek().value.(float64)
	newtokens = append(newtokens, Token{NUMBER, number, p.back().line, p.back().column})
	newtokens = append(newtokens, p.array[p.index+2:]...)
	p.array = newtokens
}

func (p *Parser) parse_function() []Token {
	if p.advance().value != LPAREN {
		throw_error("wrong left parenth at", p.current.line, p.current.column)
	}
	arguments := []Token{}
	for p.peek().value != RPAREN {
		if p.advance().value != COMMA {
			arguments = append(arguments, p.current)
		}
	}
	return arguments
}

func (p *Parser) search_for_token(token_val string, token_type, err_msg string) {
	save_pos := p.index
	count := 0
	for {
		if !p.can_peek() {
			throw_error(err_msg, p.array[save_pos].line, p.array[save_pos].column)
		}
		p.advance()

		if p.current.value == WHILE || p.current.value == IF || p.current.value == FUNCTION {
			count++
		} else if p.current.value == END {
			if count == 0 {
				break
			}
			count--
		}
	}
	p.set_current(save_pos)
}

func (p *Parser) parse() []Command {
	print_comment(p.array)
	// Change dot token into numbers
	for p.current.typ != EOF {
		p.advance()
		if p.current.value == DOT {
			if p.back().typ == NUMBER && p.peek().typ == NUMBER {
				p.replace_dot()
			} else {
				throw_error("wrong used dot", p.current.line, p.current.column)
			}
		}
	}
	p.restart()
	// Correct numbers with minus before it
	for p.current.typ != EOF {
		p.advance()
		if p.current.value == MINUS {
			if p.back().typ == SIGN && p.peek().typ == NUMBER {
				p.replace_minus()
			} else if (p.back().typ != NUMBER && p.back().value != RPAREN) && (p.peek().typ != NUMBER && p.peek().value != LPAREN) {
				throw_error("wrong used minus", p.current.line, p.current.column)
			}
		}
	}
	p.restart()
	depth := 0
	for p.current.typ != EOF {
		p.advance()
		switch p.current.typ {
		case IDENTIFIER:
			if p.peek().typ != SIGN {
				p.sign_error()
			}
			switch p.advance().value {
			case EQUAL:
				if !p.can_peek() || p.peek().value == NEWLINE {
					throw_error("wrong used definition, nothing to process", p.current.line, p.current.column)
				}
				p.addCommand(Command{depth, SET, []Token{p.back()}, p.findEnd(Token{SIGN, NEWLINE, 0, 0}), p.current.line})
			default:
				if p.back().value == "print" {
					p.addCommand(Command{depth, PRINT, p.array[p.index+1 : search_boundaries(p.Iterator, LPAREN, RPAREN)], []Token{}, p.current.line})
					go_to_boundary(&p.Iterator, LPAREN, RPAREN)
				} else if p.back().value == "panic" {
					p.addCommand(Command{depth, PANIC, p.array[p.index+1 : search_boundaries(p.Iterator, LPAREN, RPAREN)], []Token{}, p.current.line})
					go_to_boundary(&p.Iterator, LPAREN, RPAREN)
				} else if p.back().value == "file_write" {
					p.addCommand(Command{depth, FILE_WRITE, p.array[p.index+1 : search_boundaries(p.Iterator, LPAREN, RPAREN)], []Token{}, p.current.line})
					go_to_boundary(&p.Iterator, LPAREN, RPAREN)
				} else {
					throw_error("wrong used identifier "+fmt.Sprintf("%v", p.back().value), p.back().line, p.back().column)
				}
			}
		case KEYWORD:
			switch p.current.value {
			case IF:
				p.addCommand(Command{depth, IF, p.findEnd(Token{KEYWORD, THEN, 0, 0}), []Token{}, p.current.line})
				depth++
				p.search_for_token(END, KEYWORD, "the IF statement isn't closed")
			case ELSEIF:
				depth--
				p.addCommand(Command{depth, ELSEIF, p.findEnd(Token{KEYWORD, THEN, 0, 0}), []Token{}, p.current.line})
				p.search_for_token(END, KEYWORD, "the ELSEIF statement isn't closed")
				depth++
			case ELSE:
				depth--
				p.addCommand(Command{depth, ELSE, []Token{}, []Token{}, p.current.line})
				p.search_for_token(END, KEYWORD, "the ELSE statement isn't closed")
				depth++
			case END:
				depth--
				p.addCommand(Command{depth, END, []Token{}, []Token{}, p.current.line})
			case WHILE:
				p.addCommand(Command{depth, WHILE, p.findEnd(Token{KEYWORD, DO, 0, 0}), []Token{}, p.current.line})
				p.search_for_token(END, KEYWORD, "the WHILE statement isn't closed")
				depth++
			case BREAK:
				p.addCommand(Command{depth, BREAK, []Token{}, []Token{}, p.current.line})
			case FUNCTION:
				p.addCommand(Command{depth, FUNCTION, []Token{p.advance()}, p.parse_function(), p.current.line})
				p.search_for_token(END, KEYWORD, "the FUNCTION statement isn't closed")
				depth++
			case RETURN:
				p.addCommand(Command{depth, RETURN, p.findEnd(Token{SIGN, NEWLINE, 0, 0}), []Token{}, p.current.line})
			}
		case SIGN:
		case EOF:
		default:
			throw_error("wrong used token "+fmt.Sprintf("%v", p.current.value), p.current.line, p.current.column)
		}
	}
	p.addCommand(Command{depth, EOF, []Token{}, []Token{}, p.current.line})
	return p.commands
}
