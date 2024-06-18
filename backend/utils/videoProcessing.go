package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	//"strings"
)

func ChunkVideo(path, fileName string) {
	//// Extract base filename without extension
	//baseFilename := strings.TrimSuffix(handler.Filename, filepath.Ext(handler.Filename))

	// Create output folder based on base filename
	fmt.Println(fileName)
	videoFileName := fileName + ".mp4"
	//mpdFileName := fileName + ".mpd"
	videoFilePath := filepath.Join(path, videoFileName)
	fmt.Println("videoFileName:", videoFileName)
	fmt.Println("videoFilePath:", videoFilePath)
	outputFolder := filepath.Join(path, fileName)
	fmt.Println("outputFolder:", outputFolder)
	err := os.MkdirAll(outputFolder, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating chunks folder")
		fmt.Println(err)
		//http.Error(rw, "Unable to create chunks folder", http.StatusInternalServerError)
		return
	}

	//cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-map", "0", "-f", "segment", "-segment_time", "10", "-reset_timestamps", "1", filepath.Join(outputFolder, "chunk_%03d.mp4"))
	log.Println("Running ffmpeg on uploaded video")
	watermarkText := "Mehedi"
	
	cmd := exec.Command("ffmpeg", "-i", videoFilePath, "-vf", fmt.Sprintf("drawtext=text='%s':fontcolor=white:fontsize=24:x=10:y=10", watermarkText), "-map", "0", "-b:v", "2400k", "-s:v", "1920x1080", "-c:v", "libx264", "-f", "dash", filepath.Join(outputFolder, "video.mpd"))
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error processing the file with ffmpeg")
		fmt.Println(err)
		//http.Error(rw, "Unable to process the file with ffmpeg", http.StatusInternalServerError)
		return
	}


	log.Println("Finished processing video file")
}
