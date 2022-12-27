package main

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
)

// TODO: Write test cases.

func TestCreateClient(t *testing.T) {
	err := godotenv.Load()
	if err != nil && os.Getenv("DUMMY_HOST") == "" {
		t.Errorf(Fata("Error loading .env file"))
		return
	}

	var dummyFile string = "dummy_file"

	// Start by creating dummy client
	client, err := createClient(os.Getenv("DUMMY_HOST"), os.Getenv("DUMMY_KEY"), os.Getenv("DUMMY_SECRET"))
	if err != nil || client == nil {
		t.Errorf("Failed connecting to dummy client\n")
		return
	}

	// Create a bucket
	client.MakeBucket(context.Background(), os.Getenv("DUMMY_BUCKET"), minio.MakeBucketOptions{Region: os.Getenv("DUMMY_REGION")})
	if err != nil {
		t.Errorf("Error creating bucket %s", err.Error())
		return
	}

	// Upload a file
	dummyData := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	dummyReader := bytes.NewReader(dummyData)

	_, err = client.PutObject(context.Background(), os.Getenv("DUMMY_BUCKET"), dummyFile, dummyReader, int64(len(dummyData)), minio.PutObjectOptions{})
	if err != nil {
		t.Errorf("Error uploading file: %s", err)
		return
	}

	// Download dummy file
	object, err := client.GetObject(context.Background(), os.Getenv("DUMMY_BUCKET"), dummyFile, minio.GetObjectOptions{})
	if err != nil {
		t.Errorf("Error downloading file: %s", err)
		return
	}
	defer object.Close()

	var downloadBuffer = new(bytes.Buffer)
	downloadBuffer.ReadFrom(object)
	downloadBytes := downloadBuffer.Bytes()

	if !bytes.Equal(dummyData, downloadBytes) {
		t.Errorf("Downloaded data does not equal uploaded data.")
		return
	}

	// Delete dummy file
	err = client.RemoveObject(context.Background(), os.Getenv("DUMMY_BUCKET"), dummyFile, minio.RemoveObjectOptions{})
	if err != nil {
		t.Errorf("Error deleting file: %s", err)
		return
	}

	// Delete the bucket
	err = client.RemoveBucket(context.Background(), os.Getenv("DUMMY_BUCKET"))
	if err != nil {
		t.Errorf("Error deleting bucket: %s", err)
		return
	}
}
