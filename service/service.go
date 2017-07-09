package main

import (
	"github.com/go-ozzo/ozzo-dbx"
	"log"
)

type Service struct {
	DB *dbx.DB
}

func (service *Service) Initialize()(error) {
	log.Println("Initializing service")
	service.CreatePlayersTable()
	service.CreateTournamentsTables()
        return nil
}

func (service *Service) Reset()(error) {
        log.Println("Reset DB")
        return nil
}

func (service *Service) CreatePlayersTable()(error) {
	log.Println("Create player table")
        return nil
}

func (service *Service) CreateTournamentsTables()(error) {
	log.Println("Create tournament tables")
        return nil
}

func (service *Service) Fund(player string, points int64)(error) {
	log.Println("Fund:", player, points)
        return nil
}

func (service *Service) Take(player string, points int64)(error) {
	log.Println("Take:", player, points)
        return nil
}

func (service *Service) AnnounceTournament(tournament int64, deposit int64)(error) {
	log.Println("AnnounceTournament:", tournament, deposit)
        return nil
}

func (service *Service) JoinTournament(tournament int64, player string, backers []string)(error) {
	log.Println("JoinTournament:", tournament, player, backers)
        return nil
}

func (service *Service) ResultTournament(result string)(error) {
	log.Println("ResultTournament:", result)
        return nil
}

func (service *Service) Balance(player string)(int64, error) {
	log.Println("Balance:", player)
        return 0, nil
}
