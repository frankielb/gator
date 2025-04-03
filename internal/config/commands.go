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
	if len(cmd.Args) < 1 {
		return fmt.Errorf("usage: agg <time between requests>")
	}
	timeBetweenReqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("error setting time: %v", err)
	}

	fmt.Printf("collecting feeds every %v\n", timeBetweenReqs)
	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		if err := ScrapeFeeds(s); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}

}

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("no name or URL given")
	}
	name := cmd.Args[0]
	url := cmd.Args[1]
	/*
		currentUser := s.CurrentConfig.CurrentUserName
		user, err := s.Db.GetUser(context.Background(), currentUser)
		if err != nil {
			return fmt.Errorf("error finding user id: %v", err)
		}
	*/
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
	// edited
	newFollow := uuid.New()
	out, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        newFollow,
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    userID,
		FeedID:    feedID,
	})
	if err != nil {
		return fmt.Errorf("error following feed: %v", err)
	}

	fmt.Printf("Feed created successfully:\n%v\nAt:%v", feed.Name, out.CreatedAt)
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

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no url given")
	}
	url := cmd.Args[0]
	//currentUser := s.CurrentConfig.CurrentUserName
	newFollow := uuid.New()
	now := time.Now()
	/*
		user, err := s.Db.GetUser(context.Background(), currentUser)
		if err != nil {
			return fmt.Errorf("error finding user id: %v", err)
		}
	*/
	userID := user.ID
	feed, err := s.Db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error finding feed id: %v", err)
	}
	feedID := feed.ID
	out, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        newFollow,
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    userID,
		FeedID:    feedID,
	})
	if err != nil {
		return fmt.Errorf("error creating follow: %v", err)
	}
	fmt.Printf("Feed name: %v\nUser name: %v\n", out.FeedName, out.UserName)
	return nil

}

func HandlerFollowing(s *State, cmd Command, user database.User) error {
	/*
		currentUser := s.CurrentConfig.CurrentUserName
		user, err := s.Db.GetUser(context.Background(), currentUser)
		if err != nil {
			return fmt.Errorf("error finding user id: %v", err)
		}
	*/

	follows, err := s.Db.GetFeedFollowsUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error getting following: %v", err)
	}
	if len(follows) == 0 {
		fmt.Println("You aren't following any feeds yet.")
		return nil
	}
	for _, follow := range follows {
		fmt.Printf("%v\n", follow.FeedName)
	}
	return nil
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("no url given")
	}
	url := cmd.Args[0]
	err := s.Db.DeleteFeedByUser(context.Background(), database.DeleteFeedByUserParams{
		UserID: user.ID,
		Url:    url,
	})
	if err != nil {
		return fmt.Errorf("error unfollowing feed: %v", err)
	}
	return nil
}

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		user, err := s.Db.GetUser(context.Background(), s.CurrentConfig.CurrentUserName)
		if err != nil {
			return fmt.Errorf("error finding user id: %v", err)
		}
		return handler(s, cmd, user)
	}
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

func ScrapeFeeds(s *State) error {
	nextFeed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error finding next feed: %v", err)
	}

	if err := s.Db.MarkFeedFetched(context.Background(), nextFeed.ID); err != nil {
		return fmt.Errorf("error marking feed as fetched: %v", err)
	}
	feed, err := rss.FetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return fmt.Errorf("error getting feed from url: %v", err)
	}
	for _, i := range feed.Channel.Item {
		fmt.Println(i.Title)
	}
	return nil

}
