## A dirty and quick blog system

Go + Bootstrap + jQuery

Run `./install.sh` to install.

## Requirements

- Go 1.24+
- SQLite (via mattn/go-sqlite3)

## Features

- Yinwang-inspired theme (home + article)
- Tags and tag filter
- Pagination (10 posts per page)
- Syntax highlighting (highlight.js, tomorrow theme)
- Mobile-friendly layout

## TODO

- [x] paginate
- [x] add tag to article (search by tag)
- [x] security (parameterized queries; GORM DAO available)

## NOTE
First-time login will create a record in the database with the given username and password.
