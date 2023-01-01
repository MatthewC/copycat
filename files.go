package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
)

// Main files entrypoint. Given an array of arguments, handles calling the
// appropriate sub-function.
func files(args []string) {
	if len(args) < 1 {
		fmt.Println(Warn("At least one argument is needed"))
		help(true)
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		list(true)
	case "help":
		help(true)
	default:
		validEnv(args[0], args[1:])
	}
}

// Checks if an environment exists. Terminates the program if it doesn't.
func validEnv(env string, options []string) {
	envs := list(false)

	for _, e := range envs {
		if e == env {
			handleEnv(env, options)
			return
		}
	}

	log.Fatalln(Fata("Environment not found. Use ") + Teal("copycat files list") + Fata(" to view a list of valid environments."))
}

// Depending on the option provided, and an environment, call the
// appropriate sub-function.
func handleEnv(env string, options []string) {
	switch options[0] {
	case "list":
		listFiles(env)
	case "upload":
		requireArgs(options, 1, false, true)
		fileUpload(env, options[1:])
	case "download":
		requireArgs(options, 1, false, true)
		fileDownload(env, options[1:])
	}
}

// Given an environment, list all the files in that environment.
func listFiles(env string) {
	minioClient, bucket, err := getClient()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	objectCh := minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    env + "_uploads/",
		Recursive: false,
	})

	fmt.Println(White(env + "files:"))

	if len(objectCh) == 0 {
		fmt.Println("... " + Warn("Empty!"))
		return
	}

	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}
		fmt.Println(Teal(strings.Replace(object.Key, env+"_uploads/", "", 1)))
	}
}

// Given an environment, and an array which may contain the following:
//   - 0: file to upload
//   - 1: upload name
//
// upload the specified file.
func fileUpload(env string, args []string) {
	minioClient, bucket, err := getClient()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	ensureBucket(minioClient, bucket)

	uploadName := args[0]

	if len(args) == 2 {
		uploadName = args[1]
	}

	fmt.Print(Teal("Uploading " + args[0] + " as " + uploadName + " under environment " + env + "... "))

	objectName := env + "_uploads/" + uploadName
	filePath := args[0]
	contentType := "text/plain"

	// Upload the env file.
	err = uploadFile(minioClient, objectName, filePath, contentType, bucket)
	if err != nil {
		fmt.Println(Fata("FAILED!"))
		log.Fatalln(err)
	}

	fmt.Println(OK("DONE!"))
}

// Given an environment, and an array which may contain the following:
//   - 0: file to download
//   - 1: download name
//
// download the specified file.
func fileDownload(env string, args []string) {
	minioClient, bucket, err := getClient()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	ensureBucket(minioClient, bucket)

	dlName := args[0]

	if len(args) == 2 {
		dlName = args[1]
	}

	fmt.Print(Teal("Downloading " + args[0] + " from environment " + env + " as " + dlName + "... "))

	object, err := minioClient.GetObject(context.Background(), bucket, env+"_uploads/"+args[0], minio.GetObjectOptions{})
	_, exists := object.Stat()
	if err != nil || exists != nil {
		fmt.Println(Fata("FAILED!"))
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(exists)
		}
		return
	}

	localFile, err := os.Create("./" + dlName)
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
