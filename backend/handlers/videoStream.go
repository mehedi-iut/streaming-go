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
	"strings"
	"sync"
	"video_stream/routes"
	"video_stream/utils"
)

var fileName *string

const bucket = "streaming767397690733"
const folder = "uploads"

//var key = fileName + "/"

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

	s.l.Println("Processing uploaded file....")
	utils.ChunkVideo("./uploads", *fileName)

	s.l.Println("called for chunk upload")
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	// Open the directory
	filePath := filepath.Join(folder, *fileName)

	log.Printf("Reading directory %s", filePath)
	files, err := os.ReadDir(filePath)
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

			filename := filepath.Join(filePath, file.Name())
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
			s3Folder := *fileName + "/"
			key := s3Folder + fileInfo.Name()
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

func (s VideoStream) FrontendUpload(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("Handling frontend upload")
	r.ParseMultipartForm(32 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		s.l.Printf("Failed to get file from form: %v", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	baseFilename := strings.TrimSuffix(handler.Filename, filepath.Ext(handler.Filename))
	fileName = &baseFilename

	// Create a folder to save the file
	// Create a folder to save the file
	s.l.Println("Creating file from form")
	uploadFolder := "./uploads"
	err = os.MkdirAll(uploadFolder, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating upload folder")
		fmt.Println(err)
		http.Error(rw, "Unable to create upload folder", http.StatusInternalServerError)
		return
	}

	// Create the file
	filePath := filepath.Join(uploadFolder, handler.Filename)
	destFile, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating the file")
		fmt.Println(err)
		http.Error(rw, "Unable to create the file", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// Copy the uploaded file to the destination file
	_, err = io.Copy(destFile, file)
	if err != nil {
		fmt.Println("Error saving the file")
		fmt.Println(err)
		http.Error(rw, "Unable to save the file", http.StatusInternalServerError)
		return
	}

	// Send response to frontend
	response := map[string]string{"status": "upload completed"}
	routes.ConvertToJsonResponse(rw, response, http.StatusOK)

	go s.ChunkUpload(rw, r)
}
