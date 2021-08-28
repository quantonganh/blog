---
title: "database/sql: never ignore errors"
date: Wed Oct 16 09:50:28 +07 2019
categories:
    - Programming
tags:
    - golang
    - restful
    - sqlite3
---
I'm reading [Building RESTful Web Services with Go](https://www.amazon.com/Building-RESTful-Web-services-gracefully-ebook/dp/B072QB8KL10). And in chapter 4, there is [an example](https://github.com/narenaryan/gorestful/blob/master/chapter4/sqliteFundamentals.go) to play around with SQLite:

```go
db, err := sql.Open("sqlite3", "./books.db")
if err != nil {
	log.Fatal(err)
}

statement, err := db.Prepare("CREATE TABLE IF NOT EXISTS books (id INTERGER PRIMARY KEY, isbn INTEGER, author VARCHAR(64), name VARCHAR(64) NULL)")
if err != nil {
	log.Fatal(err)
} else {
	log.Println("Created table books successfully")
}
statement.Exec()

statement, err = db.Prepare("INSERT INTO books (name, author, isbn) VALUES (?, ?, ?)")
if err != nil {
	log.Fatal(err)
}
statement.Exec("Life is a joke", "The Javna brothers", 123456789)
log.Println("Inserted first book into db")
rows, err := db.Query("SELECT id, name, author FROM books")
var tempBook Book
for rows.Next() {
	rows.Scan(&tempBook.id, &tempBook.name, &tempBook.author)
	log.Printf("ID: %d, Book: %s, Author: %s\n", tempBook.id, tempBook.name, tempBook.author)
}
```

Run this and I got:

```
2019/10/16 09:59:06 Created table books successfully
2019/10/16 09:59:06 Inserted first book into db
2019/10/16 09:59:06 ID: 0, Book: , Author: 
```

Why the ID is zero and name and author is empty?

Checking db from the command line I saw this:

```sql
❯❯❯❯ sqlite3
SQLite version 3.28.0 2019-04-15 14:49:49
Enter ".help" for usage hints.
Connected to a transient in-memory database.
Use ".open FILENAME" to reopen on a persistent database.
sqlite> .open books.db
sqlite> SELECT * FROM books;
|123456789|The Javna brothers|Life is a joke
```

So, there is no `id`. Looking carefully at the schema:

```sql
sqlite> .schema
CREATE TABLE books (id INTERGER PRIMARY KEY, isbn INTEGER, author VARCHAR(64), name VARCHAR(64) NULL);
```

There is a typo: `id` should be `INTEGER`, not `INTERGER`.

How can I prevent this kind of errors?

[Scan](https://golang.org/pkg/database/sql/#Rows.Scan) (or almost all operations with `database/sql`) return an error as the last value. You should always check that. Never ignore them.

```go
if err := rows.Scan(&tempBook.id, &tempBook.name, &tempBook.author); err != nil {
	log.Fatal(err)
}
```

then you can quickly found what the error is:

```
2019/10/16 10:11:12 Created table books successfully
2019/10/16 10:11:12 Inserted first book into db
2019/10/16 10:11:12 sql: Scan error on column index 0, name "id": converting NULL to int is unsupported
```
