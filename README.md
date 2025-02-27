# Gator CLI

Gator CLI is a command-line tool for managing RSS feeds and posts. It allows users to register, log in, follow feeds, and browse posts. The tool is built with Go and uses PostgreSQL as the database.

## Prerequisites

Before you begin, ensure you have the following installed on your machine:

- [Go](https://golang.org/doc/install) (version 1.16 or later)
- [PostgreSQL](https://www.postgresql.org/download/)

## Installation

To install the Gator CLI, use the `go install` command:

```sh
go install github.com/yourusername/gator@latest
```
This will install the gator binary to your $GOPATH/bin directory.

## Configuration

Create a configuration file named config.json in the root of your project. 
The configuration file should contain the following information:

{
  "database_url": "postgres://your_user:your_password@your_host:your_port/your_database?sslmode=disable",
  "current_user_name": ""
}

Replace your_user, your_password, your_host, your_port, and your_database with your actual PostgreSQL connection details.

## Running the Program

To run the Gator CLI, navigate to the root of your project and use the go run command:
```sh
go run .
```

You can also build the project and run the binary:
```sh
go build -o gator
./gator
```

## Commands
Here are some of the commands you can run with the Gator CLI:

### Register
Register a new user:
```sh
go run . register your_username
```

### Login
Log in with an existing user:
```sh
go run . login your_username
```

### Add Feed
Add a new feed to follow:
```sh
go run . addfeed "Feed Name" "https://example.com/feed.xml"
```

### Follow Feed
Follow an existing feed by URL:
```sh
go run . follow "https://example.com/feed.xml"
```

### Browse Posts
Browse posts for the current user. You can specify an optional limit parameter. If not provided, the default limit is 2:
```sh
go run . browse 10
```

### Aggregate Feeds
Aggregate feeds periodically. Specify the time interval between requests (e.g., 1m for 1 minute):
```sh
go run . agg 1m
```

### List Users
List all registered users:
```sh
go run . users
```

### Reset
Delete all users and their associated data:
```sh
go run . reset
```

## License
This project is licensed under the MIT License.

