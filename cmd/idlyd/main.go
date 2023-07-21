package main

import (
	"context"
	"fmt"
	"idly/internal/api"
	"idly/internal/config"
	"idly/internal/dao"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	ctx, cancel := context.WithCancelCause(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	// Start a goroutine to listen for signals and gracefully shutdown the program.
	go func() {
		sig := <-signals
		err := fmt.Errorf("received signal: %v, shutting down", sig)
		fmt.Println(err)
		cancel(err)

		time.Sleep(10 * time.Second)
		fmt.Println("Has not terminated gracefully, Terminating.")
		os.Exit(1)
	}()

	db, err := dao.NewBadger(config.Get().BadgerURI)
	if err != nil {
		fmt.Println("Could not start db,", err)
		os.Exit(1)
	}

	api.Start(ctx, db)
	fmt.Println("[Shutdown] Shutting down database")
	_ = db.Close()
	fmt.Println("[Shutdown] Database is shutdown")
	fmt.Println("[Shutdown] Terminating")
}
