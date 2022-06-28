package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
		fmt.Println(Fata("FAILED!"))

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
		os.Exit(1)
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

func uploadFile(minioClient *minio.Client, objectName string, filePath string, contentType string) error {

	_, err := minioClient.FPutObject(context.Background(), "copycat-env", objectName, filePath, minio.PutObjectOptions{ContentType: contentType})

	return err
}

func requireArgs(args []string, count int, strict bool, files bool) {
	if (strict && len(args) != count) || len(args) < count {
		fmt.Println(Warn("Expected " + strconv.Itoa(count) + " argument(s), got " + strconv.Itoa(len(args))))
		help(files)
		os.Exit(1)
	}
}

func update() {
	// Check if update exists to begin with.

	// TODO

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

}
