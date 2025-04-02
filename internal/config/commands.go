package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/frankielb/gator/internal/database"
	"github.com/frankielb/gator/internal/rss"
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

func HandlerAgg(s *State, cmd Command) error {
	w := "https://www.wagslane.dev/index.xml"
	rssFeed, err := rss.FetchFeed(context.Background(), w)
	if err != nil {
		return fmt.Errorf("error fetching feed: %v", err)
	}
	fmt.Printf("%+v\n", rssFeed)
	return nil

}

func HandlerAddFeed(s *State, cmd Command) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("no name or URL given")
	}
	name := cmd.Args[0]
	url := cmd.Args[1]
	currentUser := s.CurrentConfig.CurrentUserName
	user, err := s.Db.GetUser(context.Background(), currentUser)
	if err != nil {
		return fmt.Errorf("error finding user id: %v", err)
	}
	userID := user.ID
	now := time.Now()
	feedID := uuid.New()

	feed, err := s.Db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        feedID,
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
		Url:       url,
		UserID:    userID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}
	fmt.Printf("Feed created successfully:\n%v", feed)
	return nil
}

func HandlerFeeds(s *State, cmd Command) error {
	feeds, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting feeds: %v", err)
	}
	for _, feed := range feeds {
		fmt.Printf("Feed: %v\n -URL: %v\n -Username: %v\n\n", feed.Name, feed.Url, feed.Username)
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
