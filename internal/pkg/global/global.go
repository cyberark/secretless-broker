package global

import (
    "sync"
    "os"
    "os/signal"
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

func WaitForGlobalCleanUp() {
    shutdownNotifyWaitGroup.Wait()
}