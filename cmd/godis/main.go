package main

import (
	"log"

	"github.com/joswayski/godis/internal/config"
)

func main() {
	log.Print("Yellow")

	config.GetConfig()

	log.Println("Got token")
}
