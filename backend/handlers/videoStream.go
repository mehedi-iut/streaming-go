package handlers

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"video_stream/routes"
)

const bucket = "stream211125645228"
const folder = "video"
const key = "video/"

type VideoStream struct {
	l *log.Logger
}

func NewVideoStream(l *log.Logger) *VideoStream {
	return &VideoStream{l}
}

func (s VideoStream) GetMpd(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("called for mpd")
	filePath := filepath.Join("video", r.URL.Path)
	http.ServeFile(rw, r, filePath)
}

func (s VideoStream) GetChunk(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("called for chunk")
	segments := r.URL.Path[len("/video/"):]
	chunkPath := filepath.Join("video", segments)
	http.ServeFile(rw, r, chunkPath)
}

func (s VideoStream) ChunkUpload(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("called for chunk upload")
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	// Open the directory
	log.Printf("Reading directory %s", folder)
	files, err := os.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}

	bs := BucketBasics{S3Client: client}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // limit concurrency to 10 uploads at a time

	// Iterate through the files and upload each one
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		wg.Add(1)
		go func(file os.DirEntry) {
			defer wg.Done()
			sem <- struct{}{}        // acquire semaphore
			defer func() { <-sem }() // release semaphore

			filename := filepath.Join(folder, file.Name())
			log.Printf("Reading file %s", filename)

			fileHandle, err := os.Open(filename)
			if err != nil {
				log.Printf("Failed to open file %s: %v", filename, err)
				return
			}
			defer fileHandle.Close()

			fileInfo, err := fileHandle.Stat()
			if err != nil {
				log.Printf("Failed to get file info for %s: %v", filename, err)
				return
			}

			log.Printf("Creating buffer for file %s", filename)
			buffer := make([]byte, fileInfo.Size())

			_, err = fileHandle.Read(buffer)
			if err != nil && err != io.EOF {
				log.Printf("Failed to read file %s: %v", filename, err)
				return
			}

			log.Printf("Uploading file %s to s3", filename)
			key := key + fileInfo.Name()
			err = bs.UploadLargeObject(bucket, key, buffer)
			if err != nil {
				log.Printf("Failed to upload %s to s3: %v", filename, err)
			} else {
				fmt.Printf("Successfully uploaded %s of size %d\n", key, fileInfo.Size())
			}
		}(file)
	}

	wg.Wait()
	s.l.Println("Upload complete")

	// Send response to frontend
	response := map[string]string{"status": "upload completed"}
	routes.ConvertToJsonResponse(rw, response, http.StatusOK)

}
