package main

import (
	"github.com/go-ozzo/ozzo-dbx"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	dsn := os.Getenv("SQL_DB")

	// PostgreSQL
	db, err := dbx.MustOpen("postgres", dsn)
	if err != nil {
		log.Fatal(err)
		log.Println("Connection to DB failed, aborting...")
	}
	defer db.Close()

	// Exit with return code 0 on kill.
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM)
	go func() {
		<-done
		os.Exit(0)
	}()

	for {
	}
}
