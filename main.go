/*
copycat is an environment file manager.
The program allows for ".env" files to be uploaded, this creates an
"environment". With an environment created, files can be uploaded, and
associated with that given environment. Files can then later be downloaded.
Any S3-compliant system (i.e., Amazon S3, MinIO, CloudFlare R2) can be used to
store the files, provided you can create an access key and secret for a
specified bucket.

CopyCat now also supports profiles. By default, the "default" profile is used.
Profiles allow for multiple configurations to be created, and later referenced.

Usage:

	copycat [-profile <name>] <command>

The commands are:

	help
		Prints out the help message
	list
		Lists the environments which have been uploaded
	download <environment>
		Downloads a given .env file corresponding to the environment name
	upload <environment>
		Uploads a given .env file
	files <sub-command>
		See below.

As of now, copycat expects the file ".env" to exist, and that is the file it
will automatically upload. Once an environment is created
(using the upload command), CopyCat will also allow files to be uploaded. File
management is handled via the following sub-commands:

	help
		Prints out the files help message
	<environment> list
		Lists the files available in a given environment
	<environment> upload <file name> [upload name]
		Uploads the specified file under the given environment
	<environment> download <file name> [download name]
		Downloads the specified file, allowing for it's name to be
		overwritten
*/
package main

import (
	"flag"
	"fmt"
	"os"
)

const version string = "v1.5.0"

var VersionHost string
var VersionLog string

// Main function routine, serves as main entry point.
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

	// Case on passed arguments
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
