package main

import (
	"context"
	"istio-redirector/pkg/files"
	"istio-redirector/pkg/web"
	"os"
	"os/signal"
	"time"

	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
)

func main() {
	logs.SetLevel(logs.DEBUG)

	cfg, err := files.ReadConfigFile()
	if err != nil {
		logs.WithE(err).Error("can't load config file")
		os.Exit(1)
	}

	srv := web.Register(cfg)

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		logs.WithField("address", cfg.ServerURL).Info("server has started")
		if err := srv.ListenAndServe(); err != nil {
			logs.WithE(err).Info("server has stopped")
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	logs.Info("shutting down")
	os.Exit(0)
}
