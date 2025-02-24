package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sushiqiren/gator/internal/config"
	"github.com/sushiqiren/gator/internal/database"

	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	if handler, exists := c.handlers[cmd.name]; exists {
		return handler(s, cmd)
	}
	return fmt.Errorf("unknown command: %s", cmd.name)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("login command expects a username argument")
	}
	username := cmd.args[0]

	// Check if the user exists
	_, err := s.db.GetUserByName(context.Background(), username)
	if err == sql.ErrNoRows {
		return fmt.Errorf("user with name %s does not exist", username)
	} else if err != nil {
		return fmt.Errorf("error checking for existing user: %v", err)
	}

	// Set the current user in the config
	err = s.cfg.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("User has been set to: %s\n", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("register command expects a username argument")
	}
	username := cmd.args[0]

	// Check if the user already exists
	_, err := s.db.GetUserByName(context.Background(), username)
	if err == nil {
		return fmt.Errorf("user with name %s already exists", username)
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("error checking for existing user: %v", err)
	}

	// Create a new user
	newUser := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	}

	createdUser, err := s.db.CreateUser(context.Background(), newUser)
	if err != nil {
		return fmt.Errorf("error creating new user: %v", err)
	}

	// Set the current user in the config
	err = s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("error setting current user: %v", err)
	}

	fmt.Printf("User %s has been created\n", username)
	log.Printf("User created: %+v\n", createdUser)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting all users: %v", err)
	}
	fmt.Println("All users have been deleted")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting users: %v", err)
	}

	currentUser := s.cfg.CurrentUserName
	for _, user := range users {
		if user.Name == currentUser {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	ctx := context.Background()
	feedURL := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(ctx, feedURL)
	if err != nil {
		return fmt.Errorf("error fetching feed: %v", err)
	}

	fmt.Printf("Feed: %+v\n", feed)
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch feed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}

	// Unescape HTML entities in the feed
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Items {
		feed.Channel.Items[i].Title = html.UnescapeString(feed.Channel.Items[i].Title)
		feed.Channel.Items[i].Description = html.UnescapeString(feed.Channel.Items[i].Description)
	}

	return &feed, nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("addfeed command expects a name and a URL argument")
	}
	feedName := cmd.args[0]
	feedURL := cmd.args[1]

	// Get the current user from the config
	currentUser := s.cfg.CurrentUserName
	user, err := s.db.GetUserByName(context.Background(), currentUser)
	if err != nil {
		return fmt.Errorf("error getting current user: %v", err)
	}

	// Create a new feed
	newFeed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedURL,
		UserID:    user.ID,
	}

	createdFeed, err := s.db.CreateFeed(context.Background(), newFeed)
	if err != nil {
		return fmt.Errorf("error creating new feed: %v", err)
	}

	fmt.Printf("Feed created: %+v\n", createdFeed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeedsWithUserNames(context.Background())
	if err != nil {
		return fmt.Errorf("error getting feeds: %v", err)
	}

	for _, feed := range feeds {
		fmt.Printf("Feed Name: %s\n", feed.FeedName)
		fmt.Printf("Feed URL: %s\n", feed.Url)
		fmt.Printf("Created by: %s\n\n", feed.UserName)
	}
	return nil
}

func main() {
	// Read the config file
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Open a connection to the database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer db.Close()

	// Create a new instance of database.Queries
	dbQueries := database.New(db)

	// Create a new state instance
	s := &state{db: dbQueries, cfg: &cfg}

	// Create a new commands instance with an initialized map of handler functions
	cmds := &commands{handlers: make(map[string]func(*state, command) error)}

	// Register the login handler function
	cmds.register("login", handlerLogin)

	// Register the register handler function
	cmds.register("register", handlerRegister)

	// Register the reset handler function
	cmds.register("reset", handlerReset)

	// Register the users handler function
	cmds.register("users", handlerUsers)

	// Register the agg handler function
	cmds.register("agg", handlerAgg)

	// Register the addfeed handler function
	cmds.register("addfeed", handlerAddFeed)

	// Register the feeds handler function
	cmds.register("feeds", handlerFeeds)

	// Use os.Args to get the command-line arguments passed in by the user
	if len(os.Args) < 2 {
		log.Fatalf("Error: expected at least 2 arguments, got %d", len(os.Args))
	}

	// Split the command-line arguments into the command name and the arguments slice
	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]
	cmd := command{name: cmdName, args: cmdArgs}

	// Run the given command and print any errors returned
	if err := cmds.run(s, cmd); err != nil {
		log.Fatalf("Error running command: %v", err)
	}
}
