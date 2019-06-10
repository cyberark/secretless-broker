package util

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func RegisterShutdownSignalCallback(shutdownChannel chan<- bool) {
	log.Println("Registering shutdown signal listener...")
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGUSR1,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	go func() {
		exitSignal := <-signalChannel
		log.Printf("Intercepted exit signal '%v'...", exitSignal)
		shutdownChannel <- true
		signal.Reset()
	}()
}
