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

	// Get user's home directory
	home, err := os.UserHomeDir()

	if err != nil {
		fmt.Println(Fata("A fatal error occured: "), err)
		os.Exit(1)
	}

	// Check if ~/.config folder exists, if not, create it.
	fmt.Printf("Checking for existing .config folder... ")

	_, folderErr := os.Stat(home + "/.config/")

	if folderErr != nil {
		fmt.Println(Fata("NOT FOUND."))

		fmt.Printf("Creating .config folder (" + home + "/.config/)... ")

		configErr := os.Mkdir(home+"/.config/", 0644)

		if configErr != nil {
			fmt.Println(Fata("FAILED"))

			fmt.Println("Could not create .config directory: ", configErr)
		}

		fmt.Println(OK("CREATED!"))
	} else {
		fmt.Println(OK("FOUND!"))
	}

	// Check for existing .copycat configuration.
	fmt.Printf("Checking for existing .copycat configuration... ")
	if _, err := os.Stat(home + "/.config/.copycat"); err == nil {
		fmt.Println(Warn("EXISTS!"))

		// Ask if they want to overwrite configuartion
		fmt.Print(Info("Delete existing configuration [Y\\n]? "))
		var confirm string
		fmt.Scanln(&confirm)

		if confirm == "Y" || confirm == "y" {
			fmt.Print(Warn("Deleting existing configuration... "))

			// Delete file, so we can overwrite it.
			if e := os.Remove(home + "/.config/.copycat"); e != nil {
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
	fmt.Print(Info("Hostname (e.g., https://google.com): "))
	fmt.Scanln(&host)

	fmt.Printf("Connecting to host... ")

	if err := checkHost(host); err != nil {
		fmt.Println(Fata("FAILED!"))
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println(OK("CONNECTED!"))
	}

	var username string
	fmt.Print(Info("KEY: "))
	fmt.Scanln(&username)

	var password string
	fmt.Print(Info("SECRET: "))
	fmt.Scanln(&password)

	fmt.Printf("Attempting to connect... ")

	minioClient := createClient(host, username, password)

	fmt.Println(OK("DONE!"))

	fmt.Printf("Ensuring copycat-env bucket exists... ")

	if err = ensureBucket(minioClient); err != nil {
		fmt.Println(Fata("FAILED!"))
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(OK("DONE!"))

	fmt.Printf("Creating .copycat config... ")

	createConfig(host, username, password, home+"/.config/")

	fmt.Println(OK("DONE!"))

	fmt.Println(Info("Configuration created & saved successfully!"))

	fmt.Println("\nRun " + OK("copycat help") + " to see a list of available commands!")
}

func list() {
	minioClient := getClient()

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	objectCh := minioClient.ListObjects(ctx, "copycat-env", minio.ListObjectsOptions{
		Prefix:    "env_",
		Recursive: true,
	})

	fmt.Println(White("Environments:"))

	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}
		fmt.Println(Teal(strings.Replace(object.Key, "env_", "", 1)))
	}
}

func download(key string) {
	minioClient := getClient()

	ensureBucket(minioClient)

	fmt.Print(Teal("Download " + key + " environment as .env... "))

	object, err := minioClient.GetObject(context.Background(), "copycat-env", "env_"+key, minio.GetObjectOptions{})
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
	minioClient := getClient()

	ensureBucket(minioClient)

	fmt.Print(Teal("Uploading .env with key " + key + "... "))

	objectName := "env_" + key
	filePath := "./.env"
	contentType := "text/plain"

	// Upload the zip file with FPutObject
	_, err := minioClient.FPutObject(context.Background(), "copycat-env", objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		fmt.Println(Fata("FAILED!"))
		log.Fatalln(err)
	}

	fmt.Println(OK("DONE!"))
}

func help() {
	fmt.Println(White("CopyCat Client\n"))

	fmt.Println("Usage:")
	fmt.Println("	copycat help")
	fmt.Println("	copycat list")
	fmt.Println("	copycat download <environment>")
	fmt.Println("	copycat upload <environment>")
}
