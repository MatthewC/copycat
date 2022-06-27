package main

import (
	"context"
	"fmt"
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
		if e == env || env == "sample" {
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
