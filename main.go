package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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
	filenameRegex := regexp.MustCompile("verwijder dit bestand om '([[:graph:]]+)' veilig te verwijderen")
	for {
		select {
		case ev := <-watcher.Event:
			if ev.Mask == inotify.IN_CREATE|inotify.IN_ISDIR {
				log.Println("[-] disk mounted: " + ev.Name)
				unmountFile := filepath.Join(filepath.Dir(ev.Name), fmt.Sprintf("verwijder dit bestand om '%s' veilig te verwijderen", filepath.Base(ev.Name)))
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
				log.Println("[+] created " + unmountFile)
			}
			if ev.Mask == inotify.IN_DELETE && filenameRegex.MatchString(filepath.Base(ev.Name)) {
				matches := filenameRegex.FindStringSubmatch(ev.Name)
				mp := filepath.Join(mountpoint, matches[1])
				var deviceName string
				// get device name
				cmd := exec.Command("mount")
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Println(err)
				}
				mountpointRegex := regexp.MustCompile(mp)
				devicenameRegex := regexp.MustCompile("([a-z/]+)[0-9]+")
				for _, line := range strings.Split(string(output), "\n") {
					if mountpointRegex.MatchString(line) {
						fields := strings.Split(line, " ")
						if len(fields) > 1 {
							sm := devicenameRegex.FindStringSubmatch(fields[0])
							deviceName = sm[1]
						}
					}
				}
				// unmount partition
				cmd = exec.Command("sudo", "umount", mp)
				err = cmd.Run()
				if err != nil {
					log.Println(err)
					continue
				}
				log.Println("[+] unmounted ", mp)

				// detach (and poweroff) device
				if deviceName != "" {
					cmd = exec.Command("sudo", "udisks", "--detach", deviceName)
					err = cmd.Run()
					if err != nil {
						log.Println(err)
					}
				}
				log.Println("[+] detached ", deviceName)
			}
		case err := <-watcher.Error:
			log.Println("[*] ", err)
		}
	}

}
