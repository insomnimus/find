package main

import (
	"fmt"
	filepath "github.com/mattn/go-zglob"
	"os"
	"regexp"
	"strings"
	//"path/filepath"
	"bufio"
	"flag"
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

func filterFiles() {
	regex := regexp.MustCompile(`\.(exe|bin|png|jpg|mod|sum|lock)$`)
	tempFiles := files
	files = []string{}
	for _, f := range tempFiles {
		if regex.MatchString(f) {
			continue
		}
		files = append(files, f)
	}
}

func search(fname string) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(f)
	i := 0
	s := ""
	if isReg {
		for scanner.Scan() {
			i++
			s = scanner.Text()
			if reg.MatchString(s) {
				fmt.Printf("%s:%d:\n%s\n", fname, i, s)
			}
		}
		return
	}
	for scanner.Scan() {
		i++
		s = scanner.Text()
		if strings.Contains(strings.ToLower(s), pattern) {
			fmt.Printf("%s:%d:\n%s\n", fname, i, s)
		}
	}
}

func helpMsg() {
	fmt.Println(`file / content finder
	usage:
	find [options] <path glob> [match pattern]
	given one argument, shows files matching the glob pattern
	given two arguments, finds given text in files matching the glob pattern
	
	options are:
	-re (do a regex search)
	-gm (search a go method by name, the name is also part of the regex so you can extend it)
	text search is not glob, it checks if lines contain the keyword
	but you can pass the -re flag to do a regex search instead
	
	all pattern matching is case insensitive
	`)
}

var help bool

var method bool

func main() {
	flag.BoolVar(&isReg, "re", false, "use a regex pattern instead")
	flag.BoolVar(&help, "h", false, "show usage")
	flag.BoolVar(&method, "gm", false, "search for a go method by the name")
	flag.Parse()
	if help {
		helpMsg()
		return
	}
	args := flag.Args()
	if len(args) == 0 {
		helpMsg()
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
	if method {
		isReg = true
		var err error
		reg, err = regexp.Compile(`(?i)^func\s?\([^\)]+\)[\s]*` + args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	} else if isReg && !method {
		temp, err := regexp.Compile("(?i)" + args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		reg = temp
	} else {
		pattern = strings.ToLower(args[1])
	}
	filterFiles()
	numberJobs := len(files)
	jobs := make(chan string, numberJobs)
	results := make(chan bool, numberJobs)
	workerCount := 5
	if workerCount > len(files) {
		workerCount = len(files)
	}
	for i := 0; i < workerCount; i++ {
		go worker(jobs, results)
	}
	for _, f := range files {
		jobs <- f
	}
	close(jobs)
	for i := 0; i < numberJobs; i++ {
		<-results
	}
}

func worker(jobs <-chan string, results chan<- bool) {
	for j := range jobs {
		search(j)
		results <- true
	}
}
