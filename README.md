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

## Environment

| Environment Variable        | Type     | Default Value          | Usage                                                                                               | 
|-----------------------------|----------|------------------------|-----------------------------------------------------------------------------------------------------|
| `SUMMARIZE_CONFIG_FILE`     | `String` | `./config.yaml`        | Contents of the YAML Configuration to use for [figtree](https://github.com/andreimerlescu/figtree). |
| `SUMMARIZE_IGNORE_CONTAINS` | `List`   | \* see below           | Add items to this default list by creating your own new list here, they get concatenated.           |
| `SUMMARIZE_INCLUDE_EXT`     | `List`   | \*\* see below \*      | Add extensions to include in the summary in this environment variable, comma separated.             |
| `SUMMARIZE_EXCLUDE_EXT`     | `List`   | \*\*\* see below \* \* | Add exclusionary extensions to ignore to this environment variable, comma separated.                |


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

## Options

| Name            | Argument | Type     | Usage                                                  |
|-----------------|----------|----------|--------------------------------------------------------|
| `kSourceDir`    | `-d`     | `string` | Source directory path.                                 |
| `kOutputDir`    | `-o`     | `string` | Summary destination output directory path.             |
| `kExcludeExt`   | `-x`     | `list`   | Comma separated string list of extensions to exclude.  |
| `kSkipContains` | `-s`     | `list`   | Comma separated string to filename substrings to skip. |
| `kIncludeExt`   | `-i`     | `list`   | Comma separated string of extensions to include.       |
| `kFilename`     | `-f`     | `string` | Summary filename (writes to `-o` dir).                 | 

## Contribution 

Feel free to fork the project, submit a PR and contribute to open source projects. 
