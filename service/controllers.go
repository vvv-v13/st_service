package main

import (
	"github.com/go-ozzo/ozzo-routing"
	"log"
	"net/http"
	"strconv"
)

func announceTournamentController(c *routing.Context, service Service) error {
	id := c.Query("tournamentId")
	d := c.Query("deposit")

	if id == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "tournamentId is requred")
	}

	if d == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "deposit is requred")
	}

	tournament, err := strconv.ParseInt(id, 10, 64)

	if err != nil {
		log.Println("AnnounceTournament:", err)
		return routing.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	deposit, err := strconv.ParseInt(d, 10, 64)
	if err != nil {
		log.Println("AnnounceTournament:", err)
		return routing.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = service.AnnounceTournament(tournament, deposit)
	if err != nil {
		log.Println("AnnounceTournament:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Write(map[string]string{})
}

func resetDBController(c *routing.Context, service Service) error {
	err := service.ResetDB()
	if err != nil {
		log.Println("ResetDB:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Write(map[string]string{})
}
