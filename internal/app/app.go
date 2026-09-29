package app

import (
	"fmt"
	"os"

	"github.com/eduardnesterov/task-manager/internal/platform/config"
	"github.com/eduardnesterov/task-manager/internal/platform/database"
	"github.com/eduardnesterov/task-manager/internal/platform/server"
	"github.com/eduardnesterov/task-manager/internal/seed"
)

type App struct{}

func New() *App {
	return &App{}
}

func isDebugMode() bool {
	envLocal := os.Getenv("TM_DEBUG")
	return envLocal != "false" && envLocal != "0" && envLocal != ""
}

func (a *App) Run() error {
	debug := isDebugMode()

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := database.Open(cfg, debug)
	if err != nil {
		return err
	}

	if err = database.Migrate(db); err != nil {
		return err
	}

	if err = seed.SeedLaborData(db); err != nil {
		return fmt.Errorf("seed labor data: %w", err)
	}

	handlers, scheduler := newAppHandlers(db)
	go scheduler.Start()

	return server.Run(buildRouter(handlers, debug), debug)
}
