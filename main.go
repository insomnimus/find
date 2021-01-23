package main

import (
	"fmt"
	filepath "github.com/mattn/go-zglob"
	"os"
	"regexp"
	"strings"
	//"path/filepath"
	"flag"
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
