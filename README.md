# Summarize 

A go utility that will capture files with an extension pattern into a single markdown formatted
file that looks like: 

```md
# Project Summary

### `filename.ext`

<full source code>

### `filename.ext`

... etc.

```

## Installing

Built for Go **1.24.0**. Universal application.

```bash
go install github.com/andreimerlescu/summarize@latest
```

## Usage

```bash
cd ~/work/figtree
summarize 
ls -la summaries
```

You can also use a config file like this: 

```json
{"i": "go,sh,ts", "d": "/home/user/work/project", "o": "/home/user/summaries/project"}
```

Saved to something like `/home/user/work/config.yaml` can be called like: 

```bash
export SUMMARIZE_CONFIG_FILE=/home/user/work/summarize.config.yaml
cd ~/work/project
sumarize
unset SUMMARIZE_CONFIG_FILE
cd ~/work/anotherProject
summarize -d anotherProject -o /home/user/summaries/anotherProject
```

Since `figtree` is designed to be very functional, its lightweight but feature 
intense design through simple biology memetics makes it well suited for this program. 

## Options

| Name            | Argument | Type     | Usage                                                  |
|-----------------|----------|----------|--------------------------------------------------------|
| `kSourceDir`    | -d`      | `string` | Source directory path.                                 |
| `kOutputDir`    | -o`      | `string` | Summary destination output directory path.             |
| `kExcludeExt`   | `-x`     | `list`   | Comma separated string list of extensions to exclude.  |
| `kSkipContains` | `-s`     | `list`   | Comma separated string to filename substrings to skip. |
| `kIncludeExt`   | `-i`     | `list`   | Comma separated string of extensions to include.       |
| `kFilename`     | `-f`     | `string` | Summary filename (writes to `-o` dir).                 | 

## Contribution 

Feel free to fork the project, submit a PR and contribute to open source projects. 
