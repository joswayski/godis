package main

import (
	"log"

	"github.com/joswayski/godis/internal/config"
)

func main() {
	log.Println("Godis starting...")

	config.GetConfig()

}
