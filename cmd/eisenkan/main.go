package main

import (
	"log"

	"github.com/rknuus/eisenkan/client/ui"
)

func main() {
	// Create and start the application
	app := ui.NewApplicationRoot()
	if app == nil {
		log.Fatal("Failed to create Application Root")
	}

	// Start the application (this will show a window)
	if err := app.StartApplication(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
