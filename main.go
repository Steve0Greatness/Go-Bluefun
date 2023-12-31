package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const versionName string = "PREv0.1.0"
const docsUrl string = "https://github.com/Steve0Greatness/Go-Bluefun/wiki"

var variables = map[string]string{}
var line int = 1
var arrays = map[string][]string{}
var runningLive = false
var runAllowed = true
var newline = regexp.MustCompile("\r?\n")
var willLoop = false

var allowDebug = false
var loopDelay = false

var pathToFileBlueFun = `<path\to\file.bluefun>`

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

func sortForThanOperations(thing1 string, thing2 string) []any {
	sortable := []any{}
	numberA, errorA := strconv.ParseFloat(getVar(thing1), 32)
	numberB, errorB := strconv.ParseFloat(getVar(thing2), 32)
	switch ok := []bool{errorA == nil, errorB == nil}; {
	case ok[0] && ok[1]:
		sortInts := []float64{numberA, numberB}
		sort.Float64s(sortInts)
		sortable = []any{sortInts[0], sortInts[1]}
	case ok[0] && !ok[1]:
		sortable = []any{numberA, thing2}
	case !ok[0] && ok[1]:
		sortable = []any{numberB, thing1}
	case !ok[0] && !ok[1]:
		sortStrs := []string{getVar(thing1), getVar(thing2)}
		sort.Strings(sortStrs)
		sortable = []any{sortStrs[0], sortStrs[1]}
	default:
		showErrorF("Somehow, a condition wasn't covered: %t, %t", ok[0], ok[1])
	}
	return sortable
}

func ifBody(operation string, thing1 string, thing2 string) bool {
	var returned bool = false
	switch operation {
	case "=":
		returned = getVar(thing1) == getVar(thing2)
	case ">":
		sortable := sortForThanOperations(thing1, thing2)
		returned = sortable[0] == getVar(thing1)
	case "<":
		sortable := sortForThanOperations(thing1, thing2)
		returned = sortable[1] == getVar(thing1)
	case ">=":
		sortable := sortForThanOperations(thing1, thing2)
		returned = sortable[0] == getVar(thing1) || getVar(thing1) == getVar(thing2)
	case "<=":
		sortable := sortForThanOperations(thing1, thing2)
		returned = sortable[1] == getVar(thing1) || getVar(thing1) == getVar(thing2)
	case "!=":
		returned = getVar(thing1) != getVar(thing2)
	default:
		showErrorF("%s is an invalid operation.", operation)
	}
	return returned
}

func showDebugF(format string, v ...any) {
	if !allowDebug {
		return
	}
	fmt.Printf(format, v...)
}

func showErrorF(format string, v ...any) {
	if !runningLive {
		fmt.Printf("Line %d: ", line)
	}
	fmt.Printf(format, v...)
	if !runningLive {
		os.Exit(1)
	}
}

func runTokens(tokens []string) error {
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
			errorText := fmt.Sprintf("has unneeded data after the last valid boolean check(shown):\n%s", strings.Join(strings.Split(checks[len(checks)-1], " ")[:2], " "))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
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
			errorText := fmt.Sprintf("cannot wait for a non-number amount of time\n")
			showErrorF(errorText)
			return fmt.Errorf(errorText)
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
			errorText := fmt.Sprintf("%s is not a variable", tokens[1])
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		number, err := strconv.ParseFloat(val, 64)
		if err != nil {
			errorText := fmt.Sprintf("%s is not a number", tokens[1])
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables[tokens[1]] = fmt.Sprint(number - 1)
	case "+":
		val, ok := variables[tokens[1]]
		if !ok {
			errorText := fmt.Sprintf("%s is not a variable", tokens[1])
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		number, err := strconv.ParseFloat(val, 64)
		if err != nil {
			errorText := fmt.Sprintf("%s is not a number", tokens[1])
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables[tokens[1]] = fmt.Sprint(number + 1)
	case "createArr":
		arrays[tokens[1]] = []string{}
	case "setArrValue":
		array, ok := arrays[tokens[1]]
		if !ok {
			errorText := fmt.Sprintf("%s is not an array.", tokens[1])
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		var number int64
		switch tokens[2] {
		case "next":
			number = int64(len(array))
		default:
			var err error
			number, err = strconv.ParseInt(getVar(tokens[2]), 10, 64)
			if err != nil {
				errorText := fmt.Sprintf("Cannot take a non-integer as input, instead got %s", getVar(tokens[2]))
				showErrorF(errorText)
				return fmt.Errorf(errorText)
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
			errorText := fmt.Sprintf("%s is not an array.", tokens[1])
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		number, err := strconv.ParseInt(getVar(tokens[2]), 10, 32)
		if err != nil {
			errorText := fmt.Sprintf("getArrValue takes an integer as a second input, instead got: %s", getVar(tokens[2]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprintf("%v", array[number])
	case "getCharAt":
		number, err := strconv.ParseInt(getVar(tokens[2]), 10, 32)
		if err != nil {
			errorText := fmt.Sprintf("Cannot take a non-integer as input, instead got: %s", getVar(tokens[2]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = string([]rune(getVar(tokens[1]))[number])
	case "getStrLength":
		str := getVar(tokens[1])
		variables["res"] = fmt.Sprint(len(str))
	case "getArrLength":
		arr, ok := arrays[tokens[1]]
		if !ok {
			errorText := fmt.Sprintf("%s is not an array", tokens[1])
			showErrorF(errorText)
			return fmt.Errorf(errorText)
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
			errorText := fmt.Sprintf("Both the high and low for random must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprint(math.Floor(rand.Float64()*(high-low+1)) + low)
	case "add":
		num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
		num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
		if err1 != nil || err2 != nil {
			errorText := fmt.Sprintf("Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprint(num1 + num2)
	case "sub":
		num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
		num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
		if err1 != nil || err2 != nil {
			errorText := fmt.Sprintf("Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprint(num1 - num2)
	case "mul":
		num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
		num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
		if err1 != nil || err2 != nil {
			errorText := fmt.Sprintf("Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprint(num1 * num2)
	case "div":
		num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
		num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
		if err1 != nil || err2 != nil {
			errorText := fmt.Sprintf("Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprint(num1 / num2)
	case "mod":
		num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
		num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
		if err1 != nil || err2 != nil {
			errorText := fmt.Sprintf("Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprint(math.Mod(num1, num2))
	case "pow":
		num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
		num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
		if err1 != nil || err2 != nil {
			errorText := fmt.Sprintf("Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprint(math.Pow(num1, num2))
	case "root":
		num1, err1 := strconv.ParseFloat(getVar(tokens[1]), 64)
		num2, err2 := strconv.ParseFloat(getVar(tokens[2]), 64)
		if err1 != nil || err2 != nil {
			errorText := fmt.Sprintf("Both the 1st number and 2nd for a calculation must be numbers, instead got: %s, %s", getVar(tokens[1]), getVar(tokens[2]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprint(math.Pow(num1, 1/num2))
	case "round":
		num, err := strconv.ParseFloat(getVar(tokens[1]), 64)
		if err != nil {
			errorText := fmt.Sprintf("The number for a calculation must be a number, instead got: %s", getVar(tokens[1]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprint(math.Round(num))
	case "floor":
		num, err := strconv.ParseFloat(getVar(tokens[1]), 64)
		if err != nil {
			errorText := fmt.Sprintf("The number for a calculation must be a number, instead got: %s", getVar(tokens[1]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprint(math.Floor(num))
	case "ceil":
		num, err := strconv.ParseFloat(getVar(tokens[1]), 64)
		if err != nil {
			errorText := fmt.Sprintf("The number for a calculation must be a number, instead got: %s", getVar(tokens[1]))
			showErrorF(errorText)
			return fmt.Errorf(errorText)
		}
		variables["res"] = fmt.Sprint(math.Ceil(num))
	case "stop":
		if willLoop {
			willLoop = false
		}
		return nil
	case "extend":
		return nil
	default:
		errorText := fmt.Sprintf("Unrecognized command: %s", tokens[0])
		showErrorF(errorText)
		return fmt.Errorf(errorText)
	}
	return nil
}

func quickCheck(tokens []string) error {
	if ArrayHas(tokens[0], []string{"breaks", "clear", "year", "month", "date", "hour", "minute", "second"}) && len(tokens) > 1 {
		errorText := fmt.Sprintf("Invalid usage of %s, it takes 0 arguments", tokens[0])
		showErrorF(errorText)
		return fmt.Errorf(errorText)
	}
	if ArrayHas(tokens[0], []string{"def", "defInCase", "setArrValue"}) && len(tokens) < 4 {
		errorText := fmt.Sprintf("Invalid usage of %s, it takes at least 4 arguments", tokens[0])
		showErrorF(errorText)
		return fmt.Errorf(errorText)
	}
	if ArrayHas(tokens[0], []string{"wait", "getStrLength", "createArr"}) && len(tokens) != 2 {
		errorText := fmt.Sprintf("Invalid usage of %s, it takes only 2 arguments", tokens[0])
		showErrorF(errorText)
		return fmt.Errorf(errorText)
	}
	if ArrayHas(tokens[0], []string{"random", "add", "sub", "mul", "div", "getArrValue", "getCharAt", "joinStr"}) && len(tokens) != 3 {
		errorText := fmt.Sprintf("Invalid usage of %s, it takes only 3 arguments", tokens[0])
		showErrorF(errorText)
		return fmt.Errorf(errorText)
	}
	if tokens[0] == "if" && len(tokens)%4 != 0 {
		errorText := fmt.Sprintf("Invalid usage of %s, arguments can only be in groups of 4", tokens[0])
		showErrorF(errorText)
		return fmt.Errorf(errorText)
	}
	if ArrayHas(tokens[0], []string{"loop"}) {
		errorText := fmt.Sprintf("Invalid usage of %s, it can only go on the first line", tokens[0])
		showErrorF(errorText)
		return fmt.Errorf(errorText)
	}
	return nil
}

func liveEnv() {
	fmt.Println(`Running live environment. Run "stop" or press ctrl+c to end the live environment.`)
	input := bufio.NewReader(os.Stdin)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Println("\nEnding BlueFun Live Environment -- Reason: [Keyboard Interrupt]")
			os.Exit(1)
		}
	}()
	for {
		var err error
		var command string
		fmt.Print("BlueFun ~> ")
		command, err = input.ReadString('\n')
		if err != nil {
			showErrorF(fmt.Sprint(err))
			continue
		}
		command = newline.Split(command, -1)[0]
		var tokens = strings.Split(command, " ")
		if tokens[0] == "stop" {
			fmt.Println("Ending BlueFun Live Environment -- Reason: [Stop Command]")
			break
		}
		err = quickCheck(tokens)
		if err != nil {
			continue
		}
		err = runTokens(tokens)
		if err != nil {
			continue
		}
		fmt.Print("\n")
	}
}

func runProgram(program string) {
	commands := newline.Split(program, -1)
	if commands[0] == "loop" {
		willLoop = true
		commands = commands[1:]
	}
	var command string
	extendIf := 0
	for _, command = range commands {
		line += 1
		command = strings.TrimLeft(command, " ")
		if strings.HasPrefix(command, "# ") || command == "" {
			continue
		}
		tokens := strings.Split(command, " ")

		if !runAllowed {
			showDebugF("Stopping execution of line %d: %s\n", line, command)
			if tokens[0] == "extend" {
				var toAdd int64 = 2
				if len(tokens) >= 2 {
					var numError error
					toAdd, numError = strconv.ParseInt(tokens[1], 10, 64)
					if numError != nil {
						showErrorF("Cannot use %s to extend an if statement, it needs to be an integer.", tokens[1])
					}
				}
				showDebugF("Adding %d from %d\n", int(toAdd), extendIf)
				extendIf += int(toAdd) - 1
				continue
			}
			if extendIf > 0 {
				showDebugF("Removing 1 from %d\n", extendIf)
				extendIf -= 1
				continue
			}
			if tokens[0] != "if" {
				runAllowed = true
			}
			continue
		}

		quickCheck(tokens)
		runTokens(tokens)
	}
	if willLoop {
		if loopDelay {
			time.Sleep(100 * time.Millisecond)
		}
		runProgram(program)
	}
}

func main() {
	if ArrayHas(runtime.GOOS, []string{"windows"}) {
		pathToFileBlueFun = "<path/to/file.bluefun>"
	}
	var program string
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Printf(`Looks like the command you attempted to run seems to be mal-formed. Here's what your command should look like:
%s %s
You can also run -help to see a list of commands.`, os.Args[0], pathToFileBlueFun)
		return
	}
	for range args {
		switch args[0] {
		case "-help", "-h", "--help":
			fmt.Printf(`%s [options] %s

*RUN TIME OPTIONS*

-eld  : Add a delay between loops, similar to how it works in the JavaScript version of BlueFun.
-ald  : Allow the interpreter to append it's own debugging <Language Development>

*PROGRAM DETAILS*

-help : print a list of all possible command line arguments
-docs : open the docs webpage
-ver  : print version information
-info : print program details
-live : open the live env, and run commands
`, os.Args[0], pathToFileBlueFun)
			if len(args) == 1 {
				return
			}
			args = args[1:]
		case "-docs":
			OpenBrowser(docsUrl)
			if len(args) == 1 {
				return
			}
			args = args[1:]
		case "-version", "-v", "-ver":
			fmt.Println(versionName)
			if len(args) == 1 {
				return
			}
			args = args[1:]
		case "-info":
			fmt.Print(`Go-BlueFun ~~ a Go(lang) implementation of BlueFun.
Licensed under GNU General Public License v3(https://gnu.org/licenses/gpl-3.0.en.html).
`)
			if len(args) == 1 {
				return
			}
			args = args[1:]
		case "-live":
			runningLive = true
			liveEnv()
			return
		case "-enable-loop-delay", "-eld":
			loopDelay = true
			if len(args) == 1 {
				showErrorF("Cannot set runtime options when no program is provided")
			}
			args = args[1:]
		case "-allow-debug-prints", "-ald":
			allowDebug = true
			if len(args) == 1 {
				showErrorF("Cannot set runtime options when no program is provided")
			}
			args = args[1:]
		default:
			break
		}
	}
	fileData, fileError := os.ReadFile(args[0])
	if fileError != nil {
		log.Fatalf("Failed to read %s\n", args[0])
		return
	}
	program = string(fileData)
	runProgram(program)
}
