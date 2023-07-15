package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var variables = map[string]string{}

var arrays = map[string][]interface{}{}

func ArrayHas(needle string, haystack []string) bool {
	for _, straw := range haystack {
		if needle == straw {
			return true
		}
	}
	return false
}

func getVar(expression string) string {
	val, ok := variables[expression[4:]]
	if strings.HasPrefix(expression, "var:") && ok {
		return val
	}
	return expression
}

func getNumber(expression string) float64 {
	floatVal, floatError := strconv.ParseFloat(expression, 64)
	if floatError != nil {
		return floatVal
	}
	return 0
}

func ifBody(operation string, thing1 string, thing2 string) bool {
	switch operation {
	case "=":
		return getVar(thing1) == getVar(thing2)
	case ">":
		sortable := []string{getVar(thing1), getVar(thing2)}
		sort.Strings(sortable)
		return sortable[0] == thing1
	case "<":
		sortable := []string{getVar(thing1), getVar(thing2)}
		sort.Strings(sortable)
		return sortable[1] == thing1
	case "!=":
		return getVar(operation) != getVar(operation)
	default:
		log.Fatalf("%s is an invalid operation.", operation)
	}
	return false
}

func main() {
	// broken := false
	runAllowed := true
	// willLoop := false
	newline := regexp.MustCompile("\r?\n")
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Looks like the command you attempted to run seems to be mal-formed. Here's what your command should look like:")
		fmt.Printf("%s <path/to/file.bluefun>\n", os.Args[0])
		return
	}
	fileData, fileError := os.ReadFile(args[0])
	if fileError != nil {
		log.Fatalf("Failed to read %s\n", args[0])
		return
	}
	program := string(fileData)
	commands := newline.Split(program, -1)
	if commands[0] == "loop" {
		commands = commands[1:]
	}
	for line, command := range commands {
		// if broken {
		// 	break
		// }
		if !runAllowed {
			runAllowed = true
			continue
		}
		command = strings.TrimLeft(command, " ")
		if strings.HasPrefix(command, "# ") || command == "" {
			continue
		}
		tokens := strings.Split(command, " ")

		if ArrayHas(tokens[0], []string{"breaks", "clear", "year", "month", "date", "hour", "minute", "second"}) && len(tokens) > 1 {
			log.Fatalf("Invalid usage of %s, it takes 0 arguments", tokens[0])
		}
		if ArrayHas(tokens[0], []string{"def", "defInCase", "setArrValue"}) && len(tokens) < 4 {
			log.Fatalf("Invalid usage of %s, it takes at least 4 arguments", tokens[0])
		}
		if ArrayHas(tokens[0], []string{"wait", "getStrLength", "createArr"}) && len(tokens) != 2 {
			log.Fatalf("Invalid usage of %s, it takes only 2 arguments", tokens[0])
		}
		if ArrayHas(tokens[0], []string{"random", "add", "sub", "mul", "div", "getArrValue", "getCharAt", "joinStr"}) && len(tokens) != 3 {
			log.Fatalf("Invalid usage of %s, it takes only 3 arguments", tokens[0])
		}
		if tokens[0] == "if" && len(tokens)%4 == 0 {
			log.Fatalf("Invalid usage of %s, arguments can only be in groups of 4", tokens[0])
		}
		if ArrayHas(tokens[0], []string{"loop"}) {
			log.Fatalf("Invalid usage of %s, it can only go on the first line", tokens[0])
		}
		switch tokens[0] {
		case "write":
			fmt.Print(getVar(strings.Join(tokens[1:], " ")))
		case "break":
			fmt.Print("\n")
		case "clear":
			fmt.Print("\033[2J")
		case "ask":
			var input string
			fmt.Print(getVar(strings.Join(tokens[1:], " ")))
			fmt.Scan(&input)
			variables["res"] = input
		case "if":
			checks := strings.Split(strings.Join(tokens[1:], " "), " or ")
			if len(checks) < len(tokens)/4 {
				log.Fatalf("Line #%d has unneeded data after the last valid boolean check(shown):\n%s", line, strings.Join(strings.Split(checks[len(checks)-1], " ")[:2], " "))
			}
			runAllowed = false
			for _, check := range checks {
				ifCheck := strings.Split(check, " ")
				if ifBody(ifCheck[1], ifCheck[0], ifCheck[2]) {
					runAllowed = true
					break
				}
			}
		case "wait":
			val, err := strconv.ParseFloat(tokens[1], 64)
			if err != nil {
				log.Fatalf("")
			}
			time.Sleep(time.Duration(val) * time.Second)
		case "def":
			variables[tokens[1]] = getVar(tokens[2])
		case "createArr":
			arrays[tokens[1]] = []interface{}{}
		case "setArrValue":
		}
	}
}
