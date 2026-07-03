package main

import (
	"database/sql"
	"log"

	"github.com/Saker233/simplebank/api"
	db "github.com/Saker233/simplebank/db/sqlc"
	"github.com/Saker233/simplebank/util"
	_ "github.com/lib/pq"
)


func main() {

	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal("Can not load config file:", err)
	}



	conn, err := sql.Open(config.DBDriver, config.DBSource)

	if err != nil {
		log.Fatal("Cannot connect to the db", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)

	if err != nil {
		log.Fatal("Cannot Start Server: ", err)
	}

}
