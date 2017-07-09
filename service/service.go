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

func (service *Service) Reset() error {
	log.Println("Reset DB")
	return nil
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

func (service *Service) Fund(player string, points int64) error {
	log.Println("Fund:", player, points)
	return nil
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
	log.Println("AnnounceTournament:", id, deposit)

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

func (service *Service) Balance(player string) (int64, error) {
	log.Println("Balance:", player)
	return 0, nil
}
