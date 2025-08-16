package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	check "github.com/andreimerlescu/checkfs"
	"github.com/andreimerlescu/checkfs/directory"
	"github.com/andreimerlescu/checkfs/file"
	"github.com/andreimerlescu/figtree/v2"
)

// newSummaryFilename returns summary.time.Now().UTC().Format(tFormat).md
var newSummaryFilename = func() string {
	return fmt.Sprintf("summary.%s.md", time.Now().UTC().Format(tFormat))
}

// callbackVerifyFile is a figtree WithCallback on the kFilename fig that uses checkfs.File to validate the file does NOT already exist
var callbackVerifyFile = func(value interface{}) error {
	return check.File(toString(value), file.Options{Exists: false})
}

// callbackVerifyReadableDirectory is a figtree WithCallback on the kSourceDir that uses checkfs.Directory to be More Permissive than 0444
var callbackVerifyReadableDirectory = func(value interface{}) error {
	return check.Directory(toString(value), directory.Options{Exists: true, MorePermissiveThan: 0444})
}

// toString uses figtree NewFlesh to return the ToString() value of the provided value argument
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

// capture assures that the d errors are not nil then runs terminate to write to os.Stderr
var capture = func(msg string, d ...error) {
	if len(d) == 0 || (len(d) == 1 && d[0] == nil) {
		return
	}
	terminate(os.Stderr, "[EXCUSE ME, BUT] %s\n\ncaptured error: %v\n", msg, d)
}

// terminate can write to os.Stderr or os.Stdout with a fmt.Fprintf format as i and a variadic interface of e that gets rendered to d either in plain text or as JSON
var terminate = func(d io.Writer, i string, e ...interface{}) {
	for _, f := range os.Args {
		if strings.HasPrefix(f, "-json") {
			mm := M{Message: fmt.Sprintf(i, e...)}
			jb, err := json.MarshalIndent(mm, "", "  ")
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error serializing json: %v\n", err)
				_, _ = fmt.Fprintf(d, i, e...)
			} else {
				fmt.Println(string(jb))
			}
			os.Exit(1)
		}
	}
	_, _ = fmt.Fprintf(d, i, e...)
	os.Exit(1)
}
