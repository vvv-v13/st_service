package main

import (
	"errors"
	"github.com/go-ozzo/ozzo-dbx"
	"github.com/lib/pq"
	"log"
	"strings"
)

type Service struct {
	db *dbx.DB
}

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

func (service *Service) ResetDB() error {

	log.Println("Reset DB")

	q := service.db.NewQuery(truncateSQL)
	_, err := q.Execute()

	return err
}

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

func (service *Service) CreateTournamentsTables() error {
	log.Println("Create tournaments tables")

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

type Tournaments struct {
	ID       string `db:"id"`
	Deposit  int64  `db:"deposit"`
	Finished bool   `db:"finished"`
}

func (service *Service) AnnounceTournament(id string, deposit int64) error {
	tournament := Tournaments{
		ID:       id,
		Deposit:  deposit,
		Finished: false,
	}

	err := service.db.Model(&tournament).Insert()
	if err != nil {
		log.Println("DB:", err)
	}

	return err
}

func (service *Service) JoinTournament(id string, player string, backers []string) error {
	var tournament Tournaments

	service.db.Select("id", "deposit", "finished").
		From("tournaments").
		Where(dbx.HashExp{"id": id, "finished": false}).
		One(&tournament)

	if tournament == (Tournaments{}) {
		return errors.New("not found")
	}

	var points int64
	backersLen := len(backers)
	points = tournament.Deposit / (1 + int64(backersLen))

	tx, _ := service.db.Begin()

	_, err := tx.Insert("games", dbx.Params{
		"tournament_id": id,
		"player_id":     player,
		"backers":       pq.Array(backers),
	}).Execute()

	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	backers = append(backers, player)

	for _, p := range backers {
		q := tx.NewQuery(takeSQL)
		q.Bind(dbx.Params{
			"id":     p,
			"points": points,
		})

		result, err := q.Execute()
		if err != nil {
			log.Println("DB:", err)
			tx.Rollback()
			return err
		}

		r, err := result.RowsAffected()
		if r == 0 {
			tx.Rollback()
			return errors.New("not found")
		}
	}

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

type PlayerWinner struct {
	ID       string `db:"id"`
	PlayerID string `db:"player_id"`
	Backers  string `db:"backers"`
}

func (service *Service) ResultTournament(id string, results []Winner) error {
	tx, _ := service.db.Begin()
	q := tx.NewQuery(resultSQL)
	q.Bind(dbx.Params{
		"id": id,
	})

	result, err := q.Execute()
	if err != nil {
		log.Println("DB:", err)
		tx.Rollback()
		return err
	}

	r, err := result.RowsAffected()
	if r == 0 {
		tx.Rollback()
		return errors.New("not found")
	}

	for _, winner := range results {
		q = tx.NewQuery(winnerSQL)
		q.Bind(dbx.Params{
			"tournamentId": id,
			"playerId":     winner.PlayerId,
		})
		var playerWinner PlayerWinner
		err = q.One(&playerWinner)
		if err != nil {
			tx.Rollback()
			e := err.Error()
			if e == "sql: no rows in result set" {
				return errors.New("not found")
			}
			log.Println("DB:", err)
			return err
		}

		pr := strings.Trim(playerWinner.Backers, "{}")
		players := strings.Split(pr, ",")
		players = append(players, winner.PlayerId)
		playersLen := len(players)

		points := winner.Prize / int64(playersLen)

		for _, p := range players {
			q := tx.NewQuery(prizeSQL)
			q.Bind(dbx.Params{
				"id":     p,
				"points": points,
			})

			result, err := q.Execute()
			if err != nil {
				log.Println("DB:", err)
				tx.Rollback()
				return err
			}

			r, err := result.RowsAffected()
			if r == 0 {
				tx.Rollback()
				return errors.New("not found")
			}
		}
	}

	tx.Commit()
	return nil
}

type Players struct {
	ID      string `db:"id" json:"playerId"`
	Balance int64  `db:"balance" json:"balance"`
}

func (service *Service) PlayerBalance(id string) (Players, error) {
	var player Players
	err := service.db.Select().Model(id, &player)
	return player, err
}
