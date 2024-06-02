package main

import (
	"context"
	"github.com/HyperGAI/serving-webhook/api"
	db "github.com/HyperGAI/serving-webhook/db/sqlc"
	"github.com/HyperGAI/serving-webhook/storage"
	"github.com/HyperGAI/serving-webhook/utils"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}
	// S3 store
	var store storage.Store = nil
	if config.AWSBucket != "" && config.AWSBucket != "empty" {
		store, err = storage.NewS3Store(config)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot create S3 store")
		}
	}
	// Redis cache
	var cache storage.Cache = nil
	if config.RedisAddress != "" && config.RedisAddress != "empty" {
		cache, err = storage.NewRedisClient(config)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot create redis cache")
		}
	}
	// Database
	var database db.Store = nil
	if config.DBSource != "" && config.DBSource != "empty" {
		connPool, err := pgxpool.New(context.Background(), config.DBSource)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot connect to db")
		}
		defer connPool.Close()

		runDBMigration(config.MigrationURL, config.DBSource)
		database = db.NewStore(connPool)
	}
	// Start model API server
	runGinServer(config, store, cache, database)
}

func runGinServer(config utils.Config, store storage.Store, cache storage.Cache, database db.Store) {
	server, err := api.NewServer(config, store, cache, database)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}
	/*
		err = server.Start(config.HTTPServerAddress)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot start server")
		}
	*/
	httpServer := &http.Server{
		Addr:    config.HTTPServerAddress,
		Handler: server.Handler(),
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("cannot start server")
		}
	}()

	// https://gin-gonic.com/docs/examples/graceful-restart-or-stop/
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown")
	}
	// catching ctx.Done(). timeout of 10 seconds.
	select {
	case <-ctx.Done():
		log.Info().Msg("timeout of 10 seconds")
	}
	log.Info().Msg("server exiting")
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance")
	}
	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("failed to run migrate up")
	}
	log.Info().Msg("db migrated successfully")
}
