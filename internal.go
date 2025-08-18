package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
)

// analyze is called from investigate where the path is inspected and the resultsChan is written to
func analyze(ext, filePath string) {
	defer maxFileSemaphore.Release() // maxFileSemaphore prevents excessive files from being opened
	defer wg.Done()                  // keep the main thread running while this file is being processed
	if strings.HasSuffix(filePath, ".DS_Store") ||
		strings.HasSuffix(filePath, ".exe") ||
		strings.HasSuffix(filePath, "-amd64") ||
		strings.HasSuffix(filePath, "-arm64") ||
		strings.HasSuffix(filePath, "aarch64") {
		return
	}
	type tFileInfo struct {
		Name string      `json:"name"`
		Size int64       `json:"size"`
		Mode os.FileMode `json:"mode"`
	}
	info, err := os.Stat(filePath)
	if err != nil {
		errs = append(errs, err)
		return
	}
	fileInfo := &tFileInfo{
		Name: filepath.Base(filePath),
		Size: info.Size(),
		Mode: info.Mode(),
	}
	infoJson, err := json.MarshalIndent(fileInfo, "", "  ")
	if err != nil {
		errs = append(errs, err)
		return
	}
	var sb bytes.Buffer // capture what we write to file in a bytes buffer
	sb.WriteString("## " + filepath.Base(filePath) + "\n\n")
	sb.WriteString("The `os.Stat` for the " + filePath + " is: \n\n")
	sb.WriteString("```json\n")
	sb.WriteString(string(infoJson) + "\n")
	sb.WriteString("```\n\n")
	sb.WriteString("Source Code:\n\n")
	sb.WriteString("```" + ext + "\n")
	content, err := os.ReadFile(filePath) // open the file and get its contents
	if err != nil {
		errs = append(errs, fmt.Errorf("Error reading file %s: %v\n", filePath, err))
		return
	}
	if _, writeErr := sb.Write(content); writeErr != nil {
		errs = append(errs, fmt.Errorf("Error writing file %s: %v\n", filePath, err))
		return
	}
	content = []byte{}          // clear memory after its written
	sb.WriteString("\n```\n\n") // close out the file footer
	seen.Add(filePath)
	resultsChan <- Result{
		Path:     filePath,
		Contents: sb.Bytes(),
		Size:     int64(sb.Len()),
	}
}

// done is responsible for printing the results to STDOUT when the summarize program is finished
func done() {
	// Print completion message
	if *figs.Bool(kJson) {
		r := M{
			Message: fmt.Sprintf("Summary generated: %s\n",
				filepath.Join(*figs.String(kOutputDir), *figs.String(kFilename)),
			),
		}
		jb, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			terminate(os.Stderr, "Error marshalling results: %v\n", err)
		} else {
			fmt.Println(string(jb))
		}
	} else {
		fmt.Printf("Summary generated: %s\n",
			filepath.Join(*figs.String(kOutputDir), *figs.String(kFilename)),
		)
	}
}

// iterate is a func that gets passed directly into the data.Range(iterate) that runs investigate concurrently with the
// wg and throttler enabled
func iterate(e, p any) bool {
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
	go investigate(&thisData, &toUpdate, ext, paths)
	return true
}

// investigate is called from iterate where it takes an extension and a slice of paths to analyze each path
func investigate(innerData *mapData, toUpdate *[]mapData, ext string, paths []string) {
	defer throttler.Release() // when we're done, release the throttler
	defer wg.Done()           // then tell the sync.WaitGroup that we are done

	paths = simplify(paths)

	innerData.Paths = paths
	*toUpdate = append(*toUpdate, *innerData)

	// process each file in the ext list (one ext per throttle slot in the semaphore)
	for _, filePath := range paths {
		if seen.Exists(filePath) {
			continue
		}
		maxFileSemaphore.Acquire()
		wg.Add(1)
		go analyze(ext, filePath)
	}
}

// populate is responsible for loading new paths into the *sync.Map called data
func populate(ext, path string) {
	todo := make([]mapData, 0)
	if data == nil {
		panic("data is nil")
	}
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
		todo = append(todo, value)
		return true
	})
	for _, value := range todo {
		data.Store(value.Ext, value)
	}
}

// receive will accept Result from resultsChan and write to the summary file. If `-chat` is enabled, the StartChat will
// get called. Once the chat session is completed, the contents of the chat log is injected into the summary file.
func receive() {
	if writerWG == nil {
		panic("writer wg is nil")
	}
	defer writerWG.Done()

	// Create output file
	srcDir := *figs.String(kSourceDir)
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
	abs, err := filepath.Abs(srcDir)
	if err == nil {
		buf.WriteString("`" + abs + "`\n\n")
	} else {
		buf.WriteString("`" + srcDir + "`\n\n")
	}

	renderMu := &sync.Mutex{}
	renderedPaths := make(map[string]int64)
	totalSize := int64(buf.Len())
	for in := range resultsChan {
		if _, exists := renderedPaths[in.Path]; exists {
			continue
		}
		runningSize := atomic.AddInt64(&totalSize, in.Size)
		if runningSize >= *figs.Int64(kMaxOutputSize) {
			continue
		}
		renderMu.Lock()
		renderedPaths[in.Path] = in.Size
		buf.Write(in.Contents)
		renderMu.Unlock()
	}

	if *figs.Bool(kChat) {
		StartChat(&buf)
		path := latestChatLog()
		contents, err := os.ReadFile(path)
		if err == nil {
			old := buf.String()
			buf.Reset()
			buf.WriteString("## Chat Log \n\n")
			body := string(contents)
			body = strings.ReplaceAll(body, "You: ", "\n### ")
			buf.WriteString(body)
			buf.WriteString("\n\n")
			buf.WriteString("## Summary \n\n")
			buf.WriteString(old)
		}
	}
	render(&buf, outputFileName)
}

// render will take the summary and either write it to a file, STDOUT or present an error to STDERR
func render(buf *bytes.Buffer, outputFileName string) {
	shouldPrint := *figs.Bool(kPrint)
	canWrite := *figs.Bool(kWrite)
	showJson := *figs.Bool(kJson)
	wrote := false

	if *figs.Bool(kCompress) {
		compressed, err := compress(bytes.Clone(buf.Bytes()))
		capture("compressing bytes buffer", err)
		buf.Reset()
		buf.Write(compressed)
		outputFileName += ".gz"
	}

	if !shouldPrint && !canWrite {
		capture("saving output file during write", os.WriteFile(outputFileName, buf.Bytes(), 0644))
		wrote = true
	}

	if canWrite && !wrote {
		capture("saving output file during write", os.WriteFile(outputFileName, buf.Bytes(), 0644))
		wrote = true
	}

	if shouldPrint {
		if showJson {
			r := Final{
				Path:     outputFileName,
				Size:     int64(buf.Len()),
				Contents: buf.String(),
			}
			jb, err := json.MarshalIndent(r, "", "  ")
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
			fmt.Println(string(jb))
		} else {
			fmt.Println(buf.String())
		}
		os.Exit(0)

	}
}

// summarize walks through a filepath recursively and matches paths that get stored inside the
// data *sync.Map for the extension.
func summarize(path string, info fs.FileInfo, err error) error {
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
			if a || b || c {
				if isDebug {
					fmt.Printf("ignoring %s in %s\n", filename, path)
				}
				return nil // skip without error
			}

			parts, err := filepath.Glob(path)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			for i := 0; i < len(parts); i++ {
				part := parts[i]
				if strings.EqualFold(part, string(os.PathSeparator)) {
					continue
				}
				if strings.Contains(part, avoidThis) || strings.HasPrefix(part, avoidThis) || strings.HasSuffix(part, avoidThis) {
					if isDebug {
						fmt.Printf("skipping file %q\n", part)
					}
					return nil
				}
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
		populate(ext, path)
	}

	// continue to the next file
	return nil
}
