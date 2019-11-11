package main

import (
	"flag"
	"github.com/radovskyb/watcher"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"time"
)

type ModTimeSorter []os.FileInfo

func (a ModTimeSorter) Len() int           { return len(a) }
func (a ModTimeSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ModTimeSorter) Less(i, j int) bool { return a[i].ModTime().Unix() > a[j].ModTime().Unix() }

func removeOldFiles(dir string, pattern *regexp.Regexp, keep int) {
	allFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var files []os.FileInfo

	for _, f := range allFiles {
		if pattern.MatchString(f.Name()) {
			files = append(files, f)
		}
	}

	sort.Sort(ModTimeSorter(files))

	for i, f := range files {
		if i < keep {
			continue
		}
		if err := os.Remove(path.Join(dir, f.Name())); err != nil {
			log.Fatal(err)
		}
	}
}


func main() {
	interval := flag.String("interval", "1s", "watcher poll interval")
	filePattern := flag.String("file-pattern", ".*", "regex to filter the files being watched")
	numberOfFiles := flag.Int("num-files", 3, "number of files to keep")
	flag.Parse()

	// Retrieve the list of files and folders.
	files := flag.Args()

	// If no files/folders were specified, watch the current directory.
	if len(files) == 0 {
		curDir, err := os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}
		files = append(files, curDir)
	}

	w := watcher.New()

	r, err := regexp.Compile(*filePattern)
	if err != nil {
		log.Fatalln(err)
	} else {
		w.AddFilterHook(watcher.RegexFilterHook(r, false))
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				if event.IsDir() {
					removeOldFiles(event.Path, r, *numberOfFiles)
				}
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	for _, file := range files {
		if err := w.Add(file); err != nil {
			log.Fatalln(err)
		}
	}

	// Parse the interval string into a time.Duration.
	parsedInterval, err := time.ParseDuration(*interval)
	if err != nil {
		log.Fatalln(err)
	}

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(parsedInterval); err != nil {
		log.Fatalln(err)
	}
}
