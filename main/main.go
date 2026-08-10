package main

import (
	"email-summary-tool/server"
	"log"
)

func main() {
	s := server.NewServer()
	defer s.Close()

	if err := s.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
