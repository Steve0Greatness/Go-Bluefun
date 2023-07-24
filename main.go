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

var variables = map[string]string{}
var line int = 1
var arrays = map[string][]string{}
var runningLive = false
var runAllowed = true

var HelpText string = fmt.Sprintf(`%s [command]/<path/to/file.bluefun
-help : print a list of all possible command line arguments
-docs : open the docs webpage
-live : open the live environment`, os.Args[0])

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
		showErrorF("%s is an invalid operation.", operation)
	}
	return returned
}

func showErrorF(format string, v ...any) {
	if !runningLive {
		fmt.Printf("Line %d", line)
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
		var tokens = strings.Split(strings.Trim(command, "\n"), " ")
		fmt.Print(tokens)
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
	}
}

func main() {
	var program string
	willLoop := false
	newline := regexp.MustCompile("\r?\n")
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Looks like the command you attempted to run seems to be mal-formed. Here's what your command should look like:")
		fmt.Printf("%s <path/to/file.bluefun>\n", os.Args[0])
		fmt.Println("You can also run -help to see a list of commands.")
		return
	}
	switch args[0] {
	case "-help", "-h", "--help":
		fmt.Print(HelpText)
		return
	case "-docs":
		OpenBrowser("https://github.com/Steve0Greatness/Go-Bluefun/wiki")
		return
	case "-live":
		runningLive = true
		liveEnv()
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

		quickCheck(tokens)
		runTokens(tokens)
	}
	if willLoop {
		time.Sleep(100 * time.Millisecond)
		main()
	}
}
