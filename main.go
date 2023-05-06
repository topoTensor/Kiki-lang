package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

func check_err(e error) {
	if e != nil {
		panic(e)
	}
}

func throw_error(msg string, line, column int) {
	panic("\033[91m ERROR: (" + strconv.Itoa(line) + " : " + strconv.Itoa(column) + "). " + msg + " \033[39m")
}

func print_comment(str ...interface{}) {
	if DO_COMMENT {
		fmt.Println(str...)
	}
}
func print_commentf(way string, str ...interface{}) {
	if DO_COMMENT {
		fmt.Printf(way, str...)
	}
}

var DO_COMMENT bool = false

func replace_import_by_data(str string) string {
	// import("test.a", "new.a")
	lines := []string{}
	line := ""
	new_str := ""
	for i, s := range str {
		if s == '\n' {
			lines = append(lines, line)
			line = ""
		} else if i+1 == len(str) {
			line += string(s)
			lines = append(lines, line)
			line = ""
		} else {
			line += string(s)
		}
	}
	for _, l := range lines {
		r := regexp.MustCompile(`import\(`)
		if r.MatchString(l) {
			scan := new_scanner(l)
			lex := scan.lex()
			arguments := []Token{}
			for _, le := range lex {
				if le.typ == STRING {
					arguments = append(arguments, le)
				}
			}
			for _, arg := range arguments {
				import_data, err := os.ReadFile(arg.value.(string))
				check_err(err)
				new_str += "\n" + string(import_data)
			}
		} else {
			new_str += "\n" + l
		}
	}
	print_comment(new_str)
	return new_str
}

func do_imports(file_name string) string {
	file, err := os.Open(file_name)
	check_err(err)
	defer file.Close()
	data, err := os.ReadFile(file_name)
	check_err(err)
	new_data := replace_import_by_data(string(data))
	return new_data
}

func main() {
	if len(os.Args) == 1 {
		panic("provide the source file")
	} else if len(os.Args) > 3 {
		panic("too much arguments for interpreter, need only the name of file and interpretation type (disabled by default, 0 - disabled, 1 - enabled)")
	}
	file_data := do_imports(os.Args[1])
	interpretation_type := os.Args[2]

	if interpretation_type == "1" {
		DO_COMMENT = true
	}

	var scan Scanner = new_scanner(file_data + "\n")
	var parser Parser = new_parser(scan.lex())
	var commands []Command = parser.parse()
	var interpreter Inter = new_interpreter(commands, map[string]Variable{}, map[string]Function{})
	print_comment("\033[91m*************************|||START|||*************************\033[39m")
	variables, functions, breaked_call, return_call := interpreter.call_run()
	for name, f := range functions {
		print_function(name, f)
	}
	print_comment("LAST ITERATION VARIABLES:", variables, ".\ndid break? -", breaked_call, "\ndid return? - ", return_call)
}
