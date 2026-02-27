package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hyraxhomie/gator/internal/config"
	"github.com/hyraxhomie/gator/internal/database"
	"github.com/hyraxhomie/gator/internal/models"
)

type State struct{
	db *database.Queries
	config *config.Config
}

type Commands struct{
	commands map[string]func(*State, Command) error
}

func (c *Commands) run(s *State, cmd Command) error{
	err := c.commands[cmd.name](s, cmd)
	return err
}

func (c *Commands) register(name string, f func(*State, Command) error) {
	c.commands[name] = f
}

type Command struct{
	name string
	args []string
}

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Not Enough Args. A username is required.")
	}
	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("User does not exist.\n%w",err)
	} 
	err = s.config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("Unable to set user config.\n%w",err)
	} else {
		fmt.Println("User has been set.")
	}
	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Not Enough Args. A username is required.")
	}
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Name: cmd.args[0]})
	if err != nil{
		fmt.Println(fmt.Errorf("An error occurred.\n%w", err))
		os.Exit(1)
	}
	s.config.SetUser(user.Name)
	fmt.Printf("A new (%s) user was registered.\n", user.Name)
	fmt.Println(user)
	return nil
}

func handlerReset(s *State, _ Command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error resetting users table.\n%w",err)
	} else {
		fmt.Println("User table reset.")
	}
	return nil
}

func handlerUsers(s *State, _ Command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Error getting users.\n%w",err)
	} 

	for _, v := range users {
		fmt.Printf("* %s", v.Name)
		if s.config.CurrentUserName == v.Name {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}
	return nil
}

func handlerAgg(s *State, _ Command) error {
	url := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(context.Background(), url)
	if err != nil{
		return err
	}
	fmt.Println(feed)
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*models.RSSFeed, error) {
	var feed *models.RSSFeed
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil{
		return nil, fmt.Errorf("Error creating request for '%s'.\n%w", feedURL, err)
	}
	req.Header.Set("User-Agent", "gator")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil{
		return nil, fmt.Errorf("Error getting feed for '%s'.\n%w", feedURL, err)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil{
		return nil, fmt.Errorf("Error reading body.\n%w", err)
	}

	err = xml.Unmarshal(bytes, &feed)
	if err != nil{
		return nil, fmt.Errorf("Error unmarshalling body.\n%w", err)
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return feed, nil
}