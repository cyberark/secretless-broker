package global

import (
    "sync"
    "os"
    "os/signal"
    "log"
    "syscall"
)

var shutdownNotifyWaitGroup = sync.WaitGroup{}

func ShutdownChCreator(sig ...os.Signal) (chan os.Signal, func()) {
    shutdownCh := make(chan os.Signal, 1)
    signal.Notify(shutdownCh, sig...)
    shutdownNotifyWaitGroup.Add(1)

    shutdownChCleaned := false
    shutdownChCleanedMutex := &sync.Mutex{}
    cleanUpShutdownCh := func() {
        shutdownChCleanedMutex.Lock()

        if shutdownChCleaned {
            return
        }
        close(shutdownCh)
        signal.Stop(shutdownCh)
        shutdownNotifyWaitGroup.Done()
        shutdownChCleaned = true

        shutdownChCleanedMutex.Unlock()
    }

    return shutdownCh, cleanUpShutdownCh
}

func init() {
    shutdownCh, cleanUpShutdownCh := ShutdownChCreator(syscall.SIGABRT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

    go func() {
        defer cleanUpShutdownCh()

        <- shutdownCh
        shutdownNotifyWaitGroup.Wait()

        log.Printf("Exiting...")
        os.Exit(0)
    }()
}
