package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
)

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

func listFiles(env string) {
	minioClient := getClient()

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	objectCh := minioClient.ListObjects(ctx, "copycat-env", minio.ListObjectsOptions{
		Prefix:    env + "/",
		Recursive: false,
	})
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}
		fmt.Println(object)
	}
}

func fileUpload(env string, args []string) {
	minioClient := getClient()

	ensureBucket(minioClient)

	uploadName := args[0]

	if len(args) == 2 {
		uploadName = args[1]
	}

	fmt.Print(Teal("Uploading " + args[0] + " as " + uploadName + " under environment " + env + "... "))

	objectName := env + "_uploads/" + uploadName
	filePath := args[0]
	contentType := "text/plain"

	// Upload the env file.
	err := uploadFile(minioClient, objectName, filePath, contentType)
	if err != nil {
		fmt.Println(Fata("FAILED!"))
		log.Fatalln(err)
	}

	fmt.Println(OK("DONE!"))
}

func fileDownload(env string, args []string) {
	minioClient := getClient()

	ensureBucket(minioClient)

	dlName := args[0]

	if len(args) == 2 {
		dlName = args[1]
	}

	fmt.Print(Teal("Downloading " + args[0] + " from environment " + env + " as " + dlName + "... "))

	object, err := minioClient.GetObject(context.Background(), "copycat-env", env+"_uploads/"+args[0], minio.GetObjectOptions{})
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