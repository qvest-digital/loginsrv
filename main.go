package main

import (
	_ "github.com/tarent/loginsrv/htpasswd"
	_ "github.com/tarent/loginsrv/osiam"

	"github.com/tarent/loginsrv/login"

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

	logShutdownEvent()

	configToLog := *config
	configToLog.JwtSecret = "..."
	logging.LifecycleStart(applicationName, configToLog)

	h, err := login.NewHandler(config)
	if err != nil {
		exit(nil, err)
	}

	handlerChain := logging.NewLogMiddleware(h)

	exit(nil, http.ListenAndServe(config.Host+":"+config.Port, handlerChain))
}

func logShutdownEvent() {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		exit(<-c, nil)
	}()
}

var exit = func(signal os.Signal, err error) {
	logging.LifecycleStop(applicationName, signal, err)
	if err == nil {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
