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

func playerBalanceController(c *routing.Context, service Service) error {
	id := c.Query("playerId")

	if id == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "playerId is requred")
	}

	player, err := service.PlayerBalance(id)
	if err != nil {
		e := err.Error()
		if e == "sql: no rows in result set" {
			return routing.NewHTTPError(http.StatusNotFound, "not found")
		}

		log.Println("PlayerBalanceDB:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, e)
	}

	return c.Write(player)
}

func fundController(c *routing.Context, service Service) error {
	id := c.Query("playerId")
	p := c.Query("points")

	if id == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "playerId is requred")
	}

	if p == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "points is requred")
	}

	points, err := strconv.ParseInt(p, 10, 64)

	if err != nil {
		log.Println("Fund:", err)
		return routing.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if points <= 0 {
		return routing.NewHTTPError(http.StatusBadRequest, "invalid points")
	}

	err = service.Fund(id, points)
	if err != nil {
		log.Println("Fund:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Write(map[string]string{})
}

func takeController(c *routing.Context, service Service) error {
	id := c.Query("playerId")
	p := c.Query("points")

	if id == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "playerId is requred")
	}

	if p == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "points is requred")
	}

	points, err := strconv.ParseInt(p, 10, 64)

	if err != nil {
		log.Println("Fund:", err)
		return routing.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if points <= 0 {
		return routing.NewHTTPError(http.StatusBadRequest, "invalid points")
	}

	rows, err := service.Take(id, points)
	if err != nil {
		log.Println("Take:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if rows == 0 {
		return routing.NewHTTPError(http.StatusNotFound, "not found")
	}

	return c.Write(map[string]string{})
}

func joinTournamentController(c *routing.Context, service Service) error {
	backers := c.Request.URL.Query()["backerId"]
	playerId := c.Query("playerId")
	id := c.Query("tournamentId")

	if playerId == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "playerId is requred")
	}

	if id == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "tournamentId is requred")
	}
	tournamentId, err := strconv.ParseInt(id, 10, 64)

	err = service.JoinTournament(tournamentId, playerId, backers)
	if err != nil {
		e := err.Error()
		if e == "not found" {
			return routing.NewHTTPError(http.StatusNotFound, e)
		}
		log.Println("joinTournamentController:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, e)
	}

	return c.Write(map[string]string{})
}

type Winner struct {
	Player string `json:"playerId"`
	Prize  int64  `json:"prize"`
}

type Results struct {
	Winners []Winner `json:"winners"`
}

func resultTournamentController(c *routing.Context, service Service) error {
	var postData Results

	if err := c.Read(&postData); err != nil {
		log.Println("resultTournamentController:", err)
		return routing.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	if len(postData.Winners) == 0 {
		return routing.NewHTTPError(http.StatusBadRequest, "bad request, empty winners")
	}

	err := service.ResultTournament(postData.Winners)
	if err != nil {
		e := err.Error()
		if e == "not found" {
			return routing.NewHTTPError(http.StatusNotFound, e)
		}
		log.Println("resultTournamentController:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, e)
	}

	return c.Write(map[string]string{})
}
