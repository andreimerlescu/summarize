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
	"slices"
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
	projectName          string = "github.com/andreimerlescu/summarize"
	tFormat              string = "2006.01.02.15.04.05.UTC"
	eConfigFile          string = "SUMMARIZE_CONFIG_FILE"
	eAddIgnoreInPathList string = "SUMMARIZE_IGNORE_CONTAINS"
	eAddIncludeExtList   string = "SUMMARIZE_INCLUDE_EXT"
	eAddExcludeExtList   string = "SUMMARIZE_EXCLUDE_EXT"
	kSourceDir           string = "d"
	kOutputDir           string = "o"
	kIncludeExt          string = "i"
	kExcludeExt          string = "x"
	kSkipContains        string = "s"
	kFilename            string = "f"
	kVersion             string = "v"
	kDotFiles            string = "ndf"
	kMaxFiles            string = "mf"
	kDebug               string = "debug"
)

var (
	// figs is a figtree of fruit for configurable command line arguments that bear fruit
	figs figtree.Plant

	// defaultExclude are the -exc list of extensions that will be skipped automatically
	defaultExclude = []string{
		// Compressed archives
		"7z", "gz", "xz", "zst", "zstd", "bz", "bz2", "bzip2", "zip", "tar", "rar", "lz4", "lzma", "cab", "arj",

		// Encryption, certificates, and sensitive keys
		"crt", "cert", "cer", "key", "pub", "asc", "pem", "p12", "pfx", "jks", "keystore",
		"id_rsa", "id_dsa", "id_ed25519", "id_ecdsa", "gpg", "pgp",

		// Binary & executable artifacts
		"exe", "dll", "so", "dylib", "bin", "out", "o", "obj", "a", "lib", "dSYM",
		"class", "pyc", "pyo", "__pycache__",
		"jar", "war", "ear", "apk", "ipa", "dex", "odex",
		"wasm", "node", "beam", "elc",

		// System and disk images
		"iso", "img", "dmg", "vhd", "vdi", "vmdk", "qcow2",

		// Database files
		"db", "sqlite", "sqlite3", "db3", "mdb", "accdb", "sdf", "ldb",

		// Log files
		"log", "trace", "dump", "crash",

		// Media files - Images
		"jpg", "jpeg", "png", "gif", "bmp", "tiff", "tif", "webp", "ico", "svg", "heic", "heif", "raw", "cr2", "nef", "dng",

		// Media files - Audio
		"mp3", "wav", "flac", "aac", "ogg", "wma", "m4a", "opus", "aiff",

		// Media files - Video
		"mp4", "avi", "mov", "mkv", "webm", "flv", "wmv", "m4v", "3gp", "ogv",

		// Font files
		"ttf", "otf", "woff", "woff2", "eot", "fon", "pfb", "pfm",

		// Document formats (typically not source code)
		"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "odt", "ods", "odp", "rtf",

		// IDE/Editor/Tooling artifacts
		"suo", "sln", "user", "ncb", "pdb", "ipch", "ilk", "tlog", "idb", "aps", "res",
		"iml", "idea", "vscode", "project", "classpath", "factorypath", "prefs",
		"vcxproj", "vcproj", "filters", "xcworkspace", "xcuserstate", "xcscheme", "pbxproj",
		"DS_Store", "Thumbs.db", "desktop.ini",

		// Package manager and build artifacts
		"lock", "sum", "resolved", // package-lock.json, go.sum, yarn.lock, etc.

		// Temporary and backup files
		"tmp", "temp", "swp", "swo", "bak", "backup", "orig", "rej", "patch",
		"~", "old", "new", "part", "incomplete",

		// Source maps and minified files (usually generated)
		"map", "min.js", "min.css", "bundle.js", "bundle.css", "chunk.js",

		// Configuration that's typically binary or generated
		"dat", "data", "cache", "pid", "sock",

		// Version control artifacts (though usually in ignored directories)
		"pack", "idx", "rev",

		// Other binary formats
		"pickle", "pkl", "npy", "npz", "mat", "rdata", "rds",
	}

	// defaultInclude are the -inc list of extensions that will be included in the summary
	defaultInclude = []string{
		"go", "ts", "tf", "sh", "py", "js", "Makefile", "mod", "Dockerfile", "dockerignore", "gitignore", "esconfigs", "md",
	}

	// defaultAvoid are the -avoid list of substrings in file path names to avoid in the summary
	defaultAvoid = []string{
		".min.js", ".min.css", ".git/", ".svn/", ".vscode/", ".vs/", ".idea/", "logs/", "secrets/",
		".venv/", "/site-packages", ".terraform/", "summaries/", "node_modules/", "/tmp", "tmp/", "logs/",
	}
)

// newSummaryFilename returns summary.time.Now().UTC().Format(tFormat).md
var newSummaryFilename = func() string {
	return fmt.Sprintf("summary.%s.md", time.Now().UTC().Format(tFormat))
}

// init creates a new figtree with options to use CONFIG_FILE as a way of reading a YAML file while ignoring the env
func configure() {
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
}

type result struct {
	Path     string `yaml:"path" json:"path"`
	Contents string `yaml:"contents" json:"contents"`
	Size     int64  `yaml:"size" json:"size"`
}

func main() {
	configure()
	capture("figs loading environment", figs.Load())
	isDebug := *figs.Bool(kDebug)
	if *figs.Bool(kVersion) {
		fmt.Println(Version())
		os.Exit(0)
	}

	var (
		lIncludeExt   = *figs.List(kIncludeExt)
		lExcludeExt   = *figs.List(kExcludeExt)
		lSkipContains = *figs.List(kSkipContains)

		sourceDir = *figs.String(kSourceDir)
		outputDir = *figs.String(kOutputDir)
	)

	capture("checking output directory", check.Directory(outputDir, directory.Options{
		WillCreate: true,
		Create: directory.Create{
			Kind:     directory.IfNotExists,
			Path:     outputDir,
			FileMode: 0755,
		},
	}))

	addFromEnv(eAddIgnoreInPathList, &lSkipContains)
	addFromEnv(eAddIncludeExtList, &lIncludeExt)
	addFromEnv(eAddExcludeExtList, &lExcludeExt)

	var (
		wg        = sync.WaitGroup{}
		throttler = sema.New(runtime.GOMAXPROCS(0))
	)

	// initialize the data map with all -inc extensions
	var errs []error

	type mapData struct {
		Ext   string
		Paths []string
	}

	data := &sync.Map{}
	for _, inc := range lIncludeExt {
		data.Store(inc, mapData{
			Ext:   inc,
			Paths: []string{},
		})
	}

	// populate data with the kSourceDir files based on -inc -exc -avoid lists
	capture("walking source directory", filepath.Walk(sourceDir, func(path string, info fs.FileInfo, err error) error {
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
			for _, avoidThis := range lSkipContains {
				a := strings.Contains(filename, avoidThis) || strings.Contains(path, avoidThis)
				b := strings.HasPrefix(filename, avoidThis) || strings.HasPrefix(path, avoidThis)
				c := strings.HasSuffix(filename, avoidThis) || strings.HasSuffix(path, avoidThis)
				parts, err := filepath.Glob(path)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				for i := 0; i < len(parts); i++ {
					part := parts[i]
					if part == "/" {
						continue
					}
					if strings.Contains(part, avoidThis) || strings.HasPrefix(part, avoidThis) || strings.HasSuffix(part, avoidThis) {
						if isDebug {
							fmt.Printf("skipping file %q\n", part)
						}
						return nil
					}
				}
				if a || b || c {
					if isDebug {
						fmt.Printf("ignoring %s in %s\n", filename, path)
					}
					return nil // skip without error
				}
			}

			// get the extension
			ext := filepath.Ext(path)
			ext = strings.ToLower(ext)
			ext = strings.TrimPrefix(ext, ".")

			if isDebug {
				fmt.Printf("ext: %s\n", ext)
			}

			// check the -exc list
			for _, excludeThis := range lExcludeExt {
				if strings.EqualFold(excludeThis, ext) {
					if isDebug {
						fmt.Printf("ignoring %s\n", path)
					}
					return nil // skip without error
				}
			}

			var toUpdate []mapData
			// populate the -inc list in data
			data.Range(func(e any, p any) bool {
				key, ok := e.(string)
				if !ok {
					return true // continue
				}
				value, ok := p.(mapData)
				if !ok {
					return true
				}
				if strings.EqualFold(key, ext) {
					value.Ext = key
				}
				value.Paths = append(value.Paths, path)
				toUpdate = append(toUpdate, value)
				return true
			})
			for _, value := range toUpdate {
				data.Store(value.Ext, value)
			}
		}

		// continue to the next file
		return nil
	}))

	if isDebug {
		fmt.Println("data received: ")
		data.Range(func(e any, p any) bool {
			ext, ok := e.(string)
			if !ok {
				return true // continue
			}
			thisData, ok := p.(mapData)
			if !ok {
				return true // continue
			}
			fmt.Printf("%s: %s\n", ext, strings.Join(thisData.Paths, ", "))
			return true // continue
		})
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
		buf.WriteString("# Project Summary - " + filepath.Base(*figs.String(kFilename)) + "\n")
		buf.WriteString("Generated by " + projectName + " " + Version() + "\n\n")
		buf.WriteString("AI Instructions are the user requests that you analyze their project workspace ")
		buf.WriteString("as provided here by filename followed by the contents. You are to answer their ")
		buf.WriteString("question using the source code provided as the basis of your responses. You are to ")
		buf.WriteString("completely modify each individual file as per-the request and provide the completely ")
		buf.WriteString("updated form of the file. Do not abbreviate the file, and if the file is excessive in ")
		buf.WriteString("length, then print the entire contents in your response with your updates to the ")
		buf.WriteString("specific components while retaining all existing functionality and maintaining comments ")
		buf.WriteString("within the code.  \n\n")
		buf.WriteString("### Workspace\n\n")
		buf.WriteString("<pr>" + *figs.String(kSourceDir) + "</pre>\n\n\n")

		for receivedData := range writeChan {
			buf.Write(receivedData)
		}

		capture("saving output file during write", os.WriteFile(outputFileName, buf.Bytes(), 0644))
	}()

	var toUpdate []mapData

	seen := seenStrings{m: make(map[string]bool)}

	data.Range(func(e any, p any) bool {
		ext, ok := e.(string)
		if !ok {
			return true // continue
		}
		thisData, ok := p.(mapData)
		if !ok {
			return true // continue
		}
		paths := slices.Clone(thisData.Paths)

		throttler.Acquire() // throttler is used to protect the runtime from excessive use
		wg.Add(1)           // wg is used to prevent the runtime from exiting early
		go func(innerData *mapData, toUpdate *[]mapData, ext string, paths []string) { // run this extension in a goroutine
			defer throttler.Release() // when we're done, release the throttler
			defer wg.Done()           // then tell the sync.WaitGroup that we are done

			innerData.Paths = paths
			*toUpdate = append(*toUpdate, *innerData)

			// process each file in the ext list (one ext per throttle slot in the semaphore)
			for _, filePath := range paths {
				if seen.Exists(filePath) {
					continue
				}
				maxFileSemaphore.Acquire()
				wg.Add(1)
				go func(ext, filePath string) {
					defer maxFileSemaphore.Release() // maxFileSemaphore prevents excessive files from being opened
					defer wg.Done()                  // keep the main thread running while this file is being processed
					if strings.HasSuffix(filePath, ".DS_Store") ||
						strings.HasSuffix(filePath, ".exe") ||
						strings.HasSuffix(filePath, "-amd64") ||
						strings.HasSuffix(filePath, "-arm64") ||
						strings.HasSuffix(filePath, "aarch64") {
						return
					}
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
					seen.Add(filePath)
					writeChan <- sb.Bytes()
				}(ext, filePath)
			}

		}(&thisData, &toUpdate, ext, paths)
		return true
	})

	wg.Wait() // wait for all files to finish processing

	for _, innerData := range toUpdate {
		data.Store(innerData.Ext, innerData)
	}

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
		flesh := figtree.NewFlesh(value)
		f := fmt.Sprintf("%v", flesh.ToString())
		return f
	}
}

var capture = func(msg string, d ...error) {
	if len(d) == 0 || (len(d) == 1 && d[0] == nil) {
		return
	}
	terminate(os.Stderr, "%s\n\ncaptured error: %v\n", msg, d)
}

var terminate = func(d io.Writer, i string, e ...interface{}) {
	_, _ = fmt.Fprintf(d, i, e...)
	os.Exit(1)
}

func simplify(t []string) []string {
	seen := make(map[string]bool)
	for _, v := range t {
		seen[v] = true
	}
	results := make([]string, len(t))
	for i, v := range t {
		if seen[v] {
			results[i] = v
		}
	}
	return results
}
func addFromEnv(e string, l *[]string) {
	v, ok := os.LookupEnv(e)
	if ok {
		flesh := figtree.NewFlesh(v)
		maybeAdd := flesh.ToList()
		for _, entry := range maybeAdd {
			*l = append(*l, entry)
		}
	}
	*l = simplify(*l)
}

type seenStrings struct {
	mu sync.RWMutex
	m  map[string]bool
}

func (s *seenStrings) Add(entry string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[entry] = true
}
func (s *seenStrings) Remove(entry string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, entry)
}

func (s *seenStrings) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m)
}

func (s *seenStrings) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprint(s.m)
}

func (s *seenStrings) True(entry string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[entry] = true
}

func (s *seenStrings) False(entry string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, entry)
}

func (s *seenStrings) Exists(entry string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.m[entry]
}
