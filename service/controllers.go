package main

import (
	"github.com/go-ozzo/ozzo-routing"
	"log"
	"net/http"
	"strconv"
)

// Announce tournament specifying the entry deposit Controller
func announceTournamentController(c *routing.Context, service Service) error {
	// Get params
	tournament := c.Query("tournamentId")
	d := c.Query("deposit")

	// Check input params
	if tournament == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "tournamentId is requred")
	}

	if d == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "deposit is requred")
	}

	// deposit must be integer
	deposit, err := strconv.ParseInt(d, 10, 64)
	if err != nil {
		log.Println("AnnounceTournament:", err)
		return routing.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Run AnnounceTournament method of ST service
	err = service.AnnounceTournament(tournament, deposit)
	if err != nil {
		log.Println("AnnounceTournament:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// If no errors response 200 with empty JSON Object
	return c.Write(map[string]string{})
}

// Reset DB to initial state Controller
func resetDBController(c *routing.Context, service Service) error {
	// Run ResetDB method of ST service
	err := service.ResetDB()
	if err != nil {
		log.Println("ResetDB:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// If no errors response 200 with empty JSON Object
	return c.Write(map[string]string{})
}

// Player balance controller
func playerBalanceController(c *routing.Context, service Service) error {
	// Get and check playerId (playerId is required)
	id := c.Query("playerId")
	if id == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "playerId is requred")
	}

	// Run PlayerBalance method of ST service
	player, err := service.PlayerBalance(id)
	if err != nil {
		e := err.Error()
		// If no player id db set status 404
		if e == "sql: no rows in result set" {
			return routing.NewHTTPError(http.StatusNotFound, "not found")
		}
		// For other errors set default status 500
		log.Println("PlayerBalanceDB:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, e)
	}
	// Send result of balance query
	return c.Write(player)
}

// Fund player account Controller
func fundController(c *routing.Context, service Service) error {
	// Get params from request
	id := c.Query("playerId")
	p := c.Query("points")

	// Check params
	// playerId is required
	if id == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "playerId is requred")
	}

	// Points is required
	if p == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "points is requred")
	}

	// Points must be integer
	points, err := strconv.ParseInt(p, 10, 64)

	if err != nil {
		log.Println("Fund:", err)
		return routing.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Points must be greater than 0
	if points <= 0 {
		return routing.NewHTTPError(http.StatusBadRequest, "invalid points")
	}

	// Run Fund method of ST service
	err = service.Fund(id, points)
	if err != nil {
		log.Println("Fund:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// If no errors response 200 with empty JSON Object
	return c.Write(map[string]string{})
}

func takeController(c *routing.Context, service Service) error {
	// Get params from request
	id := c.Query("playerId")
	p := c.Query("points")

	// Check params
	// playerId is required
	if id == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "playerId is requred")
	}

	// Points is required
	if p == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "points is requred")
	}

	// Points must be integer
	points, err := strconv.ParseInt(p, 10, 64)

	if err != nil {
		log.Println("Fund:", err)
		return routing.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Points must be greater than 0
	if points <= 0 {
		return routing.NewHTTPError(http.StatusBadRequest, "invalid points")
	}

	// Run Take method of ST service
	rows, err := service.Take(id, points)
	if err != nil {
		log.Println("Take:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// If no rows updated set status 404, (playerId doesn't exist in database)
	if rows == 0 {
		return routing.NewHTTPError(http.StatusNotFound, "not found")
	}

	// If no errors response 200 with empty JSON Object
	return c.Write(map[string]string{})
}

// Join player into a tournament and is he backed by a set of backers Controller
func joinTournamentController(c *routing.Context, service Service) error {
	// Get params from request
	backers := c.Request.URL.Query()["backerId"]
	playerId := c.Query("playerId")
	tournamentId := c.Query("tournamentId")

	// Check params
	// playerId is required
	if playerId == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "playerId is requred")
	}

	// tournamentId is required
	if tournamentId == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "tournamentId is requred")
	}

	// Run JoinTournament method of ST service
	err := service.JoinTournament(tournamentId, playerId, backers)
	if err != nil {
		e := err.Error()
		// If no rows updated set status 404, (playerId or tournamentId  doesn't exist in database)
		if e == "not found" {
			return routing.NewHTTPError(http.StatusNotFound, e)
		}
		// Response for other errors (set status 500)
		log.Println("joinTournamentController:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, e)
	}

	// If no errors response 200 with empty JSON Object
	return c.Write(map[string]string{})
}

// Structures for parse JSON from resultTournament request
type Winner struct {
	PlayerId string `json:"playerId"`
	Prize    int64  `json:"prize"`
}

type Results struct {
	Winners      []Winner `json:"winners"`
	TournamentId string   `json:"tournamentId"`
}

// Result tournament winners and prizes Controller
func resultTournamentController(c *routing.Context, service Service) error {
	var postData Results

	// Get JSON from POST request
	if err := c.Read(&postData); err != nil {
		log.Println("resultTournamentController:", err)
		return routing.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	// Validate fields in JSON
	// Winners are required
	if len(postData.Winners) == 0 {
		return routing.NewHTTPError(http.StatusBadRequest, "bad request, empty winners")
	}

	// tournamentId is required
	tournamentId := postData.TournamentId
	if tournamentId == "" {
		return routing.NewHTTPError(http.StatusBadRequest, "bad request, empty tournamentId")
	}

	// Run ResultTournament method of ST service
	err := service.ResultTournament(tournamentId, postData.Winners)
	if err != nil {
		e := err.Error()
		// If no rows updated set status 404, (playerId or tournamentId or any backerId doesn't exist in database)
		if e == "not found" {
			return routing.NewHTTPError(http.StatusNotFound, e)
		}
		// Response for other errors (set status 500)
		log.Println("resultTournamentController:", err)
		return routing.NewHTTPError(http.StatusInternalServerError, e)
	}

	// If no errors response 200 with empty JSON Object
	return c.Write(map[string]string{})
}
