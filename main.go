package main

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	check "github.com/andreimerlescu/checkfs"
	"github.com/andreimerlescu/checkfs/directory"
	"github.com/andreimerlescu/checkfs/file"
	"github.com/andreimerlescu/figtree/v2"
	"github.com/andreimerlescu/sema"
)

//go:embed VERSION
var versionBytes embed.FS

var currentVersion string

func Version() string {
	if len(currentVersion) == 0 {
		versionBytes, err := versionBytes.ReadFile("VERSION")
		if err != nil {
			return ""
		}
		currentVersion = strings.TrimSpace(string(versionBytes))
	}
	return currentVersion
}

const (
	projectName   string = "github.com/andreimerlescu/summarize"
	tFormat       string = "2006.01.02.15.04.05.UTC"
	eConfigFile   string = "SUMMARIZE_CONFIG_FILE"
	kSourceDir    string = "d"
	kOutputDir    string = "o"
	kIncludeExt   string = "i"
	kExcludeExt   string = "x"
	kSkipContains string = "s"
	kFilename     string = "f"
	kVersion      string = "v"
	kDotFiles     string = "ndf"
	kMaxFiles     string = "mf"
	kDebug        string = "debug"
)

var (
	// figs is a figtree of fruit for configurable command line arguments that bear fruit
	figs figtree.Plant

	// defaultExclude are the -exc list of extensions that will be skipped automatically
	defaultExclude = []string{
		"7z", "gz", "xz", "zstd", "bz", "bzip2", "zip", "part", // compressed files
		"crt", "key", "asc", "id_rsa", "id_dsa", "id_ed25519", // encryption files
		"log", "dll", "so", "bin", "exe", // executable binaries
		"jpg", "png", "mov", "mp3", "mp4", "heic", "avi", // media
		"ttf", "woff", "woff2", "otf", // fonts
	}

	// defaultInclude are the -inc list of extensions that will be included in the summary
	defaultInclude = []string{
		"go", "ts", "tf", "sh", "py", "js",
	}

	// defaultAvoid are the -avoid list of substrings in file path names to avoid in the summary
	defaultAvoid = []string{
		".min.js", ".min.css", ".git/", ".svn/", ".vscode/", ".vs/", ".idea/", "logs/", "secrets/",
		".venv/", "/site-packages", ".terraform/", "summaries/", "node_modules/",
	}
)

// newSummaryFilename returns summary.time.Now().UTC().Format(tFormat).md
var newSummaryFilename = func() string {
	return fmt.Sprintf("summary.%s.md", time.Now().UTC().Format(tFormat))
}

// init creates a new figtree with options to use CONFIG_FILE as a way of reading a YAML file while ignoring the env
func init() {
	figs = figtree.With(figtree.Options{
		Harvest:           9,
		IgnoreEnvironment: true,
		ConfigFile:        os.Getenv(eConfigFile),
	})
	// properties
	figs.NewString(kSourceDir, ".", "Absolute path of directory you want to summarize.")
	figs.NewString(kOutputDir, filepath.Join(".", "summaries"), fmt.Sprintf("Path of the directory to write the %s file to", newSummaryFilename()))
	figs.NewString(kFilename, newSummaryFilename(), "Output file of summary.md")
	figs.NewList(kIncludeExt, defaultInclude, "List of extensions to INCLUDE in summary.")
	figs.NewList(kExcludeExt, defaultExclude, "List of extensions to EXCLUDE in summary.")
	figs.NewList(kSkipContains, defaultAvoid, "List of path substrings if present to skip over full path.")
	figs.NewInt(kMaxFiles, 20, "Maximum number of files to process concurrently")
	figs.NewBool(kDotFiles, false, "Any path that is considered a dotfile can be included by setting this to true")
	figs.NewBool(kVersion, false, "Display current version of summarize")
	figs.NewBool(kDebug, false, "Enable debug mode")
	// validators
	figs.WithValidator(kSourceDir, figtree.AssureStringNotEmpty)
	figs.WithValidator(kOutputDir, figtree.AssureStringNotEmpty)
	figs.WithValidator(kFilename, figtree.AssureStringNotEmpty)
	figs.WithValidator(kMaxFiles, figtree.AssureIntInRange(1, 63339))
	// callbacks
	figs.WithCallback(kSourceDir, figtree.CallbackAfterVerify, callbackVerifyReadableDirectory)
	figs.WithCallback(kFilename, figtree.CallbackAfterVerify, callbackVerifyFile)
	figs.WithCallback(kOutputDir, figtree.CallbackAfterVerify, func(value interface{}) error {
		return check.Directory(toString(value), directory.Options{
			WillCreate: true,
			Create: directory.Create{
				Kind:     directory.IfNotExists,
				Path:     toString(value),
				FileMode: 0755,
			},
		})
	})
	capture(figs.Load())
}

func main() {
	if *figs.Bool(kVersion) {
		fmt.Println(Version())
		os.Exit(0)
	}

	var (
		data      map[string][]string // data is map[ext][]path of found files to summarize
		dataMutex = sync.RWMutex{}    // adding concurrency
		wg        = sync.WaitGroup{}
		throttler = sema.New(runtime.GOMAXPROCS(0))
	)

	// initialize the data map with all -inc extensions
	var errs []error
	data = make(map[string][]string)
	for _, inc := range *figs.List(kIncludeExt) {
		data[inc] = []string{}
	}

	// populate data with the kSourceDir files based on -inc -exc -avoid lists
	capture(filepath.Walk(*figs.String(kSourceDir), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err // return the error received
		}
		if !info.IsDir() {

			// get the filename
			filename := filepath.Base(path)

			if *figs.Bool(kDotFiles) {
				if strings.HasPrefix(filename, ".") {
					return nil // skip without error
				}
			}

			// check the -avoid list
			for _, avoidThis := range *figs.List(kSkipContains) {
				if strings.Contains(filename, avoidThis) {
					if *figs.Bool(kDebug) {
						fmt.Printf("ignoring %s\n", path)
					}
					return nil // skip without error
				}
				if strings.Contains(path, avoidThis) {
					if *figs.Bool(kDebug) {
						fmt.Printf("ignoring %s\n", path)
					}
					return nil // skip without error
				}
			}

			// get the extension
			ext := filepath.Ext(path)
			ext = strings.ToLower(ext)
			ext = strings.TrimPrefix(ext, ".")

			// check the -exc list
			for _, excludeThis := range *figs.List(kExcludeExt) {
				if strings.EqualFold(excludeThis, ext) {
					if *figs.Bool(kDebug) {
						fmt.Printf("ignoring %s\n", path)
					}
					return nil // skip without error
				}
			}

			// populate the -inc list in data
			if _, exists := data[ext]; exists {
				data[ext] = append(data[ext], path)
			}
		}

		// continue to the next file
		return nil
	}))

	if *figs.Bool(kDebug) {
		fmt.Println("data received: ")
		for ext, paths := range data {
			fmt.Printf("%s: %s\n", ext, strings.Join(paths, ", "))
		}
	}

	maxFileSemaphore := sema.New(*figs.Int(kMaxFiles))
	writeChan := make(chan []byte, 10240)
	writerWG := sync.WaitGroup{}
	writerWG.Add(1)
	go func() {
		defer writerWG.Done()

		// Create output file
		outputFileName := filepath.Join(*figs.String(kOutputDir), *figs.String(kFilename))
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf("# Project Summary - %s\nGenerated by %s %s\n\n## Workspace\n\n<pre>%s</pre>\n\n\n",
			filepath.Base(*figs.String(kFilename)), projectName, Version(), *figs.String(kSourceDir)))

		for data := range writeChan {
			buf.Write(data)
		}

		capture(os.WriteFile(outputFileName, buf.Bytes(), 0644))
	}()

	for ext, paths := range data { // range over data to get ext and paths
		throttler.Acquire() // throttler is used to protect the runtime from excessive use
		wg.Add(1)           // wg is used to prevent the runtime from exiting early
		go func(ext string, paths []string) { // run this extension in a goroutine
			defer throttler.Release() // when we're done, release the throttler
			defer wg.Done()           // then tell the sync.WaitGroup that we are done
			sort.Strings(paths)       // sort the paths we receive
			dataMutex.Lock()          // lock the data map
			data[ext] = paths         // write the sorted paths
			dataMutex.Unlock()        // unlock the map

			// process each file in the ext list (one ext per throttle slot in the semaphore)
			for _, filePath := range paths {
				maxFileSemaphore.Acquire()
				wg.Add(1)
				go func(ext, filePath string) {
					defer maxFileSemaphore.Release()                               // maxFileSemaphore prevents excessive files from being opened
					defer wg.Done()                                                // keep the main thread running while this file is being processed
					var sb bytes.Buffer                                            // capture what we write to file in a bytes buffer
					sb.WriteString(fmt.Sprintf("## %s\n\n```%s\n", filePath, ext)) // write the header of the summary for the file
					content, err := os.ReadFile(filePath)                          // open the file and get its contents
					if err != nil {
						errs = append(errs, fmt.Errorf("Error reading file %s: %v\n", filePath, err))
						return
					}
					if _, writeErr := sb.Write(content); writeErr != nil {
						errs = append(errs, fmt.Errorf("Error writing file %s: %v\n", filePath, err))
						return
					}
					content = []byte{}        // clear memory after its written
					sb.WriteString("\n```\n") // close out the file footer
					writeChan <- sb.Bytes()
				}(ext, filePath)
			}

		}(ext, paths)
	}

	wg.Wait()        // wait for all files to finish processing
	close(writeChan) // Signal the writer goroutine to finish
	writerWG.Wait()  // Wait for the writer to flush and close the file

	if len(errs) > 0 {
		terminate(os.Stderr, "Error writing to output file: %v\n", errs)
	}

	// Print completion message
	fmt.Printf("Summary generated: %s\n",
		filepath.Join(*figs.String(kOutputDir), *figs.String(kFilename)))
}

var callbackVerifyFile = func(value interface{}) error {
	return check.File(toString(value), file.Options{Exists: false})
}

var callbackVerifyReadableDirectory = func(value interface{}) error {
	return check.Directory(toString(value), directory.Options{Exists: true, MorePermissiveThan: 0444})
}

var toString = func(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case *string:
		return *v
	default:
		return ""
	}
}

var capture = func(d ...error) {
	if len(d) == 0 || (len(d) == 1 && d[0] == nil) {
		return
	}
	terminate(os.Stderr, "captured error: %v\n", d)
}

var terminate = func(d io.Writer, i string, e ...interface{}) {
	_, _ = fmt.Fprintf(d, i, e...)
	os.Exit(1)
}
