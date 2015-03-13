package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"syscall"

	"golang.org/x/exp/inotify"
)

func main() {
	logFile, err := os.OpenFile("./sambaunmount.log", syscall.O_WRONLY|syscall.O_APPEND|syscall.O_CREAT, 0666)
	if err == nil {
		log.SetOutput(logFile)
	}

	mountpoint := "/media/fkalter"
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.AddWatch(mountpoint, inotify.IN_CREATE|inotify.IN_DELETE)
	if err != nil {
		log.Fatal(err)
	}
	filenameRegex := regexp.MustCompile("verwijder '([[:graph:]]+)' veilig door dit bestand te verwijderen")
	for {
		select {
		case ev := <-watcher.Event:
			if ev.Mask == inotify.IN_CREATE|inotify.IN_ISDIR {
				fmt.Println("[-] disk mounted: " + ev.Name)
				unmountFile := filepath.Join(filepath.Dir(ev.Name), fmt.Sprintf("verwijder '%s' veilig door dit bestand te verwijderen", filepath.Base(ev.Name)))
				fd, err := os.Create(unmountFile)
				if err != nil {
					log.Println(err)
					continue
				}
				err = fd.Close()
				if err != nil {
					log.Println(err)
					continue
				}
				fmt.Println("[+] created " + unmountFile)
			}
			if ev.Mask == inotify.IN_DELETE && filenameRegex.MatchString(filepath.Base(ev.Name)) {
				matches := filenameRegex.FindStringSubmatch(ev.Name)
				cmd := exec.Command("sudo", "umount", filepath.Join(mountpoint, matches[1]))
				cmd.Run()
				if err != nil {
					log.Println(err)
				}
			}
		case err := <-watcher.Error:
			log.Println("[*] ", err)
		}
	}

}
