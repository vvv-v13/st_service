package main

import (
	"github.com/go-ozzo/ozzo-dbx"
	"github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/access"
	"github.com/go-ozzo/ozzo-routing/content"
	"github.com/go-ozzo/ozzo-routing/fault"
	"github.com/go-ozzo/ozzo-routing/slash"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	// Ozzo-router
	router := routing.New()

	// Middlewares
	router.Use(
		access.Logger(log.Printf),
		slash.Remover(http.StatusMovedPermanently),
		content.TypeNegotiator(content.JSON),
		fault.Recovery(log.Printf),
	)

	// API endpoints
	router.Get(`/`, func(c *routing.Context) error { return view(c) })

	// Http server
	server := &http.Server{
		Addr:           ":8080",
		Handler:        nil,
		ReadTimeout:    100 * time.Second,
		WriteTimeout:   100 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Router
	http.Handle("/", router)

	// Start HTTP server
	log.Println("Server listen on 8080")
	panic(server.ListenAndServe())

}

func view(c *routing.Context) error {
	return c.Write(map[string]string{"result": "ok"})
}
