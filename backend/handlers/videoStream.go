package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"video_stream/routes"
	"video_stream/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var fileName *string

type VideoStream struct {
	l        *log.Logger
	s3Client *s3.Client
	bucket   string
}

func NewVideoStream(l *log.Logger, cfg aws.Config, bucket string) *VideoStream {
	s3Client := s3.NewFromConfig(cfg)
	return &VideoStream{l, s3Client, bucket}
}

func (s VideoStream) GetMpd(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("called for mpd")
	key := filepath.Join(*fileName, r.URL.Path)
	// http.ServeFile(rw, r, filePath)
	s.s3GetFile(rw, r, key)
}

func (s VideoStream) GetChunk(rw http.ResponseWriter, r *http.Request) {
	s.l.Println("called for chunk")
	segments := r.URL.Path[len("/video/"):]
	chunkPath := filepath.Join(*fileName, segments)
	// http.ServeFile(rw, r, chunkPath)
	s.s3GetFile(rw, r, chunkPath)
}

func (s VideoStream) s3GetFile(rw http.ResponseWriter, r *http.Request, key string) {
	req, err := s.s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		http.Error(rw, "Failed to get file from S3", http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	// rw.Header().Set("Content-Type", aws.ToString(req.ContentType))
	// rw.Header().Set("Content-Length", fmt.Sprintf("%d", req.ContentLength))
	// io.Copy(rw, req.Body)

	buff, buffErr := io.ReadAll(req.Body)
	if buffErr != nil {
		s.l.Println(buffErr)
		http.Error(rw, "Failed to read body", http.StatusInternalServerError)
		return
	}

	reader := bytes.NewReader(buff)

	http.ServeContent(rw, r, key, time.Now(), reader)
}

func (s VideoStream) ChunkUpload(rw http.ResponseWriter, r *http.Request) {

	s.l.Println("Processing uploaded file....")
	utils.ChunkVideo(os.Getenv("folder"), *fileName)

	bucket := os.Getenv("bucket")
	folderPath := os.Getenv("folder")
	injectString := os.Getenv("injectString")

	// Adding BaseURL in mpd file
	// Open the directory
	filePath := filepath.Join(folderPath, *fileName)
	s.l.Println("File path: ", filePath)
	// mpdFileName := *fileName + ".mpd"
	//mpdTempFileName := *fileName + "_temp" + ".mpd"
	mpdFile := filepath.Join(filePath, "video.mpd")
	//mpdTempFile := filepath.Join(filePath, mpdTempFileName)
	s.l.Println("MPD file path ", mpdFile)
	utils.InjectString(mpdFile, injectString)

	s.l.Println("called for chunk upload")
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	// Open the directory
	//filePath := filepath.Join(folder, *fileName)

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
	uploadFolder := os.Getenv("folder")
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
