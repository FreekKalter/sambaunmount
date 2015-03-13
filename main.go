package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"golang.org/x/exp/inotify"
)

func main() {

	watcher, err := inotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	err = watcher.AddWatch("/media/fkalter", inotify.IN_CREATE)
	if err != nil {
		panic(err)
	}
	err = watcher.AddWatch("/media/fkalter", inotify.IN_DELETE)
	if err != nil {
		panic(err)
	}
	filenameRegex := regexp.MustCompile("verwijder '([[:graph:]]+)' veilig door dit bestand te verwijderen")
	for {
		select {
		case ev := <-watcher.Event:
			if ev.Mask == inotify.IN_CREATE|inotify.IN_ISDIR {
				unmountFile := filepath.Join(filepath.Dir(ev.Name), fmt.Sprintf("verwijder '%s' veilig door dit bestand te verwijderen", filepath.Base(ev.Name)))
				fmt.Println(unmountFile)
				fd, err := os.Create(unmountFile)
				if err != nil {
					panic(err)
				}
				err = fd.Close()
				if err != nil {
					panic(err)
				}
			}
			if ev.Mask == inotify.IN_DELETE && filenameRegex.MatchString(filepath.Base(ev.Name)) {
				matches := filenameRegex.FindStringSubmatch(ev.Name)
				cmd := exec.Command("sudo", "umount", filepath.Join("/media/fkalter", matches[1]))
				cmd.Run()
				if err != nil {
					log.Println(err)
				}
			}
		case err := <-watcher.Error:
			log.Println("error:", err)
		}
	}

}
