package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/minio/minio-go"
)

var (
	dryRun bool
)

// upploadDirectory uploads
func uploadDirectory(mediaType, path string, m *minio.Client) error {
	// bad grammar but eh
	medias, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, media := range medias {
		if !media.IsDir() {
			continue
		}

		path := filepath.Join(path, media.Name())
		log.Printf("uploading %s '%s'", mediaType, media.Name())
		err := uploadFiles(mediaType, media.Name(), path, m)
		if err != nil {
			return err
		}
	}

	return nil
}

func uploadFiles(mediaType, path, filePath string, m *minio.Client) error {
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			return fmt.Errorf("Found unexpected folder in media: %s (%s)", path, file.Name())
		}

		key := filepath.Join(mediaType, path, file.Name())
		log.Printf(" ... uploading file '%s' (key: %s)", file.Name(), key)

		if !dryRun {
			_, err := m.FPutObject("triton-media", key, filepath.Join(filePath, file.Name()), minio.PutObjectOptions{
				NumThreads: uint(runtime.NumCPU()),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Missing arg 1, baseDir")
		os.Exit(1)
	}

	minioEndpoint := os.Getenv("MINIO_URL")
	minioAKID := os.Getenv("MINIO_ACCESS_KEY")
	minioSAK := os.Getenv("MINIO_SECRET_KEY")

	if minioEndpoint == "" {
		minioEndpoint = "127.0.0.1:9000"
	}

	var err error
	dryRun, err = strconv.ParseBool(os.Getenv("DRYRUN"))
	if err != nil {
		dryRun = false
	}

	log.Printf("DRYRUN = %v", dryRun)
	log.Printf("Access Key = %s", minioAKID)
	log.Printf("Secret Key = len(%d)", len(minioSAK))
	log.Printf("Endpoint = %s", minioEndpoint)

	failed := false
	if minioAKID == "" {
		log.Println("Missing $MINIO_ACCESS_KEY")
		failed = true
	}

	if minioSAK == "" {
		log.Println("Missing $MINIO_SECRET_KEY")
		failed = true
	}

	if failed {
		os.Exit(1)
	}

	// Initialize minio client object.
	minioClient, err := minio.New(minioEndpoint, minioAKID, minioSAK, false)
	if err != nil {
		log.Fatalln(err)
	}

	baseDir := os.Args[1]
	log.Printf("reading dir: %s", baseDir)

	dirs, err := ioutil.ReadDir(baseDir)
	if err != nil {
		log.Printf("Failed to read base dir: %v", err)
		os.Exit(1)
	}

	for _, entry := range dirs {
		if !entry.IsDir() {
			continue
		}

		log.Printf("Processing directory: %s", entry.Name())
		err := uploadDirectory(entry.Name(), filepath.Join(baseDir, entry.Name()), minioClient)
		if err != nil {
			log.Printf("Failed to upload directory: %s. %v", entry.Name(), err)
		}
	}
}
