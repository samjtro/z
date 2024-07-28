package main

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

var (
	pathBase string
)

type DirTree []*Dir
type FileTree []*File

type Dir struct {
	DirTree  DirTree
	FileTree FileTree
	Path     string
}

type File struct {
	Name string
	Path string
}

type QueryResults []QueryResult

type QueryResult struct {
	Path       string
	Components []string
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

/* check if dirs contains subdirs
func hasSubDirs(dirs []fs.DirEntry) bool {
	truth := false
	for _, dir := range dirs {
		if dir.IsDir() {
			truth = true
		}
	}
	return truth
}*/

// walk a dir
func (d *Dir) walkDir() {
	dirs, err := os.ReadDir(d.Path)
	check(err)
	for _, dir := range dirs {
		//prefix := strings.Repeat("  ", depth)
		if dir.IsDir() && dir.Name()[0] != '.' {
			if dir.Name() != "node_modules" {
				newD := &Dir{DirTree: d.DirTree, Path: fmt.Sprintf("%s/%s", d.Path, dir.Name())}
				newD.walkDir()
				d.DirTree = append(d.DirTree, newD)
			}
		} else if dir.Name()[0] != '.' {
			d.FileTree = append(d.FileTree, &File{
				Name: dir.Name(),
				Path: fmt.Sprintf("%s/%s", d.Path, dir.Name()),
			})
		}
	}
}

// query a dir
func (d *Dir) getFilesFromDir() FileTree {
	var ft FileTree
	for _, x := range d.DirTree {
		ft = append(ft, x.FileTree...)
	}
	return ft
}

// query the zets dir
func (d DirTree) query(q []string) QueryResults {
	query := strings.Join(q, " ")
	var ft FileTree
	var qr QueryResults
	for _, x := range d {
		for _, y := range x.DirTree {
			ft = append(ft, y.getFilesFromDir()...)
		}
	}
	for _, x := range ft {
		t := false
		// credit: https://www.scaler.com/topics/golang/golang-read-file-line-by-line/
		readFile, err := os.Open(x.Path)
		check(err)
		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)
		var fileLines []string
		for fileScanner.Scan() {
			fileLines = append(fileLines, fileScanner.Text())
		}
		readFile.Close()
		for _, y := range fileLines {
			if fuzzy.Match(query, y) {
				t = true
			}
		}
		if t {
			qr = append(qr, QueryResult{
				Path:       x.Path,
				Components: fuzzy.Find(query, fileLines),
			})
		}
	}
	return qr
}

func walkZetDir() DirTree {
	dirs, err := os.ReadDir(pathBase)
	check(err)
	var dt DirTree
	for _, dir := range dirs {
		var d *Dir
		if dir.IsDir() && dir.Name()[0] != '.' {
			d = &Dir{Path: fmt.Sprintf("%s/%s", pathBase, dir.Name())}
			d.walkDir()
		}
		dt = append(dt, d)
	}
	return dt
}

// generate index.html
func generateIndex() {
	f, err := os.Create(fmt.Sprintf("%s/index.html", pathBase))
	check(err)
	template.New("tmpl.gohtml").Execute(f, walkZetDir())
	check(f.Close())
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
			dt := walkZetDir()
			fmt.Println(dt.query(os.Args[2:]))
		}
	}
}
