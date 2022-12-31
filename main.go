package main

import (
	"flag"
	"fmt"
	"os"
)

const version string = "v1.5.0"

var VersionHost string
var VersionLog string

func main() {
	if len(os.Args) < 2 {
		fmt.Println(Warn("At least one argument is needed"))
		help(false)
		os.Exit(1)
	}

	// Get default profile, unless profile is explicitly defined.
	profilePtr := flag.String("profile", "default", "profile to be used")
	flag.Parse()
	os.Setenv("COPYCAT_PROFILE", *profilePtr)

	// Load environment variables
	os.Setenv("VERSION_LOG", VersionLog)
	os.Setenv("VERSION_HOST", VersionHost)

	args := flag.Args()

	switch args[0] {
	case "configure":
		configure()

	case "list":
		list(true)

	case "download":
		requireArgs(args, 2, true, false)
		name := args[1]
		download(name)

	case "upload":
		requireArgs(args, 2, true, false)
		name := args[1]
		upload(name)

	case "files":
		files(args[1:])

	case "help":
		help(false)

	case "version", "-v", "--version":
		fmt.Println(OK(version))

	case "version-clean":
		fmt.Println(version)

	case "update":
		update()

	case "reset":
		reset()

	default:
		fmt.Println(Warn("Not a valid option."))
		help(false)
	}
}
