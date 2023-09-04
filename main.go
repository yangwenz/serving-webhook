package main

import (
	"github.com/rs/zerolog/log"
	"github.com/yangwenz/model-webhook/api"
	"github.com/yangwenz/model-webhook/storage"
	"github.com/yangwenz/model-webhook/utils"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}
	// S3 uploader
	uploader, err := storage.NewS3Uploader(config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create S3 uploader")
	}
	// Start model API server
	runGinServer(config, uploader)
}

func runGinServer(config utils.Config, uploader storage.Store) {
	server, err := api.NewServer(config, uploader)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
