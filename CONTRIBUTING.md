# Contributing Guidelines

## 1. Install tools

### Install git

```bash
git --version
```

### Install Docker and Docker-compose

```bash
docker --version;
docker-compose --version
```

### Install Go

```bash
go version;
go env
```

### Install Postgres (Optional)

You can also use dockerized Postgres instead of installing one.

### Install Redis (Optional)

You can also use dockerized Redis instead of installing one.

## Run Orchid in staging environment

Creating a `.env` file in project root folder to including all environment variables defined in `docker-compose.yaml`.

For example:

```bash
$ cat .env
LOG_LEVEL="debug"
LOG_DEVELOPMENT=true
EMAIL_FROM="xxx@outlook.com"
EMAIL_USERNAME="xxx@outlook.com"
SMTP_SERVER_HOST="smtp.office365.com"
SMTP_SERVER_PORT="587"
EMAIL_PASSWORD="xxx"
ACCESS_SECRET="xxx"
REFRESH_SECRET="xxx"
```

Then run Orchid with:

```bash
make up
```

To finish Orchid service, run:

```bash
make down
```

More make commands are in the `Makefile`.

## Database migration

When you write code and implement your design and ideas that needs to create more database schemas or change existing ones. You must be careful to add this changes to `pkg/database/schemas.go`.

```go
// Add schemas here whenever you update database schemas.
var migrations = []func(ctx context.Context, tx pgx.Tx) error{
	func(ctx context.Context, tx pgx.Tx) (err error) {
		sql := `
			CREATE TABLE IF NOT EXISTS users(
				id serial PRIMARY KEY,
				username VARCHAR (50) UNIQUE NOT NULL,
				email VARCHAR (300) UNIQUE NOT NULL
		 	);
		`
		_, err = tx.Exec(ctx, sql)
		return err
	},
}
```

Every schema change is a function including SQL appending to this migration slice. And one function may including multiple SQLs.

Orchid run database migrations on starting up every time automatically, and doesn't need external cli tools to manage it manually. Therefor just focus yourself on coding.

## Project layout and business logic

Based on Domain Driven Design, every package owns its own domain maintaining its unique context, and loosely coupled from each other.

For maintainability and extensibility, dependencies (such as database, cache, logger) are explicitly passed by parameters to business logic packages by their generators (`New` functions). Almost no global dependency variable exists.

New business logics are most likely going to under `pkg/apis` with possibly versioning directories such `v0`, `v1` for backward compatibility.

## Working flow

[Gitflow](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow) is recommended. Every time you starting new features, creating a new feature branch based on default branch `dev`, after finishing feature branch, pull a request and review code, then merge it to default branch if no problem.
