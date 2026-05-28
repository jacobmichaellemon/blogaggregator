package main

import (
	"context"
	"database/sql"
	"fmt"
	"main/internal/config"
	"main/internal/database"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"

	config, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config file %s", err)
	}

	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)
	if err != nil {
		fmt.Printf("Error connecting to the database: %s", err)
	}

	appState := state{cfg: &config, db: dbQueries}
	cmds := commands{availibleCommands: map[string]func(*state, command) error{}}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerListUsers)
	cmds.register("agg", handlerAggregator)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerListFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))

	if len(os.Args) < 2 {
		err := fmt.Errorf("no arguments passed, exiting")
		fmt.Println(err)
		os.Exit(1)
	}
	cmd := command{name: os.Args[1], args: os.Args[2:]}
	err = cmds.run(&appState, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type state struct {
	cfg *config.Config
	db  *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	availibleCommands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	runCommand, ok := c.availibleCommands[cmd.name]
	if !ok {
		err := fmt.Errorf("%s command does not exist", cmd.name)
		return err
	}

	return runCommand(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.availibleCommands[name] = f
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	fmt.Printf("Fetching Feed: %s\n", feed.Name)
	if err != nil {
		return err
	}

	s.db.MarkFeedFetched(context.Background(), feed.ID)
	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _, value := range rssFeed.Channel.Item {
		fmt.Printf("* %s\n", value.Title)
	}
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if cmd.name == "login" {
		if len(cmd.args) == 0 {
			err := fmt.Errorf("login command requires at least one argument")
			return err
		}
		user, err := s.db.GetUser(context.Background(), cmd.args[0])
		if err != nil {
			return err
		}
		err = s.cfg.SetUser(user.Name)
		fmt.Printf("The user has been set to: %s", user.Name)
	}
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if cmd.name == "register" {
		if len(cmd.args) == 0 {
			err := fmt.Errorf("register command requires at least one argument: name \n")
			return err
		}
		params := database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0]}
		user, err := s.db.CreateUser(context.Background(), params)
		if err != nil {
			return err
		}
		err = s.cfg.SetUser(user.Name)
		if err != nil {
			return err
		}
		fmt.Printf("The user has been set to: %s \n", user.Name)
	}
	return nil
}

func handlerListUsers(s *state, cmd command) error {
	if cmd.name == "users" {
		users, err := s.db.GetUsers(context.Background())
		if err != nil {
			return err
		}
		for _, value := range users {
			if value.Name == s.cfg.CurrentUserName {
				fmt.Printf("* %s (current)\n", value.Name)
			} else {
				fmt.Printf("* %s\n", value.Name)
			}
		}
	}
	return nil
}

func handlerAggregator(s *state, cmd command) error {
	if cmd.name == "agg" {
		if len(cmd.args) == 0 {
			err := fmt.Errorf("agg command requires at least one argument: time_between_reqs (1s, 1m, 1h, etc)")
			return err
		}
		timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
		if err != nil {
			return err
		}
		ticker := time.NewTicker(timeBetweenRequests)
		fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)
		for ; ; <-ticker.C {
			scrapeFeeds(s)
		}

	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if cmd.name == "addfeed" {
		if len(cmd.args) <= 1 {
			err := fmt.Errorf("add feed command requires at least 2 argument")
			return err
		}
		params := database.CreateFeedParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0], Url: cmd.args[1], UserID: user.ID}
		feed, err := s.db.CreateFeed(context.Background(), params)
		if err != nil {
			return err
		}
		fmt.Printf("New feed created for %s with name: %s and url: %s\n", s.cfg.CurrentUserName, feed.Name, feed.Url)

		follow_feed_params := database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: feed.ID}
		feed_follow, err := s.db.CreateFeedFollow(context.Background(), follow_feed_params)
		if err != nil {
			return err
		}
		fmt.Printf("A feed: %s has been followed by: %s\n", feed_follow.FeedName, feed_follow.UserName)
	}
	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	if cmd.name == "feeds" {
		feeds, err := s.db.GetFeeds(context.Background())
		if err != nil {
			return err
		}
		for _, value := range feeds {
			feedCreatedByName, err := s.db.GetUserNameById(context.Background(), value.UserID)
			if err != nil {
				return err
			}
			fmt.Println("-- Feed Information --")
			fmt.Printf("Feed Name: %s \n", value.Name)
			fmt.Printf("Feed URL : %s \n", value.Url)
			fmt.Printf("Created By:  %s \n", feedCreatedByName)
		}
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if cmd.name == "follow" {
		if len(cmd.args) == 0 {
			err := fmt.Errorf("follow command requires at least one argument: url")
			return err
		}

		feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
		if err != nil {
			return err
		}

		feed_follow_params := database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: feed.ID}
		feed_follow, err := s.db.CreateFeedFollow(context.Background(), feed_follow_params)
		if err != nil {
			return err
		}
		fmt.Printf("A feed: %s has been followed by: %s \n", feed_follow.FeedName, feed_follow.UserName)
	}
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if cmd.name == "following" {

		feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
		if err != nil {
			return err
		}

		fmt.Printf("%s is currently following these feeds:\n", s.cfg.CurrentUserName)
		for _, value := range feeds {
			fmt.Printf("* %s \n", value.FeedName)
		}
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if cmd.name == "unfollow" {
		if len(cmd.args) == 0 {
			err := fmt.Errorf("unfollow command requires at least one argument: url\n")
			return err
		}
		feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
		if err != nil {
			return err
		}
		deleteParams := database.DeleteFeedFollowRecordByUrlParams{FeedID: feed.ID, UserID: user.ID}
		s.db.DeleteFeedFollowRecordByUrl(context.Background(), deleteParams)
		fmt.Printf("Deleted feed: %s from %s follow feed \n", feed.Url, user.Name)
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func handlerReset(s *state, cmd command) error {
	if cmd.name == "reset" {
		err := s.db.DeleteUsers(context.Background())
		if err != nil {
			return err
		}
		fmt.Print("The users table has been reset \n")
	}
	return nil
}
