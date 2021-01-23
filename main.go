package main

import (
	"fmt"
	filepath "github.com/mattn/go-zglob"
	"os"
	"regexp"
	"strings"
	"sync"
	//"path/filepath"
	"flag"
	"github.com/loveleshsharma/gohive"
	"io/ioutil"
)

var (
	filePattern string
	files       []string
	pattern     string
	findFile    bool
	isReg       bool
	reg         *regexp.Regexp
)

func collectFiles() {
	var err error
	files, err = filepath.Glob(filePattern)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func showFiles() {
	for _, f := range files {
		fmt.Println(f)
	}
}

func showFinds() {
	hive := gohive.NewFixedSizePool(5)
	wg := new(sync.WaitGroup)
	regex := regexp.MustCompile(`\.(exe|bin|png|jpg|mod|sum|lock)$`)
	for _, f := range files {
		if regex.MatchString(f) {
			continue
		}
		exe := func() {
			defer wg.Done()
			search(f)
		}
		wg.Add(1)
		hive.Submit(exe)
	}
	wg.Wait()
}

var liner = regexp.MustCompile(`\n|\r`)

func search(fname string) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return
	}
	lines := liner.Split(string(data), -1)
	for i, s := range lines {
		if isReg {
			if reg.MatchString(s) {
				fmt.Printf("%s:%d:\n%s\n", fname, i+1, s)
			}
		} else {
			if strings.Contains(strings.ToLower(s), pattern) {
				fmt.Printf("%s:%d:\n%s\n", fname, i+1, s)
			}
		}
	}
}

func helpMsg() {
	fmt.Println(`file / content finder
	usage:
	find [-re] <path glob> [match pattern]
	given one argument, shows files matching the glob pattern
	given two arguments, finds given text in files matchinmg the glob pattern
	
	text search is not glob, it checks if lines contain the keyword
	but you can pass the -re flag to do a regex search instead
	
	-re does not affect file filtering
	-re is case insensitive
	`)
}

var help bool

func main() {
	flag.BoolVar(&isReg, "re", false, "use a regex pattern instead")
	flag.BoolVar(&help, "h", false, "show usage")
	flag.Parse()
	if help {
		helpMsg()
		return
	}
	args := flag.Args()
	if len(args) == 0 {
		flag.PrintDefaults()
		return
	}
	if len(args) == 1 {
		findFile = true
	}
	filePattern = args[0]
	collectFiles()
	if findFile {
		showFiles()
		return
	}
	if isReg {
		temp, err := regexp.Compile("(?i)" + args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		reg = temp
	} else {
		pattern = strings.ToLower(args[1])
	}
	showFinds()
}
