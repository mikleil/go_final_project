package main

import (
	"go_final_project/pkg/api"
	"go_final_project/pkg/db"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	// Инициализируем API обработчики
	api.Init()

	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	// Инициализация базы данных
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	webDir := "./web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	addr := ":" + port
	log.Printf("Server is running. Port: %s. Working directory: %s", port, webDir)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server startup error: %v", err)
	}

}
