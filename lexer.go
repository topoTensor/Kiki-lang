package main

import (
	"strconv"
	"strings"
)

const (
	NUMBER      = "NUM"
	SIGN        = "SGN"
	KEYWORD     = "KWD"
	IDENTIFIER  = "IDN"
	BOOLEAN     = "BOL"
	ARRAY       = "ARR"
	ARRAYSTRING = "ARS"
	STRING      = "STR"

	NEWLINE = "NWL"

	PLUS   = "PLS"
	MINUS  = "MIN"
	SLASH  = "SLH"
	STAR   = "STR"
	LPAREN = "LPR"
	RPAREN = "RPR"
	EQUAL  = "EQL"
	BANG   = "BNG"

	DOT   = "DOT"
	COMMA = "COM"

	TRUE  = "True"
	FALSE = "False"
	PIPE  = "PIP"
	AND   = "AND"

	LBRACK = "LBR"
	RBRACK = "RBR"
	TILDA  = "TLD"
	POWER  = "POW"

	QUOTIENT = "QUO"

	EQUAL_EQUAL   = "EQQ"
	NOT_EQUAL     = "NEQ"
	LESS          = "LSS"
	GREATER       = "GRT"
	LESS_EQUAL    = "LSE"
	GREATER_EQUAL = "GRE"

	REFERENCE = "REF"

	END      = "END"
	IF       = "IF "
	THEN     = "THN"
	ELSEIF   = "ELF"
	ELSE     = "ELS"
	WHILE    = "WHL"
	DO       = "DO "
	BREAK    = "BRK"
	FUNCTION = "FNC"
	RETURN   = "RET"

	EOF = "EOF"

	// command type used by parser
	SET        = "SET"
	PRINT      = "PRT"
	PANIC      = "PNC"
	FILE_WRITE = "FWR"
)

// used in the lexer to compare the names to the tokens (which are in a shorter version for comfort)
var TO_SHORTEN map[string]string = map[string]string{"IF": IF, "ELSEIF": ELSEIF, "ELSE": ELSE, "THEN": THEN, "WHILE": WHILE, "BREAK": BREAK, "DO": DO, "FUNCTION": FUNCTION, "RETURN": RETURN, "END": END}

// used in parser 'findEnd' function for representation in error messages
var TO_LONGER map[string]string = map[string]string{IF: "IF", ELSEIF: "ELSEIF", ELSE: "ELSE", THEN: "THEN", WHILE: "WHILE", BREAK: "BREAK", DO: "DO", FUNCTION: "FUNCTION", RETURN: "RETURN", END: "END"}

type Token struct {
	typ          string
	value        interface{}
	line, column int
}

type Scanner struct {
	tokens  []Token
	source  string
	index   int
	current byte
	line    int
	column  int
}

func new_scanner(data_str string) Scanner {
	return Scanner{[]Token{}, data_str, 0, ' ', 1, 1}
}

func (s *Scanner) advance() byte {
	s.index++
	s.column++
	if s.has_next_byte(0) {
		s.current = s.source[s.index]
	}
	if s.current == '\n' {
		s.line++
		s.column = 0
	}
	return s.current
}

func (s *Scanner) peek() byte {
	return s.source[s.index+1]
}

func (s *Scanner) addToken(typ string, value interface{}) {
	s.tokens = append(s.tokens, Token{typ, value, s.line, s.column})
}

func char_is_number(c byte) bool {
	return c >= '0' && c <= '9'
}

func char_is_alpha(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_'
}

func (s Scanner) has_next_byte(offset int) bool {
	return s.index+offset < len(s.source)
}

// I DON'T USE ITERATOR BECAUSE THE LEXER HAS A BIT DIFFERENT APPROACH, IT DOESN'T USE AN ARRAY AND HAS MANY CONDITIONS WITH ADVANCE METHOD
func (s *Scanner) lex() []Token {
	s.current = s.source[0]
	for s.has_next_byte(0) {
		switch s.current {
		case '=':
			if s.peek() == '=' {
				s.addToken(SIGN, EQUAL_EQUAL)
				s.advance()
			} else {
				s.addToken(SIGN, EQUAL)
			}
		case '>':
			if s.peek() == '=' {
				s.addToken(SIGN, GREATER_EQUAL)
				s.advance()
			} else {
				s.addToken(SIGN, GREATER)
			}
		case '<':
			if s.peek() == '=' {
				s.addToken(SIGN, LESS_EQUAL)
				s.advance()

			} else {
				s.addToken(SIGN, LESS)
			}
		case '!':
			if s.peek() == '=' {
				s.addToken(SIGN, NOT_EQUAL)
				s.advance()

			} else {
				s.addToken(SIGN, BANG)
			}
		case '+':
			s.addToken(SIGN, PLUS)
		case '-':
			s.addToken(SIGN, MINUS)
		case '*':
			s.addToken(SIGN, STAR)
		case '/':
			s.addToken(SIGN, SLASH)
		case '(':
			s.addToken(SIGN, LPAREN)
		case ')':
			s.addToken(SIGN, RPAREN)
		case '[':
			s.addToken(SIGN, LBRACK)
		case ']':
			s.addToken(SIGN, RBRACK)
		case '.':
			s.addToken(SIGN, DOT)
		case ',':
			s.addToken(SIGN, COMMA)
		case '^':
			s.addToken(SIGN, POWER)
		case '|':
			s.addToken(SIGN, PIPE)
		case '~':
			s.addToken(SIGN, TILDA)
		case '&':
			s.addToken(SIGN, AND)
		case '$':
			s.addToken(SIGN, REFERENCE)
		case '"':
			str := ""
			str += string(s.advance())
			for s.has_next_byte(0) && s.current != '"' {
				str += string(s.advance())
			}
			s.addToken(STRING, str[:len(str)-1]) // skip the last "
		case '#':
			for s.current != '\n' && s.has_next_byte(0) {
				s.advance()
			}
		case '\n':
			s.addToken(SIGN, NEWLINE)
		case ' ':
		case '\t':
		case 13:
		default:
			if char_is_number(s.current) {
				var number string = string(s.current)
				for s.has_next_byte(1) && char_is_number(s.peek()) {
					number += string(s.advance())
				}
				conv, err := strconv.ParseFloat(number, 64)
				if err != nil {
					throw_error("can't parse into float - "+number, s.line, s.column)
				}
				s.addToken(NUMBER, conv)
			} else if char_is_alpha(s.current) {
				var str string = string(s.current)
				for s.has_next_byte(1) && (char_is_alpha(s.peek()) || char_is_number(s.peek())) {
					str += string(s.advance())
				}

				upstr := strings.ToUpper(str)

				if short_ver, is_there := TO_SHORTEN[upstr]; is_there {
					s.addToken(KEYWORD, short_ver)
				} else if str == TRUE {
					s.addToken(BOOLEAN, true)
				} else if str == FALSE {
					s.addToken(BOOLEAN, false)
				} else {
					s.addToken(IDENTIFIER, str)
				}

			} else {
				throw_error("syntax error, can't process the data - "+string(s.current), s.line, s.column)
			}
		}
		s.advance()
	}
	s.addToken(EOF, EOF)
	return s.tokens
}
