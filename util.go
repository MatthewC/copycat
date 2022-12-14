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
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

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
	// TODO
	// Currently, this forces the use of a MinIO host, an alternative method
	// is needed to support either AWS or MinIO hosts.
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

func getVersion() (string, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(Fata("Error loading .env file"))
	}
	resp, err := http.Get(os.Getenv(("VERSION_LOG")))

	if err != nil {
		return "", err
	}

	response, err := io.ReadAll(resp.Body)

	// Ensure that file fetched actually has a version tag in it.
	if response[0] != 118 {
		return "", errors.New("file not fetched properly")
	}

	return string(response[:len(response)-1]), err
}

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

	// Get environment variables
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatal(Fata("Error loading .env file"))
	}

	// Download latest version of copycat
	url := os.Getenv("VERSION_HOST") + runtime.GOOS + "-" + runtime.GOARCH

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
	fmt.Printf(Info("Succesfully wrote %d bytes to %s/copycat!\n"), bWritten, installDir)
}
