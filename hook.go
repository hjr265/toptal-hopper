////////
// This sample is published as part of the blog article at www.toptal.com/blog
// Visit www.toptal.com/blog and subscribe to our newsletter to read great posts
////////

package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
)

type HookOptions struct {
	App    *App
	Secret string
}

func NewHookHandler(o *HookOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		evName := r.Header.Get("X-Github-Event")
		if evName != "push" {
			log.Printf("Ignoring '%s' event", evName)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if o.Secret != "" {
			ok := false
			for _, sig := range strings.Fields(r.Header.Get("X-Hub-Signature")) {
				if !strings.HasPrefix(sig, "sha1=") {
					continue
				}
				sig = strings.TrimPrefix(sig, "sha1=")
				mac := hmac.New(sha1.New, []byte(o.Secret))
				mac.Write(body)
				if sig == hex.EncodeToString(mac.Sum(nil)) {
					ok = true
					break
				}
			}
			if !ok {
				log.Printf("Ignoring '%s' event with incorrect signature", evName)
				return
			}
		}

		ev := github.PushEvent{}
		err = json.Unmarshal(body, &ev)
		if err != nil {
			log.Printf("Ignoring '%s' event with invalid payload", evName)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if ev.Repo.FullName == nil || *ev.Repo.FullName != o.App.Repo {
			log.Printf("Ignoring '%s' event with incorrect repository name", evName)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		log.Printf("Handling '%s' event for %s", evName, o.App.Repo)

		err = o.App.Update()
		if err != nil {
			return
		}
	})
}
