package main

import (
	"fmt"
	"os"
)

const version string = "1.1"

func main() {
	if len(os.Args) < 2 {
		fmt.Println(Warn("At least one argument is needed"))
		help(false)
		os.Exit(1)
	}

	args := os.Args[1:]

	switch args[0] {
	case "configure":
		configure()
	case "list":
		list(true)
	case "download":
		twoArgs(args)

		name := args[1]
		download(name)
	case "upload":
		twoArgs(args)

		name := args[1]
		upload(name)
	case "files":
		files(args[1:])
	case "help":
		help(false)
	case "-v":
		fmt.Println(OK("v" + version))
	case "--version":
		fmt.Println(OK("v" + version))
	default:
		fmt.Println(Warn("Not a valid option."))
		help(false)
	}

}

func twoArgs(args []string) {
	if len(args) != 2 {
		fmt.Println(Warn("At least two arguments are needed"))
		help(false)
		os.Exit(1)
	}
}
