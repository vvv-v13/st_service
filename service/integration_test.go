package main

import (
	"fmt"
        "encoding/json"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
        "strconv"
)

func TestService(t *testing.T) {

	db := initDatabase()
	defer db.Close()
	server := httptest.NewServer(initRouter(db))
	defer server.Close()

	log.Println("Reset DB")
	url := fmt.Sprintf("%s/reset", server.URL)
	req, err := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Reset DB:", err)
		return
	}
	assert.Equal(t, res.StatusCode, 200, "Reset DB")

	funds := map[string]int64{
		"P1": 300,
		"P2": 300,
		"P3": 300,
		"P4": 500,
		"P5": 1000,
	}

	for player, points := range funds {
		log.Println("Fund 300 points for ", player)
		url := fmt.Sprintf("%s/fund", server.URL)
		req, err := http.NewRequest("GET", url, nil)
		q := req.URL.Query()
		q.Add("playerId", player)
		q.Add("points", strconv.FormatInt(points, 10))
		req.URL.RawQuery = q.Encode()
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Fund 300 points for ", player, err)
			return
		}
		assert.Equal(t, res.StatusCode, 200, "Fund 300 points for ", player)
	}

	log.Println("Announce tournament ID 1, deposit 1000")
	url = fmt.Sprintf("%s/announceTournament", server.URL)
	req, err = http.NewRequest("GET", url, nil)
        q := req.URL.Query()
	q.Add("tournamentId", "1")
	q.Add("deposit", "1000")
	req.URL.RawQuery = q.Encode()
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Announce tournament ID 1, deposit 1000", err)
		return
	}
	assert.Equal(t, res.StatusCode, 200, "Announce tournament ID 1, deposit 1000")

	log.Println("P5 joins on his own")
	url = fmt.Sprintf("%s/joinTournament", server.URL)
	req, err = http.NewRequest("GET", url, nil)
	q = req.URL.Query()
	q.Add("tournamentId", "1")
	q.Add("playerId", "P5")
	req.URL.RawQuery = q.Encode()
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("P5 joins on his own", err)
		return
	}
	assert.Equal(t, res.StatusCode, 200, "P5 joins on his own")

	log.Println("P1 joins backed by P2, P3, P4")
	url = fmt.Sprintf("%s/joinTournament", server.URL)
	req, err = http.NewRequest("GET", url, nil)
	q = req.URL.Query()
	q.Add("tournamentId", "1")
	q.Add("playerId", "P1")
	q.Add("backerId", "P2")
	q.Add("backerId", "P3")
	q.Add("backerId", "P4")
	req.URL.RawQuery = q.Encode()
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println("P1 joins backed by P2, P3, P4", err)
		return
	}
	assert.Equal(t, res.StatusCode, 200, "P1 joins backed by P2, P3, P4")

	resultBalances := map[string]int64{
		"P1": 550,
		"P2": 550,
		"P3": 550,
		"P4": 750,
		"P5": 0,
	}

        for player, balance := range resultBalances {
		log.Println("Balance for ", player)
                url := fmt.Sprintf("%s/balance", server.URL)
                req, err := http.NewRequest("GET", url, nil)
                q := req.URL.Query()
                q.Add("playerId", player)
                req.URL.RawQuery = q.Encode()
                res, err = http.DefaultClient.Do(req)
                if err != nil {
                        log.Println("Balance for ", player, err)
                        return
                }

                defer res.Body.Close()
                decoder := json.NewDecoder(res.Body)

                var result Players
                err = decoder.Decode(&result)
		if err != nil {
                        log.Println(err)
                        return
                }

                log.Println(balance, result)
                assert.Equal(t, res.StatusCode, 200, "Balance for ", player)
                assert.Equal(t, balance, result.Balance, "Balance for ", player)
        }
}
