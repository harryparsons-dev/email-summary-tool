package main

import (
	"email-summary-tool/config"
	"email-summary-tool/routes"
	"email-summary-tool/server"
	"log"
	"os"
)

func main() {
	jwtConfig, err := config.NewJWTConfig(os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	s := server.NewServer(jwtConfig)
	defer s.Close()

	routes.InitializeRoutes(s)

	if err := s.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
