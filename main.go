////////
// This sample is published as part of the blog article at www.toptal.com/blog
// Visit www.toptal.com/blog and subscribe to our newsletter to read great posts
////////

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pelletier/go-toml"
)

func main() {
	cfg, err := toml.LoadFile("config.tml")
	catch(err)

	url, ok := cfg.Get("buildpack.url").(string)
	if !ok {
		log.Fatal("buildpack.url not defined")
	}
	err = UpdateBuildpack(url)
	catch(err)

	repo, ok := cfg.Get("app.repo").(string)
	if !ok {
		log.Fatal("app.repo not defined")
	}

	env := []string{}
	tree, ok := cfg.Get("app.env").(*toml.TomlTree)
	if !ok {
		log.Fatal("app.env not defined")
	}
	for _, k := range tree.Keys() {
		v, ok := tree.Get(k).(string)
		if !ok {
			log.Fatal("app.env is invalid")
		}
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	procs := map[string]int{}
	tree, ok = cfg.Get("app.procs").(*toml.TomlTree)
	if !ok {
		log.Fatal("app.procs not defined")
	}
	for _, k := range tree.Keys() {
		v, ok := tree.Get(k).(int64)
		if !ok {
			log.Fatal("app.procs is invalid")
		}
		procs[k] = int(v)
	}

	app, err := NewApp(repo, env, procs)
	catch(err)

	err = app.Update()
	catch(err)

	secret, _ := cfg.Get("hook.secret").(string)

	http.Handle("/hook", NewHookHandler(&HookOptions{
		App:    app,
		Secret: secret,
	}))

	addr, ok := cfg.Get("core.addr").(string)
	if !ok {
		log.Fatal("core.addr not defined")
	}

	err = http.ListenAndServe(addr, nil)
	catch(err)
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}
