package main

import (
	"errors"
	"github.com/go-ozzo/ozzo-dbx"
	"github.com/lib/pq"
	"log"
	"strings"
)

// Service for impement Social Tournament login
type Service struct {
	db *dbx.DB
}

// Method for create tables and indexes in database
func (service *Service) Initialize() error {
	log.Println("Initializing service")
	service.CreatePlayersTable()
	service.CreateTournamentsTables()
	return nil
}

const truncateSQL = `
	TRUNCATE games;
	TRUNCATE tournaments;
	TRUNCATE players;
`

// Method for reset DB for initial state
func (service *Service) ResetDB() error {

	log.Println("Reset DB")

	q := service.db.NewQuery(truncateSQL)
	_, err := q.Execute()

	return err
}

// Method for create players table id database
func (service *Service) CreatePlayersTable() error {
	log.Println("Create players table")

	q := service.db.CreateTable("players", map[string]string{
		"id":      "text primary key",
		"balance": "bigint",
	})

	_, err := q.Execute()
	if err != nil {
		log.Println("DB:", err)
		return err
	}

	return err
}

const gamesIndexesSQL = `
    CREATE UNIQUE INDEX ON games USING btree(tournament_id, player_id)
`

// Method for create tournaments table
// and create games table with unique(tournamentId,playerId) index
func (service *Service) CreateTournamentsTables() error {
	log.Println("Create tournaments tables")

	// Create tournaments table
	q := service.db.CreateTable("tournaments", map[string]string{
		"id":       "text primary key",
		"deposit":  "bigint",
		"finished": "bool",
	})

	_, err := q.Execute()
	if err != nil {
		log.Println("DB:", err)
		//return err
	}

	// Create games table
	q = service.db.CreateTable("games", map[string]string{
		"id":            "bigserial",
		"tournament_id": "bigint",
		"player_id":     "text",
		"backers":       "text[]",
	})

	_, err = q.Execute()
	if err != nil {
		log.Println("DB:", err)
		//return err
	} else {

		// If games table was created, create index
		q = service.db.NewQuery(gamesIndexesSQL)
		_, err := q.Execute()
		if err != nil {
			log.Println("DB:", err)
		}
	}

	return err
}

const fundSQL = `
	INSERT INTO players
    		(id, balance)
	VALUES
    		({:id}, {:points})
	ON
 		CONFLICT (id)
	DO UPDATE SET
    		balance = players.balance + {:points}
                
`

// Method for fund player with points
// Add playerId into database, if player doesn't exist
func (service *Service) Fund(player string, points int64) error {
	q := service.db.NewQuery(fundSQL)
	q.Bind(dbx.Params{
		"id":     player,
		"points": points,
	})

	_, err := q.Execute()
	if err != nil {
		log.Println("DB:", err)
	}

	return err
}

const takeSQL = `
        UPDATE players
	SET balance = balance - {:points}
	WHERE 
		id = {:id}
		AND balance >= {:points}
`

// Method for take points from player
func (service *Service) Take(player string, points int64) (int64, error) {
	q := service.db.NewQuery(takeSQL)
	q.Bind(dbx.Params{
		"id":     player,
		"points": points,
	})

	result, err := q.Execute()
	if err != nil {
		log.Println("DB:", err)
		return 0, err
	}

	r, err := result.RowsAffected()

	return r, err
}

// Structure (Model) for insert new Tournaments into database
type Tournaments struct {
	ID       string `db:"id"`
	Deposit  int64  `db:"deposit"`
	Finished bool   `db:"finished"`
}

// Method for insert tournaments into database
func (service *Service) AnnounceTournament(id string, deposit int64) error {
	// Prepare model
	tournament := Tournaments{
		ID:       id,
		Deposit:  deposit,
		Finished: false,
	}
	// Insert into database
	err := service.db.Model(&tournament).Insert()
	if err != nil {
		log.Println("DB:", err)
	}

	return err
}

// Method for implement Join tournament logic
func (service *Service) JoinTournament(id string, player string, backers []string) error {
	var tournament Tournaments

	// Load from database  tournament by id (and it's not finished)
	service.db.Select("id", "deposit", "finished").
		From("tournaments").
		Where(dbx.HashExp{"id": id, "finished": false}).
		One(&tournament)

		// Check if wanted tournament exit
	if tournament == (Tournaments{}) {
		return errors.New("not found")
	}

	// Calculate points per player/backer
	var points int64
	backersLen := len(backers)
	points = tournament.Deposit / (1 + int64(backersLen))

	// Start transaction
	tx, _ := service.db.Begin()

	// Save player with backers to database
	_, err := tx.Insert("games", dbx.Params{
		"tournament_id": id,
		"player_id":     player,
		"backers":       pq.Array(backers),
	}).Execute()

	// if error do transaction rollback
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	// Take points from player balance and backers balances
	backers = append(backers, player)

	for _, p := range backers {
		q := tx.NewQuery(takeSQL)
		q.Bind(dbx.Params{
			"id":     p,
			"points": points,
		})

		result, err := q.Execute()

		// if error do transaction rollback
		if err != nil {
			log.Println("DB:", err)
			tx.Rollback()
			return err
		}

		// If no row afected, it's mean no backerId or playerId found id database, do rollback
		r, err := result.RowsAffected()
		if r == 0 {
			tx.Rollback()
			return errors.New("not found")
		}
	}

	// Commit success transaction
	tx.Commit()

	return nil
}

const resultSQL = `
    UPDATE tournaments
    SET finished = 't'
    WHERE id = {:id}
`
const prizeSQL = `
    UPDATE players
    SET balance = balance + {:points}
    WHERE id = {:id}
`

const winnerSQL = `
    SELECT 
        id,
        player_id,
        backers
    FROM games
    WHERE
         tournament_id = {:tournamentId}        
         AND player_id = {:playerId}
    LIMIT 1
`

// Structure (Model) for load winner from database
type PlayerWinner struct {
	ID       string `db:"id"`
	PlayerID string `db:"player_id"`
	Backers  string `db:"backers"`
}

// Method for imprement Result Tournament logic
func (service *Service) ResultTournament(id string, results []Winner) error {
	// Start transaction, do rollback if any errors
	tx, _ := service.db.Begin()

	// Finish tournament
	q := tx.NewQuery(resultSQL)
	q.Bind(dbx.Params{
		"id": id,
	})

	result, err := q.Execute()

	// Check for load error
	if err != nil {
		log.Println("DB:", err)
		tx.Rollback()
		return err
	}

	// Tournament must be in database and updated
	r, err := result.RowsAffected()
	if r == 0 {
		tx.Rollback()
		return errors.New("not found")
	}

	// Process winners
	for _, winner := range results {

		// Load winner from database
		q = tx.NewQuery(winnerSQL)
		q.Bind(dbx.Params{
			"tournamentId": id,
			"playerId":     winner.PlayerId,
		})
		var playerWinner PlayerWinner
		err = q.One(&playerWinner)
		// winner must be
		if err != nil {
			tx.Rollback()
			e := err.Error()
			if e == "sql: no rows in result set" {
				return errors.New("not found")
			}
			log.Println("DB:", err)
			return err
		}
		// Get backers for winner
		pr := strings.Trim(playerWinner.Backers, "{}")
		players := strings.Split(pr, ",")
		players = append(players, winner.PlayerId)

		// Calculate prize points for player/backers
		playersLen := len(players)
		points := winner.Prize / int64(playersLen)

		// Update player and backers balances
		for _, p := range players {
			q := tx.NewQuery(prizeSQL)
			q.Bind(dbx.Params{
				"id":     p,
				"points": points,
			})

			result, err := q.Execute()

			// If error do rollback transaction
			if err != nil {
				log.Println("DB:", err)
				tx.Rollback()
				return err
			}

			// If balance not updated do rollback transaction
			r, err := result.RowsAffected()
			if r == 0 {
				tx.Rollback()
				return errors.New("not found")
			}
		}
	}

	// Commit success transaction
	tx.Commit()
	return nil
}

// Structure for player balance response
type Players struct {
	ID      string `db:"id" json:"playerId"`
	Balance int64  `db:"balance" json:"balance"`
}

// Method for get player balance from database
func (service *Service) PlayerBalance(id string) (Players, error) {
	var player Players
	err := service.db.Select().Model(id, &player)
	return player, err
}
