package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

var (
	pathBase string
)

type Dir struct {
	dir     fs.DirEntry
	relPath string
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
	currentUser, err := user.Current()
	check(err)

	return fmt.Sprintf("/home/%s", currentUser.Username)
}

// generate style.css
func generateStyle() {
	f, err := os.Create(fmt.Sprintf("%s/style.css", pathBase))
	check(err)
	_, err = f.WriteString(`
* {
  margin: 0;
  padding: 0;
}

#zet-reader {
  display: flex;
  justify-content: space-around;
}
`)
}

func (d Dir) walkDir(depth int) []string {
	dirs, err := os.ReadDir(d.relPath)
	check(err)
	var lines []string
	for _, dir := range dirs {
		prefix := strings.Repeat("  ", depth)
		if (dir.IsDir() == true) && (dir.Name()[0] != '.') {
			lines = append(lines, fmt.Sprintf("      %s<ul><li><p id=\"%s\"><a href=\"#\">%s</a></p></li>", prefix, dir.Name(), dir.Name()))
			depth++
			d := Dir{dir: dir, relPath: fmt.Sprintf("%s/%s", d.relPath, dir.Name())}
			d.walkDir(depth)
		} else if dir.Name()[0] != '.' {
			lines = append(lines, fmt.Sprintf("      %s<li><p id=\"%s\"><a href=\"#\">%s</a></p></li>", prefix, dir.Name(), dir.Name()))
		}
		lines = append(lines, fmt.Sprintf("      %s</ul>", prefix))
	}
	return lines
}

func walkDirs(dirs []fs.DirEntry) []string {
	var lines []string
	for _, dir := range dirs {
		if (dir.IsDir() == true) && (dir.Name()[0] != '.') {
			d := Dir{dir: dir, relPath: fmt.Sprintf("%s/%s", pathBase, dir.Name())}
			lines = append(lines, d.walkDir(0)...)
		}
	}
	return lines
}

// generate index.html
func generateIndex() {
	f, err := os.Create(fmt.Sprintf("%s/index.html", pathBase))
	check(err)
	_, err = f.WriteString(`
<!DOCTYPE html>
<html>
  <head>
    <title>zettelkasten</title>
    <script type="module" src="https://md-block.verou.me/md-block.js"></script>
  </head>

  <body>
    <div id="zet-reader">
`)
	f.Sync()
	dirs, err := os.ReadDir(pathBase)
	check(err)
	lines := walkDirs(dirs)
	for _, line := range lines {
		fmt.Fprintln(f, line)
	}
	f.Sync()
	_, err = f.WriteString(`
    </div>
  </body>
</html>
`)
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
		}
	}
}
