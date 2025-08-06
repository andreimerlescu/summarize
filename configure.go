package main

import (
	"fmt"
	"path/filepath"

	"github.com/andreimerlescu/figtree/v2"
)

// init creates a new figtree with options to use CONFIG_FILE as a way of reading a YAML file while ignoring the env
func configure() {
	// figs is a tree of figs that ignore the ENV
	figs = figtree.With(figtree.Options{
		Harvest:           9,
		IgnoreEnvironment: true,
		ConfigFile:        envVal(eConfigFile, "./config.yaml"),
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
	figs = figs.NewBool(kPrint, envIs(eAlwaysPrint), "Print generated file contents to STDOUT")
	figs = figs.NewBool(kWrite, envIs(eAlwaysWrite), "Write generated contents to file")
	figs = figs.NewBool(kJson, envIs(eAlwaysJson), "Enable JSON formatting")
	figs = figs.NewBool(kCompress, envIs(eAlwaysCompress), "Use gzip compression in output")
	figs = figs.NewBool(kVersion, false, "Display current version of summarize")
	figs = figs.NewBool(kDebug, false, "Enable debug mode")
	figs = figs.NewBool(kShowExpanded, false, "Show expand menu")

	// ai mode
	figs = figs.NewBool(kAiEnabled, envIs(eDisableAi) == false, "Enable AI Features")
	figs = figs.NewString(kAiProvider, envVal(eAiProvider, dAiProvider), "AI Provider to use. (eg. ollama, openai, claude)")
	figs = figs.NewString(kAiModel, envVal(eAiModel, dAiModel), "AI Model to use for query")
	figs = figs.NewInt(kAiMaxTokens, envInt(eAiMaxTokens, dAiMaxTokens), "AI Max Tokens to use for query")
	figs = figs.NewInt(kAiSeed, envInt(eAiSeed, dAiSeed), "AI Seed to use for query")
	figs = figs.NewString(kAiApiKey, envVal(eAiApiKey, ""), "AI API Key to use for query (leave empty for ollama)")
	figs = figs.NewBool(kAiAlwaysAsk, envIs(eAiAlwaysAsk), "AI Always ask a question about the summary file you're summarizing and include the response in the output")
	figs = figs.NewBool(kAiAlwaysFollowUp, envIs(eAiAlwaysFollowUp), "Look until Ctrl+C by asking additional prompts for the chat conversation with the AI about the summary")

	// validators run internal figtree Assure<Mutagensis><Rule> funcs as arguments to validate against
	figs = figs.WithValidator(kSourceDir, figtree.AssureStringNotEmpty)
	figs = figs.WithValidator(kOutputDir, figtree.AssureStringNotEmpty)
	figs = figs.WithValidator(kFilename, figtree.AssureStringNotEmpty)
	figs = figs.WithValidator(kMaxFiles, figtree.AssureIntInRange(1, 17_369))
	figs = figs.WithValidator(kMaxOutputSize, figtree.AssureInt64InRange(369, 369_369_369_369))
	figs = figs.WithValidator(kAiSeed, figtree.AssureIntInRange(-1, 369_369_369_369))
	figs = figs.WithValidator(kAiMaxTokens, figtree.AssureIntInRange(-1, 369_369_369_369))

	// callbacks as figtree.CallbackAfterVerify run after the Validators above finish
	figs = figs.WithCallback(kSourceDir, figtree.CallbackAfterVerify, callbackVerifyReadableDirectory)
	figs = figs.WithCallback(kFilename, figtree.CallbackAfterVerify, callbackVerifyFile)
}
