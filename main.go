package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

func check(err error) {
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func execCommand(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	check(err)
}

func homeDir() string {
	currentUser, err := user.Current()
	check(err)

	return fmt.Sprintf("/home/%s", currentUser.Username)
}

func main() {
	pathBase := homeDir()

	if _, err := os.Stat(pathBase + "/zets"); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("zets url: ")
		var uI string
		fmt.Scanln(&uI)
		split := strings.Split(uI, "/")
		execCommand(exec.Command("git", "clone", uI))
		pathBase = pathBase + split[len(split)-1]
		execCommand(exec.Command("mv", split[len(split)-1], pathBase))
		fmt.Printf("[log] %s created\n", pathBase)
	} else {
		pathBase = pathBase + "/zets"
	}

	if len(os.Args) == 1 {
		fmt.Printf("welcome to your zettelkasten control panel.\nyour zets were imported locally to %s.\nto create a new zet, use the command 'z create [nameOfTopic] [nameOfZet]'.\nthis will create the following folder structure: nameOfTopic/currentYear/nameOfZet.\ncd into any directory of your knowledge base and use 'z search [keywords]' to search your zets.\n", pathBase)
	} else {
		switch os.Args[1] {
		case "create", "c":
			path := fmt.Sprintf("%s/%s/%d/%s", pathBase, os.Args[2], time.Now().Year(), os.Args[3])
			execCommand(exec.Command("mkdir", "-p", path))
			execCommand(exec.Command("touch", fmt.Sprintf("%s/README.md", path)))
			fmt.Printf("vim %s/README.md ?", path)
			var uI string
			fmt.Scanln(&uI)
			if uI == "y" || uI == "yes" {
				execCommand(exec.Command("vim", fmt.Sprintf("%s/README.md", path)))
			}
		case "search", "s":
			dir, err := os.Getwd()
			check(err)
			files, err := os.ReadDir(dir)
			fmt.Println(files[0].Name())
		}
	}
}
