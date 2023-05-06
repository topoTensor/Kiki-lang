package main

import (
	"fmt"
	"log"
	"os"
)

// Executes the given function using recursive interpretation
func (inter *Inter) call_function(func_name string, arguments [][]Token, call_line, call_column int) Variable {
	print_comment("	\033[32mCALL_FUNCTION____________________	\033[39m", func_name, arguments, inter.variables)
	function, is_there := inter.functions[func_name]
	if !is_there {
		throw_error("no such function "+func_name, call_line, call_column)
	} else if len(arguments) != len(function.input_args) {
		throw_error("wrong number of arguments in function: "+func_name, call_line, call_column)
	}
	new_variables := map[string]Variable{}
	new_functions := map[string]Function{}
	new_arguments := []Token{}
	for _, a := range arguments {
		ret := inter.evaluate_set(a, "call function")
		q := Token{ret.typ, ret.val, a[0].line, a[0].column}
		new_arguments = append(new_arguments, q)
	}
	for i, a := range function.input_args {
		if new_arguments[i].typ == FUNCTION {
			new_functions[a.value.(string)] = new_arguments[i].value.(Function)
		} else {
			new_variables[a.value.(string)] = Variable{new_arguments[i].typ, new_arguments[i].value}
		}
	}
	newinter := new_interpreter(function.commands, inter.variables, inter.functions)
	merge_maps(&newinter.variables, new_variables)
	merge_maps(&newinter.functions, new_functions)
	return_var, _, _, _ := newinter.call_run()
	print_comment("	\033[34mNEWINTER VARIABLES:", return_var, "INPUT:", newinter.array, "VARIABLES:", inter.variables, "	\033[39m")
	return return_var["RETURN_FROM_CALL"]
}

func remove_token(array []Token, name string) []Token {
	newarr := []Token{}
	for _, v := range array {
		if v.value != name {
			newarr = append(newarr, v)
		}
	}
	return newarr
}

// 2+2,3 -> [[2+2],[3]]
func lex_list(input []Token) [][]Token {
	if len(input) == 0 {
		return [][]Token{}
	}
	iterator := new_iterator(input)
	output := [][]Token{}
	left_index := 0
	in_function := false
	function_end_index := -1
	print_comment("LEX_LIST INPUT:", input)
	for {
		if !in_function {
			if iterator.can_peek() && iterator.current.typ == IDENTIFIER && iterator.peek().value == LPAREN {
				iterator.advance()
				function_end_index = search_boundaries(iterator, LPAREN, RPAREN)
				in_function = true
			} else if iterator.current.value == COMMA {
				output = append(output, input[left_index:iterator.index])
				left_index = iterator.index + 1
			} else if iterator.index+1 == len(input) {
				output = append(output, input[left_index:iterator.index+1])
				return output
			}
		} else {
			for iterator.index != function_end_index && iterator.index+2 < len(iterator.array) {
				iterator.advance()
			}
			in_function = false
		}
		iterator.advance()
	}
}

// doesn't move the interpreter
func (inter *Inter) replace_identifiers(input []Token, from string) []Token {
	print_comment("\033[33m REPLACE IDENTIFIERS:\033[39m", input, "FROM", from)
	if len(input) == 0 {
		return []Token{}
	}
	iterator := new_iterator(input)
	new_input := []Token{}
	for {
		//
		if iterator.current.typ == IDENTIFIER {
			if iterator.can_peek() && iterator.peek().value == LPAREN {
				// func
				func_name := iterator.current.value.(string)
				if _, ok := inter.functions[func_name]; ok {
					// do arguments
					iterator.advance()
					bound_index := search_boundaries(iterator, LPAREN, RPAREN)
					// lex list . replace_identifiers . input[from lparen to rparen]
					ret := inter.call_function(func_name,
						lex_list(inter.replace_identifiers(input[iterator.index+1:bound_index], "replace_identifiers function in function section")),
						iterator.current.line,
						iterator.current.column)
					new_input = append(new_input, Token{ret.typ, ret.val, iterator.current.line, iterator.current.column})
					iterator.move_index(bound_index)
				} else {
					if iterator.current.value == "len" {
						iterator.advance()
						bound_index := search_boundaries(iterator, LPAREN, RPAREN)
						lex := lex_list(inter.replace_identifiers(input[iterator.index+1:bound_index], "len function"))
						if len(lex) > 1 {
							throw_error("wrong number of arguments for len function", inter.current.line, inter.current.left_side[0].column)
						}
						lex2 := inter.Shunting_Yard(lex[0], "len function")
						if lex2.typ != ARRAY && lex2.typ != ARRAYSTRING {
							throw_error("wrong type for len function", inter.current.line, inter.current.left_side[0].column)
						}
						print_comment("LEX2", lex2)
						length := len(lex2.value.([]Token))
						new_input = append(new_input, Token{NUMBER, float64(length), iterator.current.line, iterator.current.column})
						iterator.move_index(bound_index)
					} else {
						if func_name == "read" {
							var r string
							_, err := fmt.Scanln(&r)
							if err != nil {
								log.Fatal(err)
							}
							scanner := new_scanner(r)
							lex := scanner.lex()
							val := inter.Shunting_Yard(lex, "read function evaluation")
							new_input = append(new_input, val)
						} else if func_name == "byte" {
							iterator.advance()
							bound_index := search_boundaries(iterator, LPAREN, RPAREN)
							lex := lex_list(inter.replace_identifiers(input[iterator.index+1:bound_index], "len function"))
							if len(lex) > 1 {
								throw_error("function `byte` accepts only one string character as an argument", iterator.current.line, iterator.current.column)
							}
							val := inter.Shunting_Yard(lex[0], "from byte function evaluation")
							new_input = append(new_input, Token{NUMBER, float64([]byte(val.value.(string))[0]), iterator.current.line, iterator.current.column})
							iterator.set_current(bound_index)
						} else if func_name == "file_read" {
							iterator.advance()
							bound_index := search_boundaries(iterator, LPAREN, RPAREN)
							lex := lex_list(inter.replace_identifiers(input[iterator.index+1:bound_index], "len function"))
							if len(lex) > 1 {
								throw_error("function `file_read` accepts only one file name as an argument", iterator.current.line, iterator.current.column)
							}
							val := inter.Shunting_Yard(lex[0], "from byte function evaluation")
							data, err := os.ReadFile(sarray_to_string(val).value.(string))
							check_err(err)
							new_input = append(new_input, Token{STRING, string(data), iterator.current.line, iterator.current.column})
							iterator.set_current(bound_index)
						} else {
							throw_error("no such function "+fmt.Sprintf("%v", iterator.current), iterator.current.line, iterator.current.column)
						}
					}
				}
			} else {
				// iden
				if val, ok := inter.variables[iterator.current.value.(string)]; ok {
					new_input = append(new_input, Token{val.typ, val.val, iterator.current.line, iterator.current.column})
				} else if iterator.back().value == REFERENCE {
					new_input = append(new_input, Token{FUNCTION, inter.functions[iterator.current.value.(string)], iterator.current.line, iterator.current.column})
				} else {
					throw_error("no such variable "+fmt.Sprintf("%v", iterator.current.value), iterator.current.line, iterator.current.column)
				}
			}
		} else if iterator.current.value == LBRACK {
			// array
			save_index := iterator.index
			array := go_to_boundary(&iterator, LBRACK, RBRACK)
			array = array[:len(array)-1]
			lex := lex_list(array)
			newlex := []Token{}
			for _, l := range lex {
				newlex = append(newlex, inter.Shunting_Yard(l, "lexing of replace identifiers array section"))
			}
			new_input = append(new_input, Token{ARRAY, newlex, iterator.array[save_index].line, iterator.array[save_index].column})
		} else {
			// number/bool/sign
			if iterator.current.value != REFERENCE {
				new_input = append(new_input, iterator.current)
			}
		}
		//
		if !iterator.can_peek() {
			break
		}
		iterator.advance()
	}
	print_comment("\033[33m OUTPUT RETURN IDENT \033[39m", new_input)
	return new_input
}

func String_tok_to_array(tok Token) Token {
	arr := []Token{}
	for _, v := range tok.value.(string) {
		arr = append(arr, Token{NUMBER, float64(v), tok.line, tok.column})
	}
	return Token{ARRAYSTRING, arr, tok.line, tok.column}
}

func sarray_to_string(tok Token) Token {
	str := ""
	for _, v := range tok.value.([]Token) {
		str += string(byte(v.value.(float64)))
	}
	return Token{STRING, str, tok.line, tok.column}
}

func (inter *Inter) Shunting_Yard(input []Token, from string) Token {
	precendence := map[string]int{
		TILDA: 6, POWER: 3,
		EQUAL_EQUAL: 4, NOT_EQUAL: 4, GREATER: 4, GREATER_EQUAL: 4, LESS: 4, LESS_EQUAL: 4,
		BANG: 4, AND: 3, PIPE: 2,
		SLASH: 5, STAR: 5, MINUS: 4, PLUS: 4,
		LPAREN: 0, RPAREN: 0}
	operators := new_stack[Token](nil)
	output := new_stack[Token](nil)

	if input[0].typ == FUNCTION {
		return input[0]
	}
	input = inter.replace_identifiers(input, "shunting yard")
	if len(input) == 0 {
		throw_error("nothing to process in shunting yard", inter.current.line, -1)
	}
	for _, i := range input {
		print_comment("input:", i)
	}

	for _, n := range input {
		if n.typ == NUMBER || n.typ == BOOLEAN || n.typ == ARRAY || n.typ == ARRAYSTRING {
			output.push(n)
		} else if n.typ == STRING {
			output.push(String_tok_to_array(n))
		} else if n.typ == SIGN {
			if n.value == LPAREN {
				operators.push(n)
			} else if n.value == RPAREN {
				for operators.last().value != LPAREN {
					output.push(operators.pop())
				}
				operators.pop()
			} else {
				if operators.len > 0 {
					if _, ok := precendence[operators.last().value.(string)]; !ok {
						throw_error("You forgot to add a new value in precendence map in eval.go: "+fmt.Sprintf("%v", operators.last().value.(string)), -1, -1)
					}
				}
				if operators.len > 0 && precendence[operators.last().value.(string)] >= precendence[n.value.(string)] {
					output.push(operators.pop())
					operators.push(n)
				} else {
					operators.push(n)
				}
			}
		}
	}
	for operators.len > 0 {
		output.push(operators.pop())
	}
	print_commentf("operators %v\n output: %v\n", operators, output)
	for _, o := range output.array {
		print_commentf("%v ", o.value)
	}
	print_comment()

	res := Calculate_Postfix(output.array)
	print_commentf("\033[94m return from shunting yard: %v, from %s \033[39m\n", res, from)
	return res
}

func eval_token_express(t1, t2, op Token) Token {
	print_comment("t1 t2", t1.typ, t2.typ)
	// ! BETTER TYPE CHECKING
	if op.value == PLUS || op.value == MINUS || op.value == STAR || op.value == SLASH || op.value == GREATER || op.value == GREATER_EQUAL || op.value == LESS || op.value == LESS_EQUAL {
		if t1.typ != NUMBER || t2.typ != NUMBER {
			throw_error("can't do operations on different value types "+fmt.Sprintf("%v", t1)+" "+fmt.Sprintf("%v", t2)+" with sign "+fmt.Sprintf("%v", op), t1.line, t1.column)
		}
	} else if op.value == AND || op.value == PIPE {
		if t1.typ != BOOLEAN || t2.typ != BOOLEAN {
			throw_error("can't do operations on different value types "+fmt.Sprintf("%v", t1)+" "+fmt.Sprintf("%v", t2)+" with sign "+fmt.Sprintf("%v", op), t1.line, t1.column)
		}
	} else if op.value == TILDA {
		if t1.typ != ARRAY && t1.typ != ARRAYSTRING && t2.typ != NUMBER {
			throw_error("can't do operations on different value types "+fmt.Sprintf("%v", t1)+" "+fmt.Sprintf("%v", t2)+" with sign "+fmt.Sprintf("%v", op), t1.line, t1.column)
		}
	} else if op.value == POWER {
		if (t1.typ != ARRAY && t1.typ != ARRAYSTRING) || (t2.typ != ARRAY && t2.typ != ARRAYSTRING) {
			throw_error("can't do operations on different value types "+fmt.Sprintf("%v", t1)+" "+fmt.Sprintf("%v", t2)+" with sign "+fmt.Sprintf("%v", op), t1.line, t1.column)
		}
	}
	switch op.value {
	case PLUS:
		return Token{NUMBER, t1.value.(float64) + t2.value.(float64), t1.line, t2.column}
	case MINUS:
		return Token{NUMBER, t1.value.(float64) - t2.value.(float64), t1.line, t2.column}
	case STAR:
		return Token{NUMBER, t1.value.(float64) * t2.value.(float64), t1.line, t2.column}
	case SLASH:
		return Token{NUMBER, t1.value.(float64) / t2.value.(float64), t1.line, t2.column}
	case EQUAL_EQUAL:
		if t1.typ == NUMBER {
			return Token{BOOLEAN, t1.value.(float64) == t2.value.(float64), t1.line, t2.column}
		} else {
			return Token{BOOLEAN, t1.value.(bool) == t2.value.(bool), t1.line, t2.column}
		}
	case NOT_EQUAL:
		if t1.typ == NUMBER {
			return Token{BOOLEAN, t1.value.(float64) != t2.value.(float64), t1.line, t2.column}
		} else {
			return Token{BOOLEAN, t1.value.(bool) != t2.value.(bool), t1.line, t2.column}
		}
	case GREATER:
		return Token{BOOLEAN, t1.value.(float64) > t2.value.(float64), t1.line, t2.column}
	case GREATER_EQUAL:
		return Token{BOOLEAN, t1.value.(float64) >= t2.value.(float64), t1.line, t2.column}
	case LESS:
		return Token{BOOLEAN, t1.value.(float64) < t2.value.(float64), t1.line, t2.column}
	case LESS_EQUAL:
		return Token{BOOLEAN, t1.value.(float64) <= t2.value.(float64), t1.line, t2.column}
	case AND:
		return Token{BOOLEAN, t1.value.(bool) && t2.value.(bool), t1.line, t2.column}
	case PIPE:
		return Token{BOOLEAN, t1.value.(bool) || t2.value.(bool), t1.line, t2.column}
	case TILDA:
		item := t1.value.([]Token)[int(t2.value.(float64))]
		return Token{item.typ, item.value, item.line, item.column}
	case POWER:
		if t1.typ == ARRAYSTRING {
			return Token{ARRAYSTRING, append(t1.value.([]Token), t2.value.([]Token)...), t1.line, t2.column}
		} else {
			return Token{ARRAY, append(t1.value.([]Token), t2.value.([]Token)...), t1.line, t2.column}
		}
	}
	throw_error("Wrong operation"+fmt.Sprintf("%v", op)+" for tokens"+fmt.Sprintf("%v %v", t1, t2), t1.line, t1.column)
	return Token{}
}
func eval_bang(tok Token) Token {
	return Token{BOOLEAN, !tok.value.(bool), tok.line, tok.column}
}
func Calculate_Postfix(arr []Token) Token {
	i := 2
	for {
		if len(arr) <= 1 {
			break
		}
		if arr[i-1].value == BANG {
			arr = replace_array_part(arr, i-2, i-1, eval_bang(arr[i-2]))
			i = 1
		} else if arr[i].typ == SIGN && arr[i].value != BANG {
			arr = replace_array_part(arr, i-2, i, eval_token_express(arr[i-2], arr[i-1], arr[i]))
			i = 1
		}
		i++
	}
	return arr[len(arr)-1]
}
