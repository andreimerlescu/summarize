# Summarize 

working on a project. The `summarize` command give you a powerful interface that is managed by arguments and environment
variables that define include/exclude extensions, and avoid substrings list while parsing paths. The binary has 
concurrency built into it and has limits for the output file. It ignores its default output directory so it won't
recursively build summaries upon itself. It defaults to writing to a new directory that it'll try to create in the
current working directory called `summaries`, that I recommend that you add to your `.gitignore` and `.dockerignore`.

![Diagram](/assets/diagram.png)

The **Summarize** package was designed for developers who wish to leverage the use of Artificial Intelligence while
I've found it useful to leverage the `make summary` command in all of my projects. This way, if I need to ask an AI a
question about a piece of code, I can capture the source code of the entire directory quickly and then just `cat` the
output file path provided and _voila_! The `-print` argument allows you to display the summary contents in the STDOUT
instead of the `Summary generated: summaries/summary.2025.07.29.08.59.03.UTC.md` that it would normally generate.

The **Environment** can be used to control the native behavior of the `summarize` binary, such that you won't be required
to type the arguments out each time. If you use _JSON_ all the time, you can enable its output format on every command
by using the `SUMMARIZE_ALWAYS_JSON`. If you always want to write the summary, you can use the `SUMMARIZE_ALWAYS_WRITE`
variable. If you want to always print the summary to STDOUT instead of the success message, you can use the variable
`SUMMARIZE_ALWAYS_PRINT`. If you want to compress the rendered summary every time, you can use the variable
`SUMMARIZE_ALWAYS_COMPRESS`. These `SUMMARIZE_ALWAYS_*` environment variables are responsible for customizing the 
runtime of the `summarize` application. 

When the `summarize` binary runs, it'll do its best to ignore files that it can't render to a text file. This includes
images, videos, binary files, and text files that are commonly linked to secrets.

The developer experience while using `summarize` is designed to enable quick use with just running `summarize` from 
where ever you wish to summarize. The `-d` for **source directory** defaults to `.` and the `-o`/`-f` for **output path**
defaults to a new timestamped file (`-f`) in the (`-o`) `summaries/` directory from the `.` context. The `-i` and `-x` are used to
define what to <b>i</b>nclude and e<b>x</b>clude various file extensions like `go,ts,py` etc.. The `-s` is used to 
**skip** over substrings within a scanned path. Dotfiles can completely be ignored by all paths by using `-ndf` as a flag.

Performance of the application can be tuned using the `-mf=<int>` to assign **Max Files** that will concurrently be
processed. The default is 369. The `-max=<int64>` represents a limit on how large the rendered summary can become.

Once the program finishes running, the rendered file will look similar to: 

```md
# Project Summary

<AI prompt description>

### `filename.go`

<File Info>

<full source code>

### `filename.cs`

<File Info>

<full source code>

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

## Options

| Name             | Argument | Type     | Usage                                                             |
|------------------|----------|----------|-------------------------------------------------------------------|
| `kSourceDir`     | `-d`     | `string` | Source directory path.                                            |
| `kOutputDir`     | `-o`     | `string` | Summary destination output directory path.                        |
| `kExcludeExt`    | `-x`     | `list`   | Comma separated string list of extensions to exclude.             |
| `kSkipContains`  | `-s`     | `list`   | Comma separated string to filename substrings to skip.            |
| `kIncludeExt`    | `-i`     | `list`   | Comma separated string of extensions to include.                  |
| `kFilename`      | `-f`     | `string` | Summary filename (writes to `-o` dir).                            | 
| `kVersion`       | `-v`     | `bool`   | When `true`, the binary version is shown                          |
| `kCompress`      | `-gz`    | `bool`   | When `true`, **gzip** is used on the contents of the summary      |
| `kMaxOutputSize` | `-max`   | `int64`  | Maximum size of the generated summary allowed                     |
| `kPrint`         | `-print` | `bool`   | Uses STDOUT to write contents of summary                          |
| `kWrite`         | `-write` | `bool`   | Uses the filesystem to save contents of summary                   |
| `kDebug`         | `-debug` | `bool`   | When `true`, extra content is written to STDOUT aside from report | 


## Environment

| Environment Variable        | Type     | Default Value          | Usage                                                                                                       | 
|-----------------------------|----------|------------------------|-------------------------------------------------------------------------------------------------------------|
| `SUMMARIZE_CONFIG_FILE`     | `String` | `./config.yaml`        | Contents of the YAML Configuration to use for [figtree](https://github.com/andreimerlescu/figtree).         |
| `SUMMARIZE_IGNORE_CONTAINS` | `List`   | \* see below           | Add items to this default list by creating your own new list here, they get concatenated.                   |
| `SUMMARIZE_INCLUDE_EXT`     | `List`   | \*\* see below \*      | Add extensions to include in the summary in this environment variable, comma separated.                     |
| `SUMMARIZE_EXCLUDE_EXT`     | `List`   | \*\*\* see below \* \* | Add exclusionary extensions to ignore to this environment variable, comma separated.                        |
| `SUMMARIZE_ALWAYS_PRINT`    | `Bool`   | `false`                | When `true`, the `-print` will write the summary to STDOUT.                                                 |
| `SUMMARIZE_ALWAYS_WRITE`    | `Bool`   | `false`                | When `true`, the `-write` will write to a new file on the disk.                                             | 
| `SUMMARIZE_ALWAYS_JSON`     | `Bool`   | `false`                | When `true`, the `-json` flag will render JSON output to the console.                                       |
| `SUMMARIZE_ALWAYS_COMPRESS` | `Bool`   | `false`                | When `true`, the `-gz` flag will use gzip to compress the summary contents and appends `.gz` to the output. |


### \* Default `SUMMARIZE_IGNORE_CONTAINS` Value

```json
7z,gz,xz,zst,zstd,bz,bz2,bzip2,zip,tar,rar,lz4,lzma,cab,arj,crt,cert,cer,key,pub,asc,pem,p12,pfx,jks,keystore,id_rsa,id_dsa,id_ed25519,id_ecdsa,gpg,pgp,exe,dll,so,dylib,bin,out,o,obj,a,lib,dSYM,class,pyc,pyo,__pycache__,jar,war,ear,apk,ipa,dex,odex,wasm,node,beam,elc,iso,img,dmg,vhd,vdi,vmdk,qcow2,db,sqlite,sqlite3,db3,mdb,accdb,sdf,ldb,log,trace,dump,crash,jpg,jpeg,png,gif,bmp,tiff,tif,webp,ico,svg,heic,heif,raw,cr2,nef,dng,mp3,wav,flac,aac,ogg,wma,m4a,opus,aiff,mp4,avi,mov,mkv,webm,flv,wmv,m4v,3gp,ogv,ttf,otf,woff,woff2,eot,fon,pfb,pfm,pdf,doc,docx,xls,xlsx,ppt,pptx,odt,ods,odp,rtf,suo,sln,user,ncb,pdb,ipch,ilk,tlog,idb,aps,res,iml,idea,vscode,project,classpath,factorypath,prefs,vcxproj,vcproj,filters,xcworkspace,xcuserstate,xcscheme,pbxproj,DS_Store,Thumbs.db,desktop.ini,lock,sum,resolved,tmp,temp,swp,swo,bak,backup,orig,rej,patch,~,old,new,part,incomplete,map,min.js,min.css,bundle.js,bundle.css,chunk.js,dat,data,cache,pid,sock,pack,idx,rev,pickle,pkl,npy,npz,mat,rdata,rds
```

```go

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
	
```

### \* \* Default `SUMMARIZE_INCLUDE_EXT`

```json
go,ts,tf,sh,py,js,Makefile,mod,Dockerfile,dockerignore,gitignore,esconfigs,md
```

```go
// defaultInclude are the -inc list of extensions that will be included in the summary
defaultInclude = []string{
    "go", "ts", "tf", "sh", "py", "js", "Makefile", "mod", "Dockerfile", "dockerignore", "gitignore", "esconfigs", "md",
}
```

### \* \* \* Default `SUMMARIZE_EXCLUDE_EXT`

```json
.min.js,.min.css,.git/,.svn/,.vscode/,.vs/,.idea/,logs/,secrets/,.venv/,/site-packages,.terraform/,summaries/,node_modules/,/tmp,tmp/,logs/
```

```go
// defaultAvoid are the -avoid list of substrings in file path names to avoid in the summary
defaultAvoid = []string{
    ".min.js", ".min.css", ".git/", ".svn/", ".vscode/", ".vs/", ".idea/", "logs/", "secrets/",
    ".venv/", "/site-packages", ".terraform/", "summaries/", "node_modules/", "/tmp", "tmp/", "logs/",
}
```

## Contribution 

Feel free to fork the project, submit a PR and contribute to open source projects. 
