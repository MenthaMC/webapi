package handlers

import (
	"database/sql"
	"webapi-v2-neo/internal/config"
	"webapi-v2-neo/internal/services"
)

type Handlers struct {
	config   *config.Config
	db       *sql.DB
	services *services.Services
}

func New(cfg *config.Config, database *sql.DB) *Handlers {
	return &Handlers{
		config:   cfg,
		db:       database,
		services: services.New(database),
	}
}