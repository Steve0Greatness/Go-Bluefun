package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var variables = map[string]string{}
var line int = 1
var arrays = map[string][]string{}

func ArrayHas(needle string, haystack []string) bool {
	for _, straw := range haystack {
		if needle == straw {
			return true
		}
	}
	return false
}

func OpenBrowser(url string) {
	// https://gist.github.com/hyg/9c4afcd91fe24316cbf0
	var err error = nil

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		fmt.Printf("I'm unable to automatically open your browser for you. Please navigate to this URL: %s", url)
	}
	if err != nil {
		log.Fatal(err)
	}

}

func getVar(expression string) string {
	if strings.HasPrefix(expression, "var:") {
		val, _ := variables[expression[4:]]
		return val
	}
	return expression
}

func ifBody(operation string, thing1 string, thing2 string) bool {
	// TODO: Figure out what's going on with the stupid != operator
	var returned bool = false
	switch operation {
	case "=":
		returned = getVar(thing1) == getVar(thing2)
	case ">":
		sortable := []string{getVar(thing1), getVar(thing2)}
		sort.Strings(sortable)
		returned = sortable[1] == thing1
	case "<":
		sortable := []string{getVar(thing1), getVar(thing2)}
		sort.Strings(sortable)
		returned = sortable[0] == thing1
	case "!=":
		returned = getVar(thing1) != getVar(thing2)
	default:
		log.Fatalf("Line %d: %s is an invalid operation.", line, operation)
	}
	return returned
}

func main() {
	var program string
	runAllowed := true
	willLoop := false
	newline := regexp.MustCompile("\r?\n")
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Printf(`Looks like the command you attempted to run seems to be mal-formed. Here's what your command should look like:
%s <path/to/file.bluefun>
You can also run -help to see a list of commands.`, os.Args[0])
		return
	}
	switch args[0] {
	case "-help", "-h", "--help":
		fmt.Printf(`%s [command]/<path/to/file.bluefun
-help : print a list of all possible command line arguments
-docs : open the docs webpage`, os.Args[0])
		return
	case "-docs":
		OpenBrowser("https://github.com/Steve0Greatness/Go-Bluefun/wiki")
		return
	default:
		fileData, fileError := os.ReadFile(args[0])
		if fileError != nil {
			log.Fatalf("Failed to read %s\n", args[0])
			return
		}
		program = string(fileData)
	}
	commands := newline.Split(program, -1)
	if commands[0] == "loop" {
		willLoop = true
		commands = commands[1:]
	}
	var command string
	for _, command = range commands {
		line += 1
		command = strings.TrimLeft(command, " ")
		if strings.HasPrefix(command, "# ") || command == "" {
			continue
		}
		tokens := strings.Split(command, " ")

		if !runAllowed {
			if tokens[0] != "if" {
				runAllowed = true
			}
			continue
		}

		if ArrayHas(tokens[0], []string{"breaks", "clear", "year", "month", "date", "hour", "minute", "second"}) && len(tokens) > 1 {
			log.Fatalf("Line %d: Invalid usage of %s, it takes 0 arguments", line, tokens[0])
		}
		if ArrayHas(tokens[0], []string{"def", "defInCase", "setArrValue"}) && len(tokens) < 4 {
			log.Fatalf("Line %d: Invalid usage of %s, it takes at least 4 arguments", line, tokens[0])
		}
		if ArrayHas(tokens[0], []string{"wait", "getStrLength", "createArr"}) && len(tokens) != 2 {
			log.Fatalf("Line %d: Invalid usage of %s, it takes only 2 arguments", line, tokens[0])
		}
		if ArrayHas(tokens[0], []string{"random", "add", "sub", "mul", "div", "getArrValue", "getCharAt", "joinStr"}) && len(tokens) != 3 {
			log.Fatalf("Line %d: Invalid usage of %s, it takes only 3 arguments", line, tokens[0])
		}
		if tokens[0] == "if" && len(tokens)%4 != 0 {
			log.Fatalf("Line %d: Invalid usage of %s, arguments can only be in groups of 4", line, tokens[0])
		}
		if ArrayHas(tokens[0], []string{"loop"}) {
			log.Fatalf("Line %d: Invalid usage of %s, it can only go on the first line", line, tokens[0])
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
				log.Fatalf("Line %d has unneeded data after the last valid boolean check(shown):\n%s", line, strings.Join(strings.Split(checks[len(checks)-1], " ")[:2], " "))
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
				log.Fatalf("Line %d: cannot wait for a non-number", line)
			}
			time.Sleep(time.Duration(val) * time.Second)
		case "def":
			variables[tokens[1]] = getVar(strings.Join(tokens[3:], " "))
		case "defInCase":
			_, ok := variables[tokens[1]]
			if !ok {
				variables[tokens[1]] = getVar(strings.Join(tokens[3:], " "))
			}
		case "-":
			val, ok := variables[tokens[1]]
			if !ok {
				log.Fatalf("Line %d: %s is not a variable", line, tokens[1])
			}
			number, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Fatalf("Line %d: %s is not a number", line, tokens[1])
			}
			variables[tokens[1]] = fmt.Sprint(number - 1)
		case "+":
			val, ok := variables[tokens[1]]
			if !ok {
				log.Fatalf("Line %d: %s is not a variable", line, tokens[1])
			}
			number, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Fatalf("Line %d: %s is not a number", line, tokens[1])
			}
			variables[tokens[1]] = fmt.Sprint(number + 1)
		case "createArr":
			arrays[tokens[1]] = []string{}
		case "setArrValue":
			array, ok := arrays[tokens[1]]
			if !ok {
				log.Fatalf("Line %d: %s is not an array.", line, tokens[1])
			}
			var number int64
			switch tokens[2] {
			case "next":
				number = int64(len(array))
			default:
				var err error
				number, err = strconv.ParseInt(getVar(tokens[2]), 10, 64)
				if err != nil {
					fmt.Println(getVar(tokens[2]))
					log.Fatalf("Line %d: Cannot take a non-integer as input", line)
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
				log.Fatalf("Line %d: %s is not an array.", line, tokens[1])
			}
			number, err := strconv.ParseInt(getVar(tokens[2]), 10, 32)
			if err != nil {
				log.Fatalf("Line %d: getArrValue takes an integer as a second output, instead got: %s", line, getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprintf("%v", array[number])
		case "getCharAt":
			number, err := strconv.ParseInt(getVar(tokens[2]), 10, 32)
			if err != nil {
				log.Fatalf("Line %d: Cannot take a non-integer as input", line)
			}
			variables["res"] = string([]rune(getVar(tokens[1]))[number])
		case "getStrLength":
			str := getVar(tokens[1])
			variables["res"] = fmt.Sprint(len(str))
		case "getArrLength":
			arr, ok := arrays[tokens[1]]
			if !ok {
				log.Fatalf("Line %d: %s is not an array", line, tokens[1])
			}
			variables["res"] = fmt.Sprint(len(arr))
		case "joinStr":
			strs := []string{getVar(tokens[1]), getVar(tokens[2])}
			variables["res"] = strings.Join(strs, "")
		case "year":
			current := time.Now()
			variables["res"] = fmt.Sprint(current.Year())
		case "month":
			current := time.Now()
			month := int(current.Month()) - 1
			variables["res"] = fmt.Sprint(month)
		case "date":
			current := time.Now()
			variables["res"] = fmt.Sprint(current.Day())
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
				log.Fatalf("Line %d: Both the high and low for random must be numbers, instead got: %s, %s", line, getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(math.Floor(rand.Float64()*(high-low+1)) + low)
		case "add":
			num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
			num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
			if err1 != nil || err2 != nil {
				log.Fatalf("Line %d: Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", line, getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(num1 + num2)
		case "sub":
			num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
			num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
			if err1 != nil || err2 != nil {
				log.Fatalf("Line %d: Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", line, getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(num1 - num2)
		case "mul":
			num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
			num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
			if err1 != nil || err2 != nil {
				log.Fatalf("Line %d: Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", line, getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(num1 * num2)
		case "div":
			num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
			num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
			if err1 != nil || err2 != nil {
				log.Fatalf("Line %d: Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", line, getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(num1 / num2)
		case "mod":
			num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
			num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
			if err1 != nil || err2 != nil {
				log.Fatalf("Line %d: Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", line, getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(math.Mod(num1, num2))
		case "pow":
			num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
			num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
			if err1 != nil || err2 != nil {
				log.Fatalf("Line %d: Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", line, getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(math.Pow(num1, num2))
		case "root":
			num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
			num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
			if err1 != nil || err2 != nil {
				log.Fatalf("Line %d: Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", line, getVar(tokens[1]), getVar(tokens[2]))
			}
			variables["res"] = fmt.Sprint(math.Pow(num1, 1/num2))
		case "round":
			num, err := strconv.ParseFloat(getVar(tokens[1]), 64)
			if err != nil {
				log.Fatalf("Line %d: The number for a calculation must be a number, instead got: %s", line, getVar(tokens[1]))
			}
			variables["res"] = fmt.Sprint(math.Round(num))
		case "floor":
			num, err := strconv.ParseFloat(getVar(tokens[1]), 64)
			if err != nil {
				log.Fatalf("Line %d: The number for a calculation must be a number, instead got: %s", line, getVar(tokens[1]))
			}
			variables["res"] = fmt.Sprint(math.Floor(num))
		case "ceil":
			num, err := strconv.ParseFloat(getVar(tokens[1]), 64)
			if err != nil {
				log.Fatalf("Line %d: The number for a calculation must be a number, instead got: %s", line, getVar(tokens[1]))
			}
			variables["res"] = fmt.Sprint(math.Ceil(num))
		case "stop":
			return
		default:
			log.Fatalf("Line %d: Unrecognized command: %s", line, tokens[0])
		}
	}
	if willLoop {
		time.Sleep(100 * time.Millisecond)
		main()
	}
}
