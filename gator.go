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

func middlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error{
	return func(s *State, cmd Command) error {
		user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
		if err != nil{
			return err
		}
		return handler(s, cmd, user)
	}
}

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Not Enough Args. A username is required.")
	}
	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("User does not exist.\n%w",err)
	} 
	err = s.config.SetUser(user.Name, user.ID)
	if err != nil {
		return fmt.Errorf("Unable to set user config.\n%w",err)
	} else {
		fmt.Printf("User has logged in: %s", user.Name)
		fmt.Println()
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
	s.config.SetUser(user.Name, user.ID)
	fmt.Printf("A new user (%s) was registered.\n", user.Name)
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

func handlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("Not Enough Args. A name and URL are required")
	}
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{ID: uuid.New(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Name: cmd.args[0], Url: cmd.args[1], UserID: user.ID })
	if err != nil{
		return err
	}
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), FeedID: feed.ID, UserID: user.ID })
	if err != nil{
		return err
	}
	fmt.Printf("Added Feed: %s", feed.Name)
	fmt.Println()
	return nil
}

func handlerFeeds(s *State, _ Command) error {

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil{
		return err
	}

	for _, v := range feeds {
		fmt.Printf("- %s (%s) - %s", v.Name, v.Url, v.UserName)
		fmt.Println()
	}
	return nil
}

func handlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Not Enough Args. A URL is required.")
	}
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil{
		return err
	}

	feed_follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), UserID: user.ID, FeedID: feed.ID})
	fmt.Printf("%s followed %s\n",feed_follow.UserName, feed_follow.FeedName)
	return nil
}

func handlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Not Enough Args. A URL is required.")
	}
	feed, err := s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{Url: cmd.args[0], UserID: user.ID})
	if err != nil{
		return err
	}
	fmt.Printf("%s unfollowed %s", user.Name, feed.Name)
	return nil
}

func handlerFollowing(s *State, _ Command, user database.User) error {
	following, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil{
		return err
	}
	fmt.Printf("%s's follows:", s.config.CurrentUserName)
	fmt.Println()
	for _, v := range following {
		fmt.Printf("- %s", v.FeedName)
		fmt.Println()
	}
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