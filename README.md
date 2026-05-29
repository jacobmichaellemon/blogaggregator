This project is a CLI blog aggregator built in go 1.25.3. The purpose of the project is to be able to fetch and scrape blog posts and 
save them under a user. A user is able to add feeds to follow and aggregate them over an interval of time.


This is a project that I followed along from a training website called Boot.Dev.

The project needs Postgres and Go installed to run the project.

macOS with brew
`brew install postgresql@15`

Linux / WSL (Debian)
`sudo apt update`
`sudo apt install postgresql postgresql-contrib`

## Commands to build the project

`go build` 

`go install`

`.pathtobuildfolder/gator` to run


### Availble gator program CLI commands

`gator register username` registers a user in the program

`gator login username` logs in the requested user assuming there are multiple users

`gator users` lists the users availble in the DB

`gator addfeed "feed title" "url"` adds a feed and follows for the current logged in user

`gator feeds` lists the feeds availble to follow

`gator follow "url"` follows the feed given the url provided

`gator following` lists the followed feeds of the user that is currently logged in

`gator unfollow url` unfllows the the feed given the url provided

`gator browse number_posts_to_browse` shows the requested number of posts

`gator agg time_between_reqs`  aggregate posts from the feeds availble given the time_between_reqs (1s, 1m, 1h, etc)

`gator reset` **RESET DB, USE WITH CAUTION** 