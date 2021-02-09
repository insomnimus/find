package main

import (
	"bufio"
	"flag"
	"fmt"
	globber "github.com/mattn/go-zglob"
	"os"
	"regexp"
	"strings"
)

var (
	method, help bool
	filePattern  string
	files        []string
	pattern      string
	findFile     bool
	isReg        bool
	goStruct     bool
	reg          *regexp.Regexp
)

func collectFiles() {
	var err error
	files, err = globber.Glob(filePattern)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if method || goStruct {
		filterGoFiles()
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

func filterGoFiles() {
	goer := regexp.MustCompile(`.+\.go$`)
	var temp []string
	for _, f := range files {
		if goer.MatchString(f) {
			temp = append(temp, f)
		}
	}
	files = temp
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
	-gs (search a go struct by name, print all of its body)
	
	text search is not glob, it checks if lines contain the keyword
	but you can pass the -re flag to do a regex search instead
	
	all pattern matching is case insensitive
	`)
}

var commenter = regexp.MustCompile(`^\s*\/\/`)

func searchGoStruct(fname string) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(f)
	i := 0
	text := ""
	var cC, oC int
	var msg string
	for scanner.Scan() {
		text = scanner.Text()
		i++
		if commenter.MatchString(text) {
			continue
		}
		if reg.MatchString(text) {
			//fmt.Println(text)
			msg += fmt.Sprintf("%s:%d:\n"+
				"\t%s\n", fname, i, text)
			oC += strings.Count(text, "{")
			cC += strings.Count(text, "}")
			if cC == oC {
				continue
			}

			for scanner.Scan() {
				i++
				text = scanner.Text()
				if commenter.MatchString(text) {
					continue
				}
				oC += strings.Count(text, "{")
				cC += strings.Count(text, "}")
				msg += fmt.Sprintf("\t%s\n", text)
				if cC == oC {
					break
				}
			}
			fmt.Println(msg)
			msg = ""
			oC = 0
			cC = 0
		}

	}
}

func main() {
	flag.BoolVar(&isReg, "re", false, "use a regex pattern instead")
	flag.BoolVar(&help, "h", false, "show usage")
	flag.BoolVar(&goStruct, "gs", false, "search for a go struct")
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
	} else if goStruct {
		var err error
		reg, err = regexp.Compile(`(?i)[\s]*type[\s]+[a-zA-Z]?[a-zA-Z0-9]*` + args[1] + `[a-zA-Z0-9]*[\s]+struct[\s]*\{`)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		isReg = true
	} else if isReg && !method && !goStruct {
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
	if goStruct {
		for j := range jobs {
			searchGoStruct(j)
			results <- true
		}
		return
	}
	for j := range jobs {
		search(j)
		results <- true
	}
}
