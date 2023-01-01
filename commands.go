package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/minio/minio-go/v7"
)

// Handles setting up the CopyCat environment, prompting to
// overwrite the existing configuration if previously called. Requires no
// arguments, nor does it return anything. Will create ~/.config/ directory
// if it does not exist.
func configure() {
	fmt.Printf("Setting up COPYCAT Environment\n")

	// Get user's home directory
	home, err := os.UserHomeDir()

	if err != nil {
		fmt.Println(Fata("A fatal error occurred: "), err)
		os.Exit(1)
	}

	// Check if ~/.config folder exists, if not, create it.
	fmt.Printf("Checking for existing .config folder... ")

	_, folderErr := os.Stat(home + "/.config/")

	if folderErr != nil {
		fmt.Println(Fata("NOT FOUND."))

		fmt.Printf("Creating .config folder (" + home + "/.config/)... ")

		newDir := home + "/.config/"
		configErr := os.Mkdir(newDir, 0755)

		if configErr != nil {
			fmt.Println(Fata("FAILED"))

			fmt.Println("Could not create .config directory: ", configErr)
			os.Exit(1)
		}

		fmt.Println(OK("CREATED!"))
		os.Chmod(newDir, 0755)
	} else {
		fmt.Println(OK("FOUND!"))
	}

	fmt.Printf("Checking for copycat folder... ")
	_, folderErr = os.Stat(home + "/.config/copycat/")
	if folderErr != nil {
		fmt.Println(Fata("NOT FOUND."))
		fmt.Printf("Creating copycat folder (%s/.config/copycat)... ", home)

		newDir := home + "/.config/copycat/"
		configErr := os.Mkdir(newDir, 0644)
		if configErr != nil {
			log.Fatalf("FAILED! See: %s", configErr)
		} else {
			fmt.Println(OK("CREATED!"))
			os.Chmod(newDir, 0755)
		}
	} else {
		fmt.Println(OK("FOUND!"))
	}

	// Check for existing .copycat configuration.
	fmt.Printf("Checking for existing copycat configuration (profile: %s)... ", os.Getenv("COPYCAT_PROFILE"))

	profileDir := home + "/.config/copycat/" + os.Getenv("COPYCAT_PROFILE")
	if _, err := os.Stat(profileDir); err == nil {
		fmt.Println(Warn("EXISTS!"))

		// Ask if they want to overwrite configuration
		fmt.Print(Info("Delete existing configuration [Y/n]? "))
		var confirm string
		fmt.Scanln(&confirm)

		if confirm == "Y" || confirm == "y" {
			fmt.Print(Warn("Deleting existing configuration... "))

			// Delete file, so we can overwrite it.
			if e := os.Remove(profileDir); e != nil {
				fmt.Println(Fata("FAILED!"))
			}

			fmt.Println(OK("SUCCESS!"))
		} else {
			fmt.Println(Fata("Aborting!"))
			os.Exit(1)
		}

	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Println(OK("DOES NOT EXIST!"))
	} else {
		fmt.Println(Fata("ERROR?"))
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Println("\nConnection Details:")

	var host string
	fmt.Print(Info("Hostname (e.g., https://s3.amazonaws.com): "))
	fmt.Scanln(&host)

	var username string
	fmt.Print(Info("KEY: "))
	fmt.Scanln(&username)

	var password string
	fmt.Print(Info("SECRET: "))
	fmt.Scanln(&password)

	var bucket string
	fmt.Print(Info("BUCKET: "))
	fmt.Scanln(&bucket)

	fmt.Printf("Attempting to connect... ")

	minioClient, err := createClient(host, username, password)
	if err != nil {
		log.Println(Fata("FATAL!"))
		log.Fatalln(err)
		os.Exit(1)
	}
	fmt.Println(OK("DONE!"))

	fmt.Printf("Ensuring \"%s\" bucket exists... ", bucket)

	if err = ensureBucket(minioClient, bucket); err != nil {
		fmt.Println(Fata("FAILED!"))
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(OK("DONE!"))

	fmt.Printf("Creating .copycat config... ")

	createConfig(host, username, password, bucket, profileDir)

	fmt.Println(OK("DONE!"))

	fmt.Println(Info("Configuration created & saved successfully!"))

	fmt.Println("\nRun " + OK("copycat help") + " to see a list of available commands!")
}

// Returns the environments which have been created.
// Takes in a single bool "print" which describes whether it will print
// the environments found as a side-effect.
func list(print bool) []string {
	minioClient, bucket, err := getClient()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	objectCh := minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    "env_",
		Recursive: false,
	})

	if print {
		fmt.Println(White("Environments:"))
	}

	var env []string

	if len(objectCh) == 0 {
		fmt.Println("... " + Warn("Empty!"))
		return env
	}

	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return env
		}
		if print {
			fmt.Println(Teal(strings.Replace(object.Key, "env_", "", 1)))
		}
		env = append(env, strings.Replace(object.Key, "env_", "", 1))
	}

	return env
}

// Given an environment (key), fetches the .env corresponding to that
// environment and downloads it as ".env". Terminates program if environment
// doesn't exist.
func download(key string) {
	minioClient, bucket, err := getClient()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	ensureBucket(minioClient, bucket)

	fmt.Print(Teal("Downloading " + key + " environment as .env... "))

	object, err := minioClient.GetObject(context.Background(), bucket, "env_"+key, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println(Fata("FAILED!"))
		fmt.Println(err)
		return
	}
	localFile, err := os.Create("./.env")
	if err != nil {
		fmt.Println(Fata("FAILED!"))
		fmt.Println(err)
		return
	}
	if _, err = io.Copy(localFile, object); err != nil {
		fmt.Println(Fata("FAILED!"))
		fmt.Println(err)
		return
	}

	fmt.Println(OK("DONE!"))
}

// Creates a new environment and uploads the corresponding ".env" file.
func upload(key string) {
	minioClient, bucket, err := getClient()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	ensureBucket(minioClient, bucket)

	fmt.Print(Teal("Uploading .env with key " + key + "... "))

	objectName := "env_" + key
	filePath := "./.env"
	contentType := "text/plain"

	// Upload the env file.
	err = uploadFile(minioClient, objectName, filePath, contentType, bucket)
	if err != nil {
		fmt.Println(Fata("FAILED!"))
		log.Fatalln(err)
	}

	fmt.Println(OK("DONE!"))
}

// Prints all of CopyCat's functions to standard output.
func help(files bool) {
	if !files {
		fmt.Println(White("CopyCat Client\n"))
		fmt.Println("Usage: copycat [--profile <name>] <command>")
		fmt.Println("Commands:")
		fmt.Println("	help")
		fmt.Println("	list")
		fmt.Println("	download <environment>")
		fmt.Println("	upload <environment>")
		fmt.Println("	files help")

		fmt.Println("	")
	} else {
		fmt.Println(Teal("CopyCat File System"))
		fmt.Println("Usage: copycat [--profile] files <command>")
		fmt.Println("Commands:")
		fmt.Println("	help")
		fmt.Println("	list")
		fmt.Println("	<environment> list")
		fmt.Println("	<environment> upload <file name> [upload name]")
		fmt.Println("	<environment> download <file name> [download name]")
	}
}

// Handles main update routine, which involves checking if a newer version
// has been released, and replacing the CopyCat binary.
func update() {
	fmt.Print("Checking if update exists... ")

	// Check if update exists to begin with.
	ver, err := getVersion()

	if err != nil {
		fmt.Println(Fata("FAILED!"))
		log.Fatal(err)
		os.Exit(1)
	}

	if ver == version {
		fmt.Println(Teal("NONE!"))
		fmt.Println(Warn("You already have the latest version: ") + OK(ver))
		os.Exit(0)
	}

	fmt.Println(OK("FOUND! ") + Fata(version) + " -> " + OK(ver))

	// Get path of where executable is installed.
	fmt.Printf("Locating installation location... ")

	ex, err := os.Executable()

	if err != nil {
		fmt.Println(Fata("FAILED!"))
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Println(OK("FOUND!"))
	installDir := filepath.Dir(ex)

	// Confirm installation directory with user
	var confirm string
	fmt.Print("Confirm installation directory (" + installDir + ") [Y/n]: ")
	fmt.Scanln(&confirm)

	if confirm != "Y" && confirm != "y" {
		fmt.Println(Fata("Aborting!"))
		os.Exit(1)
	}

	// Download latest version of copycat
	url := os.Getenv("VERSION_HOST") + "copycat-" + runtime.GOOS + "-" + runtime.GOARCH

	fmt.Print(Teal("Attempting to create temporary file... "))

	// Store the download in a temporary directory
	dir := os.TempDir() + "/copycat"
	file, err := os.Create(dir)

	handleError(err, true, true)

	fmt.Println(OK("CREATED!"))
	defer file.Close()

	// Download the actual binary
	fmt.Print(Teal("Fetching latest release from " + url + "... "))
	resp, err := http.Get(url)

	handleError(err, true, true)
	defer resp.Body.Close()

	// Write the downloaded binary to the temporary file
	n, err := io.Copy(file, resp.Body)

	handleError(err, true, true)

	fmt.Println(OK("DONE!"))
	fmt.Printf(Teal("Wrote %d bytes to %s\n"), n, dir)

	// With the new binary installed, attempt to replace the currently running binary with the
	// new downloaded binary.
	fmt.Print("Attempting to overwrite binary at " + Info(installDir) + " with binary " + Info(dir) + "... ")

	// Rename current binary
	moveErr := os.Rename(installDir+"/copycat", installDir+"/old_copycat")
	handleError(moveErr, true, true)

	toReplace, errR := os.OpenFile(installDir+"/copycat", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	toInstall, errI := os.ReadFile(dir)

	// If any error occurs, we need to restore the old binary.
	if errR != nil || errI != nil {
		_ = os.Rename(installDir+"/old_copycat", installDir+"/copycat")
	}

	handleError(errR, true, true)
	handleError(errI, true, true)

	defer toReplace.Close()

	// Write the new code to the binary
	bWritten, err := toReplace.Write(toInstall)

	// Similarly, if we can't write to the new binary, we need to replace the old one.
	if err != nil {
		_ = os.Rename(installDir+"/old_copycat", installDir+"/copycat")
		handleError(err, true, true)
	}

	// Clean up any temporary files, or old binaries.
	_ = os.Remove(installDir + "/old_copycat")
	_ = os.Remove(dir)

	fmt.Println(OK("DONE!"))
	fmt.Printf(Info("Successfully wrote %d bytes to %s/copycat!\n"), bWritten, installDir)
}

// Deletes the specified configuration.
func reset() {
	config, configExists := configExists(os.Getenv("COPYCAT_PROFILE"))
	if !configExists {
		fmt.Println(Fata("Configuration does not exist."))
		fmt.Println("Use " + Info("copycat configure") + " to set up your configuration.")
		os.Exit(1)
	}

	// Prompt for user confirmation
	var prompt string
	fmt.Print(Warn("Are you sure you want to delete your configuration [Y/n]? "))
	fmt.Scanln(&prompt)

	if prompt == "N" || prompt == "n" {
		log.Fatalln(Fata("Canceled"))
	}

	// Delete the config file
	err := os.Remove(config)
	if err != nil {
		log.Fatalf("Failed to delete .copycat; see: %s\n", err)
	}
	fmt.Println(OK("Permanently deleted configuration file."))
}
