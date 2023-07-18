package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var variables = map[string]string{}

var arrays = map[string][]string{}

func ArrayHas(needle string, haystack []string) bool {
	for _, straw := range haystack {
		if needle == straw {
			return true
		}
	}
	return false
}

func getVar(expression string) string {
	if strings.HasPrefix(expression, "var:") {
		val, _ := variables[expression[4:]]
		return val
	}
	return expression
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
	willLoop := false
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
		willLoop = true
		commands = commands[1:]
	}
	for line, command := range commands {
		if !runAllowed {
			runAllowed = true
			continue
		}
		command = strings.TrimLeft(command, " ")
		if strings.HasPrefix(command, "# ") || command == "" {
			continue
		}
		tokens := strings.Split(command, " ")

		fmt.Printf("%d\n", len(tokens))
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
		if tokens[0] == "if" && len(tokens)%4 != 0 {
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
			variables[tokens[1]] = getVar(strings.Join(tokens[3:], " "))
		case "defInCase":
			_, ok := variables[tokens[1]]
			if !ok {
				variables[tokens[1]] = getVar(strings.Join(tokens[3:], " "))
			}
		case "createArr":
			arrays[tokens[1]] = []string{}
		case "setArrValue":
			array, ok := arrays[tokens[1]]
			if !ok {
				log.Fatalf("%s is not an array.", tokens[1])
			}
			var number int64
			switch tokens[2] {
			case "next":
				number = int64(len(array)) - 1
			default:
				var err error
				number, err = strconv.ParseInt(tokens[2], 10, 64)
				if err != nil {
					log.Fatal("Cannot take a non-integer as input")
				}
			}
			if len(array)-1 < int(number) {
				newArr := make([]string, int(number)-len(array)+1)
				arrays[tokens[1]] = append(arrays[tokens[1]], newArr...)
			}
			arrays[tokens[1]][number] = strings.Join(tokens[3:], " ")
		case "getArrValue":
			array, ok := arrays[tokens[1]]
			if !ok {
				log.Fatalf("%s is not an array.", tokens[1])
			}
			number, err := strconv.ParseInt(tokens[2], 10, 32)
			if err != nil {
				log.Fatal("Cannot take a non-integer as input")
			}
			variables["res"] = fmt.Sprintf("%v", array[number])
		case "getCharAt":
			number, err := strconv.ParseInt(tokens[2], 10, 32)
			if err != nil {
				log.Fatal("Cannot take a non-integer as input")
			}
			variables["res"] = string([]rune(getVar(tokens[1]))[number])
		case "getStrLength":
			str := getVar(tokens[1])
			variables["res"] = fmt.Sprint(len(str))
		case "joinStr":
			strs := []string{getVar(tokens[1]), getVar(tokens[2])}
			variables["res"] = strings.Join(strs, "")
		case "-":
			val, ok := variables[tokens[1]]
			if !ok {
				log.Fatalf("%s is not a variable", tokens[1])
			}
			number, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Fatalf("%s is not a number", tokens[1])
			}
			variables[tokens[1]] = fmt.Sprint(number - 1)
		case "+":
			val, ok := variables[tokens[1]]
			if !ok {
				log.Fatalf("%s is not a variable", tokens[1])
			}
			number, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Fatalf("%s is not a number", tokens[1])
			}
			variables[tokens[1]] = fmt.Sprint(number + 1)
		case "year":
			current := time.Now()
			variables["res"] = fmt.Sprint(current.Year())
		case "month":
			current := time.Now()
			variables["res"] = fmt.Sprint(current.Month())
		case "date":
			current := time.Now()
			variables["res"] = fmt.Sprint(current.Date())
		case "hour":
			current := time.Now()
			variables["res"] = fmt.Sprint(current.Hour())
		case "minute":
			current := time.Now()
			variables["res"] = fmt.Sprint(current.Minute())
		case "second":
			current := time.Now()
			variables["res"] = fmt.Sprint(current.Second())
		case "random":
			high, err := strconv.ParseFloat(getVar(tokens[2]), 64)
			low, lErr := strconv.ParseFloat(getVar(tokens[1]), 64)
			if lErr != nil || err != nil {
				log.Fatalf("both the high and low for random must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(math.Floor(rand.Float64()*(high-low+1)) + low)
		case "add":
			num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
			num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
			if err1 != nil || err2 != nil {
				log.Fatalf("both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(num1 + num2)
		case "sub":
			num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
			num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
			if err1 != nil || err2 != nil {
				log.Fatalf("both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(num1 - num2)
		case "mul":
			num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
			num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
			if err1 != nil || err2 != nil {
				log.Fatalf("both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(num1 * num2)
		case "div":
			num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
			num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
			if err1 != nil || err2 != nil {
				log.Fatalf("both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(num1 / num2)
		case "stop":
			return
		}
	}
	if willLoop {
		main()
	}
}
