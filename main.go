package main

import (
	"io"
	"log"
	"os"

	"github.com/gocql/gocqlsh/repl"

	"github.com/chzyer/readline"
	"github.com/gocql/gocql"
)

func connect(addr string) (*gocql.Session, error) {
	// TODO: use a single conn not a session?
	cluster := gocql.NewCluster(addr)
	return cluster.CreateSession()
}

func main() {
	db, err := connect(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r, err := readline.NewEx(&readline.Config{
		Prompt: "gocqlsh> ",
	})

	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	cql := repl.New(db, r)
	if err := cql.Run(); err != nil {
		if err == io.EOF {
			return
		}

		log.Fatal(err)
	}
}
