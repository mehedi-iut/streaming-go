package handlers

import (
	"log"
	"net/http"
	"path/filepath"
)

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
