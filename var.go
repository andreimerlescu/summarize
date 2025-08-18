package main

import (
	"sync"

	"github.com/andreimerlescu/figtree/v2"
	"github.com/andreimerlescu/sema"
)

var (
	// figs is a figtree of fruit for configurable command line arguments that bear fruit
	figs figtree.Plant

	defaultExclude = []string{
		"useExpanded",
	}
	// defaultExclude are the -exc list of extensions that will be skipped automatically
	extendedDefaultExclude = []string{
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
		"useExpanded",
	}

	extendedDefaultInclude = []string{
		"go", "ts", "tf", "sh", "py", "js", "Makefile", "mod", "Dockerfile", "dockerignore", "gitignore", "esconfigs", "md",
	}

	defaultAvoid = []string{
		"useExpanded",
	}

	// extendedDefaultAvoid are the -avoid list of substrings in file path names to avoid in the summary
	extendedDefaultAvoid = []string{
		".min.js", ".min.css", ".git/", ".svn/", ".vscode/", ".vs/", ".idea/", "logs/", "secrets/",
		".venv/", "/site-packages", ".terraform/", "summaries/", "node_modules/", "/tmp", "tmp/", "logs/",
	}

	data                                                   *sync.Map
	isDebug                                                bool
	sourceDir                                              string
	outputDir                                              string
	inc, exc, ski, lIncludeExt, lExcludeExt, lSkipContains []string
	errs                                                   []error
	toUpdate                                               []mapData
	wg, writerWG                                           *sync.WaitGroup
	throttler, maxFileSemaphore                            sema.Semaphore
	seen                                                   *seenStrings
	resultsChan                                            chan Result
)
