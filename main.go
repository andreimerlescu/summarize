package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	check "github.com/andreimerlescu/checkfs"
	"github.com/andreimerlescu/checkfs/directory"
	"github.com/andreimerlescu/sema"
)

func main() {
	configure()
	capture("figs loading environment", figs.Load())

	inc := *figs.List(kIncludeExt)
	if len(inc) == 1 && inc[0] == "useExpanded" {
		figs.StoreList(kIncludeExt, extendedDefaultInclude)
	}
	fmt.Println(strings.Join(inc, ", "))
	exc := *figs.List(kExcludeExt)
	if len(exc) == 1 && exc[0] == "useExpanded" {
		figs.StoreList(kExcludeExt, extendedDefaultExclude)
	}
	fmt.Println(strings.Join(exc, ", "))
	ski := *figs.List(kSkipContains)
	if len(ski) == 1 && ski[0] == "useExpanded" {
		figs.StoreList(kSkipContains, extendedDefaultAvoid)
	}
	fmt.Println(strings.Join(ski, ", "))

	if *figs.Bool(kShowExpanded) {
		fmt.Println("Expanded:")
		fmt.Printf("-%s=%s\n", kIncludeExt, strings.Join(*figs.List(kIncludeExt), ","))
		fmt.Printf("-%s=%s\n", kExcludeExt, strings.Join(*figs.List(kExcludeExt), ","))
		fmt.Printf("-%s=%s\n", kSkipContains, strings.Join(*figs.List(kSkipContains), ","))
		os.Exit(0)
	}
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
	resultsChan := make(chan Result, *figs.Int(kMaxFiles))
	writerWG := sync.WaitGroup{}
	writerWG.Add(1)
	go func() {
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

		throttler.Acquire()                                                            // throttler is used to protect the runtime from excessive use
		wg.Add(1)                                                                      // wg is used to prevent the runtime from exiting early
		go func(innerData *mapData, toUpdate *[]mapData, ext string, paths []string) { // run this extension in a goroutine
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
				}(ext, filePath)
			}

		}(&thisData, &toUpdate, ext, paths)
		return true
	})

	wg.Wait() // wait for all files to finish processing

	for _, innerData := range toUpdate {
		data.Store(innerData.Ext, innerData)
	}

	close(resultsChan) // Signal the writer goroutine to finish
	writerWG.Wait()    // Wait for the writer to flush and close the file

	if len(errs) > 0 {
		terminate(os.Stderr, "Error writing to output file: %v\n", errs)
	}

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
