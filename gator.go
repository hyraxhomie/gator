package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
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

func handlerAgg(s *State, c Command) error {
	if len(c.args) < 1{
		return fmt.Errorf("Not enough args. An interval is required (ex. 1s, 1m, 1h)")
	}
	duration, err := time.ParseDuration(c.args[0])
	if err != nil{
		return err
	}
	ticker := time.NewTicker(duration)
	for ;; <- ticker.C {
		scrapeFeeds(s)
	}
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

func handlerBrowse(s *State, c Command, user database.User) error {
	limit := 2
	if len(c.args) > 0 {
		var err error
		limit, err = strconv.Atoi(c.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{UserID: user.ID, NumPosts: int32(limit)})
	if err != nil{
		return err
	}
	fmt.Printf("Recent feeds for %s", user.Name)
	fmt.Println()
	for _, v := range posts {
		fmt.Printf("- %s: %s",v.Title.String, v.Url.String)
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

func scrapeFeeds(s *State) error{
	next, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil{
		return err
	}
	next, err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{UpdatedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true}, ID: next.ID})
	if err != nil{
		return err
	}
	rssFeed, err := fetchFeed(context.Background(), next.Url)
	if err != nil{
		return err
	}
	for _, v := range rssFeed.Channel.Item {
		title := sql.NullString{String: v.Title, Valid: v.Title != ""}
		url := sql.NullString{String: v.Link, Valid: v.Link != ""}
		description := sql.NullString{String: v.Description, Valid: v.Description != ""}
		pubDate, err := time.Parse(time.RFC1123Z, v.PubDate)
		if err != nil{
			pubDate, err = time.Parse(time.RFC1123, v.PubDate)
			if err != nil{
				pubDate, err = time.Parse(time.RFC3339, v.PubDate)
				if err != nil{
					pubDate = time.Now().UTC()
				}
			}
		}
		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{ID: uuid.New(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Title: title, Url: url, Description: description, PublishedAt: pubDate, FeedID: next.ID})
		if err != nil{
			fmt.Println(fmt.Errorf("Unable to add post %s from %s: %w", v.Title, v.Link, err))
		}
	}
	return nil
}