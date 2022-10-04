package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	socketFullPath := flag.String("socketFullPath", "", "the full path to the steward socket file")
	messageFullPath := flag.String("messageFullPath", "", "the full path to the message to send at intervals")
	interval := flag.Int("interval", 10, "the interval in seconds between sending messages")
	watchFolder := flag.String("watchFolder", "", "the folder to watch for new messages to send")
	flag.Parse()

	if *socketFullPath == "" {
		log.Printf("error: you need to specify the full path to the socket\n")
		return
	}

	if *interval <= 0 {
		if *messageFullPath == "" {
			log.Printf("error: you need to specify the full path to the message to be sent at intervals\n")
			return
		}

		err := sendAtInterval(*interval, *messageFullPath, *socketFullPath)
		if err != nil {
			os.Exit(1)
		}
	}

	if *watchFolder != "" {
		checkFileUpdated(*watchFolder, *socketFullPath)
	}

}

// sendAtInterval will send the given message on the intervals specified.
func sendAtInterval(interval int, messageFullPath string, socketFullPath string) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ticker := time.NewTicker(time.Second * time.Duration(interval))

	for {
		select {
		case <-ticker.C:
			err := messageFileToSocket(socketFullPath, messageFullPath)
			if err != nil {
				log.Printf("%v\n", err)
			}
		case <-sigCh:
			log.Printf("info: received signal to quit..\n")
			return nil
		}
	}
}

func messageFileToSocket(socketFullPath string, messageFullPath string) error {
	socket, err := net.Dial("unix", socketFullPath)
	if err != nil {
		return fmt.Errorf(" * failed: could not open socket file for writing: %v", err)
	}
	defer socket.Close()

	fp, err := os.Open(messageFullPath)
	if err != nil {
		return fmt.Errorf(" * failed: could not open message file for reading: %v", err)
	}
	defer fp.Close()

	_, err = io.Copy(socket, fp)
	if err != nil {
		return fmt.Errorf("error: io.Copy failed: %v", err)
	}

	log.Printf("info: succesfully wrote message to socket\n")

	err = os.Remove(messageFullPath)
	if err != nil {
		return fmt.Errorf("error: os.Remove failed: %v", err)
	}

	return nil
}

// checkFileUpdated will check for new files in the specified watchFolder.
// When a new file appears in the folder it will be read, and written to
// the specified socket file.
func checkFileUpdated(watchFolder string, socketFullPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Failed fsnotify.NewWatcher")
		return
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op == fsnotify.Create {
					log.Println("created file:", event.Name)

					fileName := filepath.Base(event.Name)
					messageFullPath := filepath.Join(watchFolder, fileName)
					err := messageFileToSocket(socketFullPath, messageFullPath)
					if err != nil {
						log.Printf("%v\n", err)
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(watchFolder)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
