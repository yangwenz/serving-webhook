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
	// S3 store
	store, err := storage.NewS3Store(config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create S3 store")
	}
	// Redis cache
	cache, err := storage.NewRedisClient(config)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create redis cache")
	}
	// Start model API server
	runGinServer(config, store, cache)
}

func runGinServer(config utils.Config, store storage.Store, cache storage.Cache) {
	server, err := api.NewServer(config, store, cache)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
