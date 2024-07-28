package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	pathBase string
)

type Dirs []Dir

type Dir struct {
	Dirs  []Dir
	Files []fs.DirEntry
	Path  string
}

type QueryResults struct {
	Dirs  Dirs
	Paths []string
}

// simple error check
func check(err error) {
	if err != nil {
		log.Fatalf(err.Error())
	}
}

// execute *exec.Cmd w stdin/stdout
func stdcmd(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	check(err)
}

// WIP: add windows/macos functionality
func homeDir() string {
	homedir, err := os.UserHomeDir()
	check(err)
	return homedir
}

// check if dirs contains subdirs
func hasSubDirs(dirs []fs.DirEntry) bool {
	truth := false
	for _, dir := range dirs {
		if dir.IsDir() {
			truth = true
		}
	}
	return truth
}

// helper for walkDirs
func (d *Dir) walkDir(depth int) {
	dirs, err := os.ReadDir(d.Path)
	check(err)
	if hasSubDirs(dirs) {
		for _, dir := range dirs {
			//prefix := strings.Repeat("  ", depth)
			if (dir.IsDir()) && (dir.Name()[0] != '.') {
				depth++
				newD := Dir{Dirs: d.Dirs, Path: fmt.Sprintf("%s/%s", d.Path, dir.Name())}
				newD.walkDir(depth)
				d.Dirs = append(d.Dirs, newD)
			} else if dir.Name()[0] != '.' {
				d.Files = append(d.Files, dir)
			}
		}
	}
}

// walk []fs.DirEntry, return Dirs
func walkDirs(dirs []fs.DirEntry) Dirs {
	var directories Dirs
	for _, dir := range dirs {
		var d *Dir
		var dirDirs []Dir
		var dirFiles []fs.DirEntry
		if (dir.IsDir()) && (dir.Name()[0] != '.') {
			d = &Dir{Dirs: dirDirs, Path: fmt.Sprintf("%s/%s", pathBase, dir.Name())}
			d.walkDir(0)
			dirDirs = append(dirDirs, *d)
		} else if dir.Name()[0] != '.' {
			dirFiles = append(dirFiles, dir)
		}
		d.Dirs = dirDirs
		d.Files = dirFiles
		directories = append(directories, *d)
	}
	return directories
}

// search zets
func query(q []string) QueryResults {
	//query := strings.Join(q, " ")
	//dirs, err := os.ReadDir(pathBase)
	var qr QueryResults
	return qr
}

// generate index.html
func generateIndex() {
	f, err := os.Create(fmt.Sprintf("%s/index.html", pathBase))
	check(err)
	t := template.New("dirs.html")
	dirs, err := os.ReadDir(pathBase)
	check(err)
	directories := walkDirs(dirs)
	t.Execute(f, directories)
	err = f.Close()
	check(err)
}

func init() {
	pathBase = homeDir()
	if _, err := os.Stat(pathBase + "/zets"); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("paste the url to your zets directory: ")
		var uI string
		fmt.Scanln(&uI)
		split := strings.Split(uI, "/")
		clonecmd := exec.Command("git", "clone", uI)
		clonecmd.Dir = pathBase
		stdcmd(clonecmd)
		pathBase += "/" + split[len(split)-1]
		err := os.WriteFile(fmt.Sprintf("%s/url", pathBase), []byte(uI), 0644)
		check(err)
		fmt.Printf("[log] %s created\n", pathBase)
	} else {
		pathBase += "/zets"
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("welcome to your zettelkasten control panel.\nyour zets were imported locally to %s.\nto create a new zet, use the command 'z create [nameOfTopic] [nameOfZet]'.\nthis will create the following folder structure: nameOfTopic/currentYear/nameOfZet.\ncd into any directory of your knowledge base and use 'z search [keywords]' to search your zets.\n", pathBase)
	} else {
		switch os.Args[1] {
		case "create", "c":
			path := fmt.Sprintf("%s/%s/%d/%s", pathBase, os.Args[2], time.Now().Year(), os.Args[3])
			stdcmd(exec.Command("mkdir", "-p", path))
			stdcmd(exec.Command("touch", fmt.Sprintf("%s/README.md", path)))
			fmt.Printf("vim %s/README.md ?", path)
			var uI string
			fmt.Scanln(&uI)
			if uI == "y" || uI == "yes" {
				stdcmd(exec.Command("vim", fmt.Sprintf("%s/README.md", path)))
			}
		case "serve", "s":
			generateIndex()
		case "query", "q":
			query(os.Args[2:])
		}
	}
}
