package main

import (
	_ "github.com/tarent/loginsrv/htpasswd"
	_ "github.com/tarent/loginsrv/osiam"

	"github.com/tarent/loginsrv/login"

	"context"
	"fmt"
	"github.com/tarent/loginsrv/logging"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const applicationName = "loginsrv"

func main() {
	config := login.ReadConfig()
	if err := logging.Set(config.LogLevel, config.TextLogging); err != nil {
		exit(nil, err)
	}

	configToLog := *config
	configToLog.JwtSecret = "..."
	logging.LifecycleStart(applicationName, configToLog)

	h, err := login.NewHandler(config)
	if err != nil {
		exit(nil, err)
	}

	handlerChain := logging.NewLogMiddleware(h)

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	port := config.Port
	if port != "" {
		port = fmt.Sprintf(":%s", port)
	}

	httpSrv := &http.Server{Addr: port, Handler: handlerChain}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logging.ServerClosed(applicationName)
			} else {
				exit(nil, err)
			}
		}
	}()
	logging.LifecycleStop(applicationName, <-stop, nil)

	ctx, ctxCancel := context.WithTimeout(context.Background(), config.GracePeriod)

	httpSrv.Shutdown(ctx)
	ctxCancel()
}

var exit = func(signal os.Signal, err error) {
	logging.LifecycleStop(applicationName, signal, err)
	if err == nil {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
