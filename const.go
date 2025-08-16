package main

import "time"

const (
	projectName string = "github.com/andreimerlescu/summarize"
	tFormat     string = "2006.01.02.15.04.05.UTC"

	// eConfigFile ENV string of path to .yml|.yaml|.json|.ini file
	eConfigFile string = "SUMMARIZE_CONFIG_FILE"

	// eAddIgnoreInPathList ENV string (comma separated list) of substrings to ignore if path contains
	eAddIgnoreInPathList string = "SUMMARIZE_IGNORE_CONTAINS"

	// eAddIncludeExtList ENV string (comma separated list) of acceptable file extensions of any scanned path
	eAddIncludeExtList string = "SUMMARIZE_INCLUDE_EXT"

	// eAddExcludeExtList ENV string (comma separated list) of rejected file extensions of any scanned path
	eAddExcludeExtList string = "SUMMARIZE_EXCLUDE_EXT"

	// eAlwaysWrite ENV string-as-bool (as "TRUE" or "true" for true) always sets -write true in CLI argument parsing
	eAlwaysWrite string = "SUMMARIZE_ALWAYS_WRITE"

	// eAlwaysPrint ENV string-as-bool (as "TRUE" or "true" for true) always sets -print true in CLI argument parsing
	eAlwaysPrint string = "SUMMARIZE_ALWAYS_PRINT"

	// eAlwaysJson ENV string-as-bool (as "TRUE" or "true" for true) always sets -json true in CLI argument parsing
	eAlwaysJson string = "SUMMARIZE_ALWAYS_JSON"

	// eAlwaysCompress ENV string-as-bool (as "TRUE" or "true" for true) always sets -gz true in CLI argument parsing
	eAlwaysCompress string = "SUMMARIZE_ALWAYS_COMPRESS"

	eDisableAi           string = "SUMMARIZE_DISABLE_AI"
	eAiProvider          string = "SUMMARIZE_AI_PROVIDER"
	eAiModel             string = "SUMMARIZE_AI_MODEL"
	eAiApiKey            string = "SUMMARIZE_AI_API_KEY"
	eAiMaxTokens         string = "SUMMARIZE_AI_MAX_TOKENS"
	eAiSeed              string = "SUMMARIZE_AI_SEED"
	eAiMemory            string = "SUMMARIZE_AI_MEMORY"
	eAiAlwaysEnableCache string = "SUMMARIZE_AI_ENABLE_CACHE"
	eAiGlobalTimeout     string = "SUMMARIZE_AI_GLOBAL_TIMEOUT"

	dAiSeed      int    = -1
	dAiMaxTokens int    = 3000
	dAiProvider  string = "ollama"
	dAiModel     string = "qwen3:8b"
	// dAiModel          string = "mistral-small3.2:24b"
	dCachingEnabled bool          = true
	dMemory         int           = 36963
	dTimeout        time.Duration = 77
	dTimeoutUnit    time.Duration = time.Second

	kAiEnabled        string = "ai"
	kAiProvider       string = "provider"
	kAiModel          string = "model"
	kAiApiKey         string = "api-key"
	kAiMaxTokens      string = "max-tokens"
	kAiSeed           string = "seed"
	kMemory           string = "memory"
	kAiCachingEnabled string = "caching"
	kAiTimeout        string = "timeout"

	kShowExpanded string = "expand"

	kChat string = "chat"

	// kSourceDir figtree fig string -d for the directory path to generate a summary of
	kSourceDir string = "d"

	// kOutputDir figtree fig string -o for the output directory where the summary is saved
	kOutputDir string = "o"

	// kIncludeExt figtree fig list (string-as-list aka comma separated list) -i for the extensions to summarize
	kIncludeExt string = "i"

	// kExcludeExt figtree fig list (string-as-list aka comma separated list) -x for the extensions NOT to summarize
	kExcludeExt string = "x"

	// kSkipContains figtree fig list (string-as-list aka comma separated list) -s for the substrings in the paths to ignore
	kSkipContains string = "s"

	// kFilename figtree fig string -f is the name of the file to save inside kOutputDir
	kFilename string = "f"

	// kPrint figtree fig bool -print will render to STDOUT the contents of the summary
	kPrint string = "print"

	// kMaxOutputSize figtree fig int64 -max will stop summarizing the kSourceDir once kFilename reaches this size in bytes
	kMaxOutputSize string = "max"

	// kWrite figtree fig bool -write will write the summary to the kFilename in the kSourceDir
	kWrite string = "write"

	// kVersion figtree fig bool -v will display the current version of the binary
	kVersion string = "v"

	// kDotFiles figtree fig bool -ndf will skip over any directory that has a prefix of "."
	kDotFiles string = "ndf"

	// kMaxFiles figtree fig int64 -mf will specify the maximum number of files that will concurrently be summarized
	kMaxFiles string = "mf"

	// kDebug figtree fig bool -debug will render addition log statements to STDOUT
	kDebug string = "debug"

	// kJson figtree fig bool -json will render the output as JSON to the console's STDOUT only
	kJson string = "json"

	// kCompress figtree fig bool -gz will gzip compress the contents of kFilename that is written to kOutputDir
	kCompress string = "gz"
)
