# Gator RSS Aggregator

## Requirements

- GO (1.25.1)
- Postgres

## Install

1. Clone the repository to your machine.
2. Create a config file at your root directory named `.gatorconfig.json` with a single entry named db_url with the connection string to your postgress schema.
   `{"db_url":"postgres://{user}:@localhost:5432/gator?sslmode=disable"}`
3. Run `go install`

## Usage

Commands:
- login
- register
- users
- addfeed
- feeds
- follow
- following
- unfollow
- browse

Background process to aggregate RSS feeds:
- agg

Reset the database: 
- reset
