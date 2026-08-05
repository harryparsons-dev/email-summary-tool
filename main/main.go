package main

import (
	"email-summary-tool/server"
	"log"
)

func main() {
	s := server.NewServer()
	if err := s.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
