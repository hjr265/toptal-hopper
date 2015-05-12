////////
// This sample is published as part of the blog article at www.toptal.com/blog
// Visit www.toptal.com/blog and subscribe to our newsletter to read great posts
////////

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hecticjeff/procfile"
)

type App struct {
	Repo  string
	Env   []string
	Procs map[string]int

	repoDir string
	appDir  string
	logsDir string

	nodes []*Node
}

func NewApp(repo string, env []string, procs map[string]int) (*App, error) {
	repoDir, err := filepath.Abs("repo")
	if err != nil {
		return nil, err
	}
	appDir, err := filepath.Abs("app")
	if err != nil {
		return nil, err
	}
	logsDir, err := filepath.Abs("logs")
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(logsDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(logsDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	a := &App{
		Repo:    repo,
		Env:     env,
		Procs:   procs,
		repoDir: repoDir,
		appDir:  appDir,
		logsDir: logsDir,
	}

	for k, v := range procs {
		for i := 0; i < v; i++ {
			n, err := NewNode(a, k, i+1, 9000+len(a.nodes))
			if err != nil {
				return nil, err
			}
			a.nodes = append(a.nodes, n)
		}
	}
	return a, nil
}

func (a *App) Update() error {
	log.Print("Updating application")

	_, err := os.Stat(a.repoDir)
	if os.IsNotExist(err) {
		err = a.initRepo()
		if err != nil {
			return err
		}
	}

	err = a.fetchChanges()
	if err != nil {
		return err
	}

	err = a.stopProcs()
	if err != nil {
		return err
	}

	err = a.compileApp()
	if err != nil {
		return err
	}

	return a.startProcs()
}

func (a *App) initRepo() error {
	log.Print("Initializing repository")

	err := os.MkdirAll(a.repoDir, 0755)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "--git-dir="+a.repoDir, "init")
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("git", "--git-dir="+a.repoDir, "remote", "add", "origin", fmt.Sprintf("git@github.com:%s.git", a.Repo))
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) fetchChanges() error {
	log.Print("Fetching changes")

	cmd := exec.Command("git", "--git-dir="+a.repoDir, "fetch", "-f", "origin", "master:master")
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) compileApp() error {
	log.Print("Compiling application")

	_, err := os.Stat(a.appDir)
	if !os.IsNotExist(err) {
		err = os.RemoveAll(a.appDir)
		if err != nil {
			return err
		}
	}
	err = os.MkdirAll(a.appDir, 0755)
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "--git-dir="+a.repoDir, "--work-tree="+a.appDir, "checkout", "-f", "master")
	cmd.Dir = a.appDir
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	buildpackDir, err := filepath.Abs("buildpack")
	if err != nil {
		return err
	}

	cmd = exec.Command("bash", filepath.Join(buildpackDir, "bin", "detect"), a.appDir)
	cmd.Dir = buildpackDir
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	cacheDir, err := filepath.Abs("cache")
	if err != nil {
		return err
	}
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return err
	}

	cmd = exec.Command("bash", filepath.Join(buildpackDir, "bin", "compile"), a.appDir, cacheDir)
	cmd.Dir = a.appDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (a *App) stopProcs() error {
	log.Print(".. stopping processes")

	for _, n := range a.nodes {
		err := n.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) startProcs() error {
	log.Print("Starting processes")

	err := a.readProcfile()
	if err != nil {
		return err
	}

	for _, n := range a.nodes {
		err = n.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) readProcfile() error {
	f, err := os.Open(filepath.Join(a.appDir, "Procfile"))
	if err != nil {
		return err
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	procs := procfile.Parse(string(buf))

	for _, node := range a.nodes {
		p := procs[node.Name]
		node.Cmd = p.Command + " " + strings.Join(p.Arguments, " ")
	}

	return nil
}
