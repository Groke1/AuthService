package main

import (
	"AuthService/config"
	"AuthService/pkg/handler"
	"AuthService/pkg/repository"
	"AuthService/pkg/repository/db/postgres"
	"AuthService/pkg/server"
	"AuthService/pkg/service"
	"AuthService/pkg/service/token"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")

	router := mux.NewRouter()

	db, err := postgres.New(postgres.Config{
		Host:    os.Getenv("DBHost"),
		Port:    os.Getenv("DBPort"),
		User:    os.Getenv("DBUser"),
		Pass:    os.Getenv("DBPass"),
		DBName:  os.Getenv("DBName"),
		SSLMode: os.Getenv("DBMode"),
	})
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.New(db)

	serv := service.New(repo, token.Config{
		TtlAccess:  cfg.Token.TtlAccess,
		TtlRefresh: cfg.Token.TtlRefresh,
		Key:        os.Getenv("TOKEN_KEY"),
	})
	h := handler.New(serv)
	h.InitRoutes(router)

	s := server.New(cfg.Port, router)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
