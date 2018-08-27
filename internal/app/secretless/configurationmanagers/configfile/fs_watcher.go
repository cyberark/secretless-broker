package configfile

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

// AttachWatcher adds a listener of chenge event to a filepath
func AttachWatcher(filename string, runner func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer watcher.Close()

		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("WARN: Watched file '%s' modified!", filename)
					go runner()
				} else if event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Printf("WARN: Watched file '%s' has been removed!",
						filename)

					// Some editors remove the old file and replace it with a new one
					// so we need to give it a bit of time and try to reattach the notifier
					time.Sleep(1 * time.Second)
					log.Printf("WARN: Trying to reattach to '%s'...", filename)
					AttachWatcher(filename, runner)

					// If the file disappeared, we know it changed so run the trigger
					go runner()
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
				break
			}
		}
	}()

	log.Printf("Attaching filesystem notifier onto %s", filename)
	err = watcher.Add(filename)
	if err != nil {
		log.Fatal(err)
		watcher.Close()
	}
}
