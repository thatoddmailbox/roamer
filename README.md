# roamer [![CI status](https://github.com/thatoddmailbox/roamer/workflows/CI/badge.svg)](https://github.com/thatoddmailbox/roamer/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/thatoddmailbox/roamer.svg)](https://pkg.go.dev/github.com/thatoddmailbox/roamer)

`roamer` is a tool that makes handling database migrations easy. It's inspired by [alembic](https://alembic.sqlalchemy.org) and [golang-migrate](https://github.com/golang-migrate/migrate).

It's available as a command-line tool that can be used with any programming language or framework; however, if you're using Go, you can also embed roamer directly into your program.

> **Note** While the command-line interface is mostly stable, the Go API should not be considered stable quite yet! It's possible that a future release of roamer might change the Go API; however, if that does happen, the breaking change would be released in a new minor version, following the Go module versioning policy.

## Documentation
First, follow the [installation instructions](https://github.com/thatoddmailbox/roamer/wiki/Installation).

Then, you're encouraged to follow along with the [guided example](https://github.com/thatoddmailbox/roamer/wiki/A-guided-example).

### Other articles
* [Migrations](https://github.com/thatoddmailbox/roamer/wiki/Migrations)
* [Offsets](https://github.com/thatoddmailbox/roamer/wiki/Offsets)

Also, if you want to provide instructions for other users on how to set up roamer with your project, you can link to the [Connecting an existing project to your database](https://github.com/thatoddmailbox/roamer/wiki/Connecting-an-existing-project-to-your-database) wiki page.