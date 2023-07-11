package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
)

func ArrayHas(needle interface{}, haystack []interface{}) bool {
	for _, straw := range haystack {
		if needle == straw {
			return true
		}
	}
	return false
}

func main() {
	// variables := map[interface{}]interface{}{}
	// arrays := map[interface{}]interface{}{}
	// ifState := true
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
}
