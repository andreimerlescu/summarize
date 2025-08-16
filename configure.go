package main

import (
	"fmt"
	"path/filepath"

	"github.com/andreimerlescu/figtree/v2"
	"github.com/andreimerlescu/goenv/env"
)

// init creates a new figtree with options to use CONFIG_FILE as a way of reading a YAML file while ignoring the env
func configure() {
	// figs is a tree of figs that ignore the ENV
	figs = figtree.With(figtree.Options{
		Harvest:           9,
		IgnoreEnvironment: true,
		ConfigFile:        env.String(eConfigFile, "./config.yaml"),
	})

	// properties define new fig fruits on the figtree
	figs = figs.NewString(kSourceDir, ".", "Absolute path of directory you want to summarize.")
	figs = figs.NewString(kOutputDir, filepath.Join(".", "summaries"), fmt.Sprintf("Path of the directory to write the %s file to", newSummaryFilename()))
	figs = figs.NewString(kFilename, newSummaryFilename(), "Output file of summary.md")
	figs = figs.NewList(kIncludeExt, defaultInclude, "List of extensions to INCLUDE in summary.")
	figs = figs.NewList(kExcludeExt, defaultExclude, "List of extensions to EXCLUDE in summary.")
	figs = figs.NewList(kSkipContains, defaultAvoid, "List of path substrings if present to skip over full path.")
	figs = figs.NewInt(kMaxFiles, 369, "Maximum number of files to process concurrently")
	figs = figs.NewInt64(kMaxOutputSize, 1_776_369, "Maximum file size of output file")
	figs = figs.NewBool(kDotFiles, false, "Any path that is considered a dotfile can be included by setting this to true")
	figs = figs.NewBool(kPrint, env.Bool(eAlwaysPrint, false), "Print generated file contents to STDOUT")
	figs = figs.NewBool(kWrite, env.Bool(eAlwaysWrite, false), "Write generated contents to file")
	figs = figs.NewBool(kJson, env.Bool(eAlwaysJson, false), "Enable JSON formatting")
	figs = figs.NewBool(kCompress, env.Bool(eAlwaysCompress, false), "Use gzip compression in output")
	figs = figs.NewBool(kVersion, false, "Display current version of summarize")
	figs = figs.NewBool(kDebug, false, "Enable debug mode")
	figs = figs.NewBool(kShowExpanded, false, "Show expand menu")
	figs = figs.NewBool(kChat, false, "AI chat session with transcript based on new summary information in summary after")

	// ai mode
	figs = figs.NewBool(kAiEnabled, env.Bool(eDisableAi, false) == false, "Enable AI Features")
	figs = figs.NewString(kAiProvider, env.String(eAiProvider, dAiProvider), "AI Provider to use. (eg. ollama, openai, claude)")
	figs = figs.NewString(kAiModel, env.String(eAiModel, dAiModel), "AI Model to use for query")
	figs = figs.NewInt(kAiMaxTokens, env.Int(eAiMaxTokens, dAiMaxTokens), "AI Max Tokens to use for query")
	figs = figs.NewInt(kAiSeed, env.Int(eAiSeed, dAiSeed), "AI Seed to use for query")
	figs = figs.NewString(kAiApiKey, env.String(eAiApiKey, ""), "AI API Key to use for query (leave empty for ollama)")
	figs = figs.NewInt(kMemory, env.Int(eAiMemory, dMemory), "AI Memory to use for query")
	figs = figs.NewBool(kAiCachingEnabled, env.Bool(eAiAlwaysEnableCache, dCachingEnabled), "Enable LLM caching")
	figs = figs.NewUnitDuration(kAiTimeout, env.UnitDuration(eAiGlobalTimeout, dTimeoutUnit, dTimeout), dTimeoutUnit, "AI Timeout on each request allowed")

	// validators run internal figtree Assure<Mutagensis><Rule> funcs as arguments to validate against
	figs = figs.WithValidator(kSourceDir, figtree.AssureStringNotEmpty)
	figs = figs.WithValidator(kOutputDir, figtree.AssureStringNotEmpty)
	figs = figs.WithValidator(kFilename, figtree.AssureStringNotEmpty)
	figs = figs.WithValidator(kMaxFiles, figtree.AssureIntInRange(1, 369))
	figs = figs.WithValidator(kMemory, figtree.AssureIntInRange(1, 17_369_369))
	figs = figs.WithValidator(kMaxOutputSize, figtree.AssureInt64InRange(369, 369_369_369_369))
	figs = figs.WithValidator(kAiSeed, figtree.AssureIntInRange(-1, 369_369_369_369))
	figs = figs.WithValidator(kAiMaxTokens, figtree.AssureIntInRange(-1, 369_369_369_369))

	// callbacks as figtree.CallbackAfterVerify run after the Validators above finish
	figs = figs.WithCallback(kSourceDir, figtree.CallbackAfterVerify, callbackVerifyReadableDirectory)
	figs = figs.WithCallback(kFilename, figtree.CallbackAfterVerify, callbackVerifyFile)
}
