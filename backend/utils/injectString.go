package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

func InjectString(filePath, injectString string) {
	//tempFilePath := "uploads/video/video_temp.mpd"
	folderPath := path.Dir(filePath)
	log.Println("Folder path: ", folderPath)
	tempFilePath := folderPath + "/video_temp.mpd"

	inputFile, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer inputFile.Close()

	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		fmt.Println("Error creating temp file:", err)
		return
	}
	defer tempFile.Close()

	scanner := bufio.NewScanner(inputFile)
	writer := bufio.NewWriter(tempFile)
	baseURL := injectString
	injected := false

	for scanner.Scan() {
		line := scanner.Text()
		writer.WriteString(line + "\n")
		if !injected && strings.Contains(line, "<Period") {
			writer.WriteString(baseURL + "\n")
			injected = true
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	writer.Flush()

	err = os.Rename(tempFilePath, filePath)
	if err != nil {
		fmt.Println("Error renaming temp file:", err)
		return
	}

	fmt.Println("BaseURL successfully injected into MPD file")
}
