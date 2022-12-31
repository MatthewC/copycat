//go:build !windows

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
)

func configure() {
	fmt.Printf("Setting up COPYCAT Environment\n")

	// This would technically resolve a problem with creating directories in Linux, but
	// since the command isn't available on Windows. Either don't build on windows, or
	// figure out how to not use syscall.Umask.

	// if runtime.GOOS != "windows" {
	// oldMask := syscall.Umask(0)
	// }

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

func help(files bool) {
	if !files {
		fmt.Println(White("CopyCat Client\n"))
		fmt.Println("Usage: copycat [--profile] <command>")
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
