package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/frankielb/gator/internal/config"
	"github.com/frankielb/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	// Load config
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	// Connect to DB

	db, err := sql.Open("postgres", DbURL)
	if err != nil {
		log.Fatalf("Failed to open databse: %v", err)
	}
	defer db.Close()
	// Generate sqlc queries
	dbQueries := database.New(db)

	state := config.State{
		CurrentConfig: &cfg,
		Db:            dbQueries,
	}

	commands := config.Commands{
		Handlers: make(map[string]func(*config.State, config.Command) error),
	}
	commands.Register("login", config.HandlerLogin)
	commands.Register("register", config.HandlerRegister)
	commands.Register("reset", config.HandlerReset)
	commands.Register("users", config.HandlerUsers)
	commands.Register("agg", config.HandlerAgg)

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Error: not enough arguments")
		os.Exit(1)
	}
	cmdName := args[1]
	cmdArgs := args[2:]

	cmd := config.Command{
		Name: cmdName,
		Args: cmdArgs,
	}
	err = commands.Run(&state, cmd)
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}

}
