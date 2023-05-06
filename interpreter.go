package main

import (
	"fmt"
	"log"
	"os"
)

type Variable struct {
	typ string
	val interface{}
}

type Inter struct {
	Iterator[Command]
	variables map[string]Variable
	functions map[string]Function
}

type Function struct {
	line       int
	input_args []Token
	commands   []Command
}

func map_copy[T any](m map[string]T) map[string]T {
	newm := make(map[string]T)
	for k, v := range m {
		newm[k] = v
	}
	return newm
}

// * You have to copy maps, otherwise they will share the same memory in the nested interpretation
func new_interpreter(commands []Command, input_variables map[string]Variable, functions map[string]Function) Inter {
	return Inter{Iterator[Command]{0, commands[0], commands}, map_copy(input_variables), map_copy(functions)}
}

func (inter *Inter) add_variable(name string, val Variable) {
	inter.variables[name] = val
}

func (inter *Inter) evaluate_set(tokens []Token, msg string) Variable {
	print_comment("evaluate_set", tokens)
	ret_tok := inter.Shunting_Yard(tokens, msg)
	return Variable{ret_tok.typ, ret_tok.value}
}

func merge_maps[T any](dest *map[string]T, src map[string]T) {
	for k, v := range src {
		(*dest)[k] = v
	}
}

func find_biggest_depth(arr []Command) int {
	max := 0
	for _, c := range arr {
		if c.depth > max {
			max = c.depth
		}
	}
	return max
}

func print_command(c Command, biggest_depth int) {
	for i := 0; i < c.depth; i++ {
		print_commentf("%s", "    ")
	}
	print_commentf("\033[95m %d \033[93m [%v] \033[39m", c.depth, c.typ)
	for i := 0; i < biggest_depth-c.depth; i++ {
		print_commentf("%s", "    ")
	}
	print_commentf("%d	 %v | %v ( ", c.line, c.left_side, c.right_side)
	for _, q := range c.left_side {
		print_commentf("%v ", q.value)
	}
	for _, q := range c.right_side {
		print_commentf("%v ", q.value)
	}
	print_comment(") ")
}
func print_function(name string, f Function) {
	print_commentf("   \033[96m FUNCTION \033[93m %s \033[39m with arguments \033[93m %v \033[39m  at line : \033[93m  %d \033[39m , \033[93m COMMANDS:\n", name, f.input_args, f.line)
	biggest_depth := find_biggest_depth(f.commands)
	for _, c := range f.commands {
		print_command(c, biggest_depth)
	}
	print_comment("____________________________________________\033[39m")
}
func print_array(tok Token) {
	print_commentf("\033[96m ARRAY: \033[39m")
	for _, q := range tok.value.([]interface{}) {
		print_commentf("%v, ", q)
	}
	print_comment()
}

func (inter *Inter) run_decision(cond bool) (did_break bool, did_return bool) {
	if cond {
		exec_commands := []Command{}
		save_depth := inter.current.depth
		for inter.advance().depth != save_depth {
			exec_commands = append(exec_commands, inter.current)
		}
		for !(inter.current.depth == save_depth && inter.current.typ == END) {
			inter.advance()
		}
		newinter := new_interpreter(exec_commands, inter.variables, inter.functions)
		return_variables, return_functions, break_call, return_call := newinter.call_run()
		merge_maps(&inter.variables, return_variables)
		merge_maps(&inter.functions, return_functions)
		return break_call, return_call
	} else {
		save_depth := inter.current.depth
		for inter.peek().depth != save_depth {
			inter.advance()
		}
	}
	return false, false
}

func printcall_arrs(print_obj Token) {
	print_iter := new_iterator(print_obj.value.([]Token))
	for {
		// don't forget that arrstring values are float64
		if print_iter.can_peek() && print_iter.current.value == 92.0 && print_iter.peek().value == 110.0 {
			print_iter.advance()
			fmt.Println()
		} else {
			fmt.Printf("%v", string(byte(print_iter.current.value.(float64))))
		}
		if print_iter.can_peek() {
			print_iter.advance()
		} else {
			break
		}
	}
}

// NOTES ABOUT BREAKING THE STATEMENT: remember that if and while loops might be breaked and transfered from the depths of interpreter calls
var RETURN_FROM_CALL string = "RETURN_FROM_CALL"

// * Some things to keep in mind about the langauge.
/*
* it's pure, functions always return, and return a single value
* functions can return only one type of values
* it's strong typed, you can't add boolean to number
* (Not yet implemented) strings are arrays of numbers
* reading from console evaluates the value on read
 */

/*
	LET'S TRY TO EVALUATE DIFFERENT VALUES AND TRY TO ERROR DETECT BASED ON THE POSTFIX, FOR EXAMPLE IF IT TRIES TO ADD NUMBER TO A BOOLEAN, IT SHOULD RETURN AN ERROR. WE WILL CAST THE INTERFACE BASED ON SECOND RETURN VALUE, THE TYPE TO CAST
*/

func (inter *Inter) call_run() (return_variables map[string]Variable, return_functions map[string]Function, is_breaked bool, is_returned bool) {
	// IF IT WAS INNER CALL (PARSERLESS), THE EOF WILL BE ADDED AUTOMATICLY
	if inter.array[len(inter.array)-1].typ != EOF {
		inter.array = append(inter.array, Command{0, EOF, []Token{}, []Token{}, inter.array[len(inter.array)-1].line + 1})
	}
	print_comment("\033[91m new call------------------------	\033[39m")
	print_comment("\033[35m COMMANDS to run-----------------\033[39m")
	biggest_depth := find_biggest_depth(inter.array)
	for _, c := range inter.array {
		print_command(c, biggest_depth)
	}
	print_comment("\033[35m-----------------------\033[39m")
	for inter.current.typ != EOF {
		switch inter.current.typ {
		case PRINT:
			lex := lex_list(inter.current.left_side)
			print_objects := []Token{}
			for _, l := range lex {
				print_objects = append(print_objects, inter.Shunting_Yard(l, "print call"))
			}
			for i, p := range print_objects {
				if DO_COMMENT {
					print_commentf("\033[93m === PRINT CALL === \033[39m %v\n", p)
				} else {
					if p.typ == NUMBER || p.typ == BOOLEAN {
						fmt.Printf("%v", p.value)
					} else if p.typ == ARRAY {
						fmt.Print("[")
						// ! CORRECT THE LAST ,
						for _, t := range p.value.([]Token) {
							if t.typ == ARRAYSTRING {
								fmt.Print("\"")
								printcall_arrs(t)
								fmt.Print("\" ")
							} else {
								fmt.Printf("%v, ", t.value)
							}
						}
						fmt.Print("]")
					} else if p.typ == ARRAYSTRING {
						printcall_arrs(p)
					}
				}
				if len(print_objects) > 1 && i+1 != len(print_objects) {
					if i == 0 || (i > 1 && print_objects[i-1].value != "\n") {
						fmt.Print("    ")
					}
				}
			}
		case PANIC:
			// * SHOULD I CHANGE THIS AS A PRINT STATEMENT IS MADE?
			print_obj := inter.Shunting_Yard(inter.current.left_side, "panic call")
			if DO_COMMENT {
				print_commentf("\033[91m !!! PANIC CALL !!! \033[39m %v\n", print_obj)
				log.Fatal()
			} else {
				fmt.Print("\033[91mpanic: \033[39m")
				if print_obj.typ == NUMBER || print_obj.typ == BOOLEAN {
					fmt.Printf("%v\n", print_obj)
				} else if print_obj.typ == ARRAY {
					for _, t := range print_obj.value.([]Token) {
						if t.typ == ARRAYSTRING {
							fmt.Print("\"")
							printcall_arrs(t)
							fmt.Print("\" ")
						} else {
							fmt.Printf("%v ", t)
						}
					}
					fmt.Println()
				} else if print_obj.typ == ARRAYSTRING {
					printcall_arrs(print_obj)
					fmt.Println()
				}
				log.Fatal()
			}
		case FILE_WRITE:
			args := []Token{}
			for _, a := range lex_list(inter.current.left_side) {
				args = append(args, sarray_to_string(inter.Shunting_Yard(a, "FILE_WRITE function")))
			}
			if len(args) != 2 {
				throw_error("`file_write` function accepts only 2 arguments, name of the file and text to write", inter.current.line, inter.current.left_side[0].column)
			}
			err := os.WriteFile(args[0].value.(string), []byte(args[1].value.(string)), 0644)
			check_err(err)
		case SET:
			print_comment("\033[92m setting variable: \033[39m", inter.current)
			v := inter.evaluate_set(inter.current.right_side, "set statement")
			inter.add_variable(fmt.Sprintf("%v", inter.current.left_side[0].value), v)
		case IF:
			print_comment("\033[92m conditional IF: \033[39m", inter.current)
			cond := inter.evaluate_set(inter.current.left_side, "check condition")
			print_comment("result from conditional IF:", cond.val.(bool), ", line : ", inter.current.line)
			did_break, did_return := inter.run_decision(cond.val.(bool))
			if did_break || did_return {
				return inter.variables, inter.functions, did_break, did_return
			}
		case ELSEIF:
			print_comment("\033[92m conditional ELSEIF:\033[39m", inter.current)
			cond := inter.evaluate_set(inter.current.left_side, "check condition")
			print_comment("result from conditional ELSEIF:", cond.val.(bool), ", line: ", inter.current.line)
			did_break, did_return := inter.run_decision(cond.val.(bool))
			if did_break || did_return {
				return inter.variables, inter.functions, did_break, did_return
			}
		case ELSE:
		case WHILE:
			print_comment("\033[92m WHILE loop: \033[39m", inter.current)
			save_position := inter.index
			cond := inter.evaluate_set(inter.current.left_side, "check condition")
			print_comment("result from WHILE loop:", cond.val.(bool), ", line : ", inter.current.line)
			if cond.val.(bool) {
				exec_commands := []Command{}
				save_depth := inter.current.depth
				for inter.advance().depth != save_depth {
					exec_commands = append(exec_commands, inter.current)
				}
				newinter := new_interpreter(exec_commands, inter.variables, inter.functions)
				return_vars, return_functions, break_call, is_returned := newinter.call_run()
				print_comment("\033[95mIS RETURN -\033[39m", is_returned)
				if is_returned {
					return return_vars, return_functions, break_call, is_returned
				}

				merge_maps(&inter.variables, return_vars)
				merge_maps(&inter.functions, return_functions)

				print_comment("\033[95m DID BREAK ----------------", break_call, "\033[39m")
				if break_call {
					for inter.peek().depth != save_depth {
						inter.advance()
					}
					break
				}
				print_comment("\033[95mSAVE POSITION------------------\033[39m", save_position)
				inter.move_index(save_position - 1) // -1 for advance method after this
			} else {
				save_depth := inter.current.depth
				for inter.peek().depth != save_depth {
					inter.advance()
				}
			}
		case BREAK:
			return inter.variables, inter.functions, true, false
		case FUNCTION:
			save_depth := inter.current.depth
			save_line := inter.current.line
			func_name := inter.current.left_side[0].value.(string)
			func_args := inter.current.right_side
			func_commands := []Command{}
			for inter.advance().depth != save_depth {
				func_commands = append(func_commands, inter.current)
			}
			print_comment("\033[95m NEW FUNCTION = \033[39m")
			inter.functions[func_name] = Function{save_line, func_args, func_commands}
			print_function(func_name, inter.functions[func_name])
		case RETURN:
			ret := inter.evaluate_set(inter.current.left_side, "return statement")
			print_comment("RETURN", inter.current.left_side, "RET", ret)
			return map[string]Variable{"RETURN_FROM_CALL": ret}, map[string]Function{}, true, true
		case EOF:
		}
		inter.advance()
	}
	print_comment("\033[91m End interpretation -----------------")
	print_comment("return variables = \033[39m", inter.variables)
	return inter.variables, inter.functions, false, false
}
