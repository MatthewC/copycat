package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func createClient(host string, key string, secret string) *minio.Client {
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
		log.Fatalln(err)
	}

	return minioClient
}

func configExists(home string) bool {
	if _, err := os.Stat(home + "/.config/.copycat"); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		log.Fatal(err)
		os.Exit(1)
		return false
	}
}

func getClient() *minio.Client {
	// Get user's home directory
	home, err := os.UserHomeDir()

	if err != nil {
		log.Fatal(Fata("A fatal error occured: "), err)
		os.Exit(1)
	}

	if !configExists(home) {
		fmt.Println("Configuration does not exist. Run " + Info("copycat configure") + " to create configuration file.")
	}

	godotenv.Load(home + "/.config/.copycat")

	return createClient(os.Getenv("HOSTNAME"), os.Getenv("KEY"), os.Getenv("SECRET"))
}

func checkHost(host string) error {
	reqURL := host + "/minio/health/live"

	_, err := http.Get(reqURL)

	return err
}

func ensureBucket(minioClient *minio.Client) error {
	found, err := minioClient.BucketExists(context.Background(), "copycat-env")

	if err != nil {
		log.Fatal(err)
		return err
	}

	if found {
		return nil
	}

	err = minioClient.MakeBucket(context.Background(), "copycat-env", minio.MakeBucketOptions{Region: "eu-west"})

	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func createConfig(host string, key string, secret string, path string) {
	config := []byte("HOSTNAME=" + host + "\nKEY=" + key + "\nSECRET=" + secret + "\n")

	err := os.WriteFile(path+".copycat", config, 0644)

	if err != nil {
		log.Fatal(err)
		return
	}
}
