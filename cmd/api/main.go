package main

import (
	"forum/server"
	"log"
)

func main() {
	log.Printf("Server started")

	app := server.NewApp()

	if err := app.Run(); err != nil {
		log.Fatalf("%s", err.Error())
	}
}
