////////
// This sample is published as part of the blog article at www.toptal.com/blog
// Visit www.toptal.com/blog and subscribe to our newsletter to read great posts
////////

package main

import (
	"log"
	"os"
	"os/exec"
)

func UpdateBuildpack(url string) error {
	log.Print("Updating buildpack")

	_, err := os.Stat("buildpack")
	if os.IsNotExist(err) {
		log.Print(".. initializing buildpack")

		err := os.MkdirAll("buildpack", 0755)
		if err != nil {
			return err
		}

		cmd := exec.Command("git", "clone", url, ".")
		cmd.Dir = "buildpack"
		return cmd.Run()
	}

	cmd := exec.Command("git", "pull")
	cmd.Dir = "buildpack"
	return cmd.Run()
}
