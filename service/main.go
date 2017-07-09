package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	dsn := os.Getenv("SQL_DB")
        log.Println("DSN:", dsn)

	// Exit with return code 0 on kill.
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM)
	go func() {
		<-done
		os.Exit(0)
	}()

        for {}
}
