package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	check "github.com/andreimerlescu/checkfs"
	"github.com/andreimerlescu/checkfs/directory"
	"github.com/andreimerlescu/sema"
)

func process() {
	preprocess()
	defer postprocess()
	// populate data with the kSourceDir files based on -inc -exc -avoid lists
	capture("walking source directory", filepath.Walk(sourceDir, summarize))

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
}

func preprocess() {
	configure()
	capture("figs loading environment", figs.Load())

	isDebug = *figs.Bool(kDebug)

	inc = *figs.List(kIncludeExt)
	if len(inc) == 1 && inc[0] == "useExpanded" {
		figs.StoreList(kIncludeExt, extendedDefaultInclude)
	}

	exc = *figs.List(kExcludeExt)
	if len(exc) == 1 && exc[0] == "useExpanded" {
		figs.StoreList(kExcludeExt, extendedDefaultExclude)
	}

	ski = *figs.List(kSkipContains)
	if len(ski) == 1 && ski[0] == "useExpanded" {
		figs.StoreList(kSkipContains, extendedDefaultAvoid)
	}

	if isDebug {
		fmt.Println("INCLUDE: ", strings.Join(inc, ", "))
		fmt.Println("EXCLUDE: ", strings.Join(exc, ", "))
		fmt.Println("SKIP: ", strings.Join(ski, ", "))
	}

	if *figs.Bool(kShowExpanded) {
		fmt.Println("Expanded:")
		fmt.Printf("-%s=%s\n", kIncludeExt, strings.Join(*figs.List(kIncludeExt), ","))
		fmt.Printf("-%s=%s\n", kExcludeExt, strings.Join(*figs.List(kExcludeExt), ","))
		fmt.Printf("-%s=%s\n", kSkipContains, strings.Join(*figs.List(kSkipContains), ","))
		os.Exit(0)
	}

	if *figs.Bool(kVersion) {
		fmt.Println(Version())
		os.Exit(0)
	}

	lIncludeExt = *figs.List(kIncludeExt)
	lExcludeExt = *figs.List(kExcludeExt)
	lSkipContains = *figs.List(kSkipContains)

	sourceDir = *figs.String(kSourceDir)
	outputDir = *figs.String(kOutputDir)
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

	wg = &sync.WaitGroup{}
	throttler = sema.New(runtime.GOMAXPROCS(0))
	data = &sync.Map{}

	for _, i := range lIncludeExt {
		data.Store(i, mapData{
			Ext:   i,
			Paths: []string{},
		})
	}

}

func postprocess() {
	maxFileSemaphore = sema.New(*figs.Int(kMaxFiles))
	resultsChan = make(chan Result, *figs.Int(kMaxFiles))
	writerWG = &sync.WaitGroup{}
	writerWG.Add(1)
	go receive()

	seen = &seenStrings{m: make(map[string]bool)}

	data.Range(iterate)

	wg.Wait() // wait for all files to finish processing

	for _, innerData := range toUpdate {
		data.Store(innerData.Ext, innerData)
	}

	close(resultsChan) // Signal the writer goroutine to finish
	writerWG.Wait()    // Wait for the writer to flush and close the file

	if len(errs) > 0 {
		terminate(os.Stderr, "Error writing to output file: %v\n", errs)
	}

	done()

}
