package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Used for internal error handling, as the code for error handling can become
// quite repetitive.
func handleError(err error, failed bool, exit bool) {
	if err != nil {

		if failed {
			fmt.Println(Fata("FAILED!"))
		}

		log.Fatal(err)

		if exit {
			os.Exit(1)
		}
	}
}

// Given a host, key, and it's secret, return a new minioClient.
func createClient(host string, key string, secret string) (*minio.Client, error) {
	useSSL := true
	var endpoint string

	if strings.Contains(host, "http://") {
		useSSL = false
		endpoint = strings.Replace(host, "http://", "", 1)
	} else {
		endpoint = strings.Replace(host, "https://", "", 1)
	}

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(key, secret, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return minioClient, nil
}

// Checks whether configuration for a given profile exists. If so, it returns
// the full path to the config file.
func configExists(profile string) (string, bool) {
	// Get user's home directory
	home, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(Fata("A fatal error occurred: %w"), err)
	}

	config := home + "/.config/copycat/" + profile
	if _, err := os.Stat(config); err == nil {
		return config, true
	} else {
		log.Fatal(err)
	}
	return "", false
}

// Gets the "active" client, and returns a minioClient, the active bucket name
// and, if successful, nil. Otherwise, the err will be set.
func getClient() (*minio.Client, string, error) {
	config, configExists := configExists(os.Getenv("COPYCAT_PROFILE"))

	if !configExists {
		fmt.Println("Configuration does not exist. Run " + Info("copycat configure") + " to create configuration file.")
		os.Exit(1)
	}

	godotenv.Load(config)

	client, err := createClient(os.Getenv("HOSTNAME"), os.Getenv("KEY"), os.Getenv("SECRET"))
	if err != nil {
		return nil, "", fmt.Errorf("error creating new client: %w", err)
	}

	return client, os.Getenv("BUCKET"), nil
}

// Given a bucket name, ensure that the bucket exists. Can be modified to
// create the bucket if it isn't found - however, default behavior is to just
// return false if the bucket does not exist.
func ensureBucket(minioClient *minio.Client, bucket string) error {
	found, err := minioClient.BucketExists(context.Background(), bucket)

	if err != nil {
		log.Fatal(err)
		return err
	}

	if found {
		return nil
	}

	// Uncomment for bucket creation if bucket does not exist
	/*
		err = minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{Region: "eu-west"})

		if err != nil {
			log.Fatal(err)
			return err
		}
	*/

	return nil
}

// Given a host, it's access information, the bucket to be used, and a path
// create a new configuration file in the given path. Returns nil if successful,
// otherwise an error.
func createConfig(host string, key string, secret string, bucket string, path string) error {
	config := []byte("HOSTNAME=" + host + "\nKEY=" + key + "\nSECRET=" + secret + "\nBUCKET=" + bucket)

	err := os.WriteFile(path, config, 0644)

	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// Wrapper function used for uploading files given it's storage name and the
// path to store it in.
func uploadFile(minioClient *minio.Client, objectName string, filePath string, contentType string, bucket string) error {
	_, err := minioClient.FPutObject(context.Background(), bucket, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})

	return err
}

// Helper function used to ensure that expected arguments are set, otherwise
// terminates the program.
func requireArgs(args []string, count int, strict bool, files bool) {
	if (strict && len(args) != count) || len(args) < count {
		fmt.Println(Warn("Expected " + strconv.Itoa(count) + " argument(s), got " + strconv.Itoa(len(args))))
		help(files)
		os.Exit(1)
	}
}

// Get's the version of the uploaded binary, and returns that. If successful,
// the version will be returned alongside a nil error value. Otherwise, err
// will be set.
func getVersion() (string, error) {
	resp, err := http.Get(os.Getenv(("VERSION_LOG")))

	if err != nil {
		return "", fmt.Errorf("HTTP error: %w", err)
	}

	response, err := io.ReadAll(resp.Body)

	// Ensure that file fetched actually has a version tag in it.
	if err != nil || response[0] != 118 {
		return "", errors.New("file not fetched properly")
	}

	// See if string starts with a "v" for legacy reasons.
	hostVersion := string(response[:len(response)-1])

	return hostVersion, nil
}
