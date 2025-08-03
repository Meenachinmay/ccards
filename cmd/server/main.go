package main

import (
	"ccards/internal/server"
	"log"
)

func main() {
	bootstrap := server.NewBootstrap()
	if err := bootstrap.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
