////////
// This sample is published as part of the blog article at www.toptal.com/blog
// Visit www.toptal.com/blog and subscribe to our newsletter to read great posts
////////

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type State int

const (
	StateDown State = iota
	StateUp
)

type NextState struct {
	State
	doneCh chan int
}

type Node struct {
	*os.Process

	App  *App
	Name string
	No   int
	Cmd  string

	Port int

	State   State
	stateCh chan NextState

	logFile *os.File
}

func NewNode(app *App, name string, no int, port int) (*Node, error) {
	logFile, err := os.OpenFile(filepath.Join(app.logsDir, fmt.Sprintf("%s.%d.txt", name, no)), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	n := &Node{
		App:     app,
		Name:    name,
		No:      no,
		Port:    port,
		stateCh: make(chan NextState),
		logFile: logFile,
	}

	go func() {
		for {
			next := <-n.stateCh
			if n.State == next.State {
				if next.doneCh != nil {
					close(next.doneCh)
				}
				continue
			}

			switch next.State {
			case StateUp:
				log.Printf("Starting process %s.%d", n.Name, n.No)

				cmd := exec.Command("bash", "-c", "for f in .profile.d/*; do source $f; done; "+n.Cmd)
				cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", n.App.appDir))
				cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", n.Port))
				cmd.Env = append(cmd.Env, n.App.Env...)
				cmd.Dir = n.App.appDir
				cmd.Stdout = n.logFile
				cmd.Stderr = n.logFile
				err := cmd.Start()
				if err != nil {
					log.Printf("Process %s.%d exited", n.Name, n.No)
					n.State = StateUp

				} else {
					n.Process = cmd.Process
					n.State = StateUp
				}

				if next.doneCh != nil {
					close(next.doneCh)
				}

				go func() {
					err := cmd.Wait()
					if err != nil {
						log.Printf("Process %s.%d exited", n.Name, n.No)
						n.stateCh <- NextState{
							State: StateDown,
						}
					}
				}()

			case StateDown:
				log.Printf("Stopping process %s.%d", n.Name, n.No)

				if n.Process != nil {
					n.Process.Kill()
					n.Process = nil
				}

				n.State = StateDown
				if next.doneCh != nil {
					close(next.doneCh)
				}
			}
		}
	}()

	return n, nil
}

func (n *Node) Start() error {
	n.stateCh <- NextState{
		State: StateUp,
	}
	return nil
}

func (n *Node) Stop() error {
	doneCh := make(chan int)
	n.stateCh <- NextState{
		State:  StateDown,
		doneCh: doneCh,
	}
	<-doneCh
	return nil
}
