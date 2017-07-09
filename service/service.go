package main

import (
	"github.com/go-ozzo/ozzo-dbx"
	"log"
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
		"backer":        "text[]",
	})

	_, err = q.Execute()
	if err != nil {
		log.Println("DB:", err)
		//return err
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

func (service *Service) Take(player string, points int64) error {
	log.Println("Take:", player, points)
	return nil
}

type Tournaments struct {
	ID       int64 `db:"id"`
	Deposit  int64 `db:"deposit"`
	Finished bool  `db:"finished"`
}

func (service *Service) AnnounceTournament(id int64, deposit int64) error {
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

func (service *Service) JoinTournament(tournament int64, player string, backers []string) error {
	log.Println("JoinTournament:", tournament, player, backers)
	return nil
}

func (service *Service) ResultTournament(result string) error {
	log.Println("ResultTournament:", result)
	return nil
}

type Players struct {
	ID      string `db:"id" json:"playerId,omitempty"`
	Balance int64  `db:"balance" json:"balance,omitempty"`
}

func (service *Service) PlayerBalance(id string) (Players, error) {
	var player Players
	service.db.Select().Model(id, &player)
	return player, nil
}
