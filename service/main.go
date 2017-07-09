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

	// Social Tournament Service
	service := Service{db: db}
	service.Initialize()

	service.Take("P6", 200)
	service.Fund("P1", 300)
	service.Fund("P2", 300)
	service.Fund("P3", 300)
	service.Fund("P4", 500)
	service.Fund("P5", 1000)

	service.AnnounceTournament(1, 1000)

	service.JoinTournament(1, "P1", []string{"P2", "P3"})
	service.JoinTournament(1, "P5", []string{})
	service.ResultTournament("HZ")

	service.Balance("P1")
	service.Balance("P2")
	service.Balance("P3")
	service.Balance("P4")
	service.Balance("P5")

	service.Reset()

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
	router.Get(`/announceTournament`, func(c *routing.Context) error { return announceTournament(c, service) })

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
