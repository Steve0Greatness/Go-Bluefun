package main

import (
	"fmt"
	"os"
)

func main() {
	variables := map[interface{}]interface{}{}
	arrays := map[interface{}]interface{}{}
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Looks like the command you attempted to run seems to be mal-formed. Here's what your command should look like:")
		fmt.Printf("%s <path/to/file.bluefun>\n", os.Args[1])
	}
}
