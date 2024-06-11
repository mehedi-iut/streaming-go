package main

import (
	"context"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"video_stream/db"
	"video_stream/handlers"
	"video_stream/middlewares"
	"video_stream/routes"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/joho/godotenv"
	// "github.com/aws/aws-sdk-go-v2/aws"
	// "github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	l := log.New(os.Stdout, "video-streaming-api-", log.LstdFlags)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// initialize the database for signup and login
	db.InitDB()

	cor := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		l.Fatalf("Unable to initialize AWS session: %v", err)
	}

	bucketName := "stream258093554996" // replace with your bucket name
	ph := handlers.NewVideoStream(l, cfg, bucketName)

	// ph := handlers.NewVideoStream(l)

	sm := http.NewServeMux()
	protectedRouter := http.NewServeMux()
	// signup
	sm.HandleFunc("POST /signup", routes.Signup)

	// login
	sm.HandleFunc("POST /login", routes.Login)

	protectedRouter.HandleFunc("GET /video.mpd", ph.GetMpd)
	protectedRouter.HandleFunc("GET /video/", ph.GetChunk)
	protectedRouter.HandleFunc("POST /upload", ph.FrontendUpload)
	//protectedRouter.HandleFunc("POST /upload", ph.FrontendUpload)
	sm.Handle("/", middlewares.Authenticate(protectedRouter))

	s := http.Server{
		Addr:         ":9090",
		Handler:      cor.Handler(sm),
		ErrorLog:     l,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		l.Println("Starting server on port 9090")
		err := s.ListenAndServe()
		if err != nil {
			l.Printf("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	sig := <-c
	log.Println("Got signal:", sig)

	// gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(ctx)
}
