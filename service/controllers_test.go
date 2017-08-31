package main

import (
	"github.com/stretchr/testify/assert"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFundEndpoint(t *testing.T) {

        db := initDatabase()
        defer db.Close()
	server := httptest.NewServer(initRouter(db))
	defer server.Close()

	url := fmt.Sprintf("%s/fund", server.URL)

	req, err := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)

	if err != nil {
            log.Println("Error:", err)
	    return
	}

	assert.Equal(t, res.StatusCode, 400, "Check params")

	q := req.URL.Query()
	q.Add("playerId", "P1")
	q.Add("points", "300")
	req.URL.RawQuery = q.Encode()
        res, err = http.DefaultClient.Do(req)
	assert.Equal(t, res.StatusCode, 200, "Fund 300 points for P1")
}
