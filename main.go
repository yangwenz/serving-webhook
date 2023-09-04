package main

import (
	"github.com/rs/zerolog/log"
	"github.com/yangwenz/model-webhook/api"
	"github.com/yangwenz/model-webhook/utils"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}
	// Start model API server
	runGinServer(config)
}

func runGinServer(config utils.Config) {
	server, err := api.NewServer(config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
