package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(Warn("At least one arguments are needed"))
		help()
		os.Exit(1)
	}

	args := os.Args[1:]

	switch args[0] {
	case "configure":
		configure()
	case "list":
		list()
	case "download":
		twoArgs(args)

		name := args[1]
		download(name)
	case "upload":
		twoArgs(args)

		name := args[1]
		upload(name)
	case "help":
		help()
	default:
		fmt.Println(Warn("Not a valid option."))
		help()
	}

}

func twoArgs(args []string) {
	if len(args) != 2 {
		fmt.Println(Warn("At least two arguments are needed"))
		help()
		os.Exit(1)
	}
}
