package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/frankielb/gator/internal/database"
	"github.com/google/uuid"
)

type State struct {
	CurrentConfig *Config
	Db            *database.Queries
}

type Command struct {
	Name string
	Args []string
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username given")
	}
	name := cmd.Args[0]

	_, err := s.Db.GetUser(context.Background(), name)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("user '%v' does not exist\n", name)
			os.Exit(1)
		}
		return fmt.Errorf("error checking for existing user '%v'", err)
	}

	if err := s.CurrentConfig.SetUser(name); err != nil {
		return err
	}

	fmt.Printf("User %v set\n", cmd.Args[0])
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no username given")
	}
	name := cmd.Args[0]

	_, err := s.Db.GetUser(context.Background(), name)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("error checking for existing user: %v", err)
		}
	} else {
		fmt.Printf("user '%v' already exists\n", name)
		os.Exit(1)
	}

	newUser := uuid.New()
	now := time.Now()

	//make the new user
	user, err := s.Db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        newUser,
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	if err := s.CurrentConfig.SetUser(name); err != nil {
		return err
	}
	fmt.Printf("Successfully registered user %s \n", user.Name)
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	err := s.Db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error resetting database: %v", err)
	}
	fmt.Println("database reset successfully")
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting users: %v", err)
	}

	for _, user := range users {
		if user == s.CurrentConfig.CurrentUserName {
			fmt.Printf("* %s (current)\n", user)
		} else {
			fmt.Printf("* %s\n", user)
		}
	}
	return nil
}

type Commands struct {
	Handlers map[string]func(*State, Command) error
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	if c.Handlers == nil {
		c.Handlers = make(map[string]func(*State, Command) error)
	}
	c.Handlers[name] = f
}
func (c *Commands) Run(s *State, cmd Command) error {
	if handler, exists := c.Handlers[cmd.Name]; exists {
		return handler(s, cmd)
	} else {
		return fmt.Errorf("Command not found: %s", cmd.Name)
	}

}
