> Easy to use build system running on lua


---

## 📚 Table of contents

- [Overview](#🔍-overview)
- [Functions](#✨-functions)
    - [project](#project)
    - [version](#version)
    - [build](#build)
    - [debug](#debug)
    - [copy](#copy)
    - [glob_files](#glob_files)
    - [glob_dirs](#glob_dirs)
    - [glob_packages](#glob_packages)
    - [sources](#sources)
    - [headers](#headers)
    - [cflags](#cflags)
    - [lflags](#lflags)
    - [asmflags](#asmflags)
    - [linkerflags](#linkerflags)
    - [packages](#packages)
    - [compiler](#compiler)
    - [assembler](#assembler)
- [Command-line interface](#command-line-interface)
    - [build (CLI)](#build-cli)
    - [version (CLI)](#version-cli)
    - [configure (CLI)](#configure-cli)
- [Command-line flags](#⚑-command-line-flags)
    - [--builddir](#--builddir)
- [Examples](#✍️-examples)
    
## 🔍 overview


MBS (Masan build system) is an easy to use build system running on top of the lua-interpreter. The build system is basically a lua library written in go that can be included to any lua file. The power of the build system running on top of lua is the fact that it retains all the portability and power of the lua language and all the build system functions can be mixed with the lua syntax to create really flexible build files.

## ✨ functions

## project 

project(name string,build_dir_path string) *Project

Initializes a new project with the specified name
build_dir_path is optional and will default to build/

**Example**

```lua
mbs.project("myCProject")
```

## version

version(version float64) void

Can be used to enforce a minimum version from the users mbs build system.

**Example**

```lua
mbs.version(1.0)
```


**Example**

```lua
myproj = mbs.project("myapp", "output/")
```


## build

build(project *Project) void

Builds project.

**Example:**

```lua
mbs.build(project)
```

## debug

debug(project *Project) void

Prints project.

**Example:**
```
mbs.debug(project)
```

## copy

copy(src_path string... , dest_path string) void

Copy files or directories from src to dest.
src and dest can both be a file or a directory.

### Notice 
* Wildcards are not supported.
* ~/ eg extended paths are not supported.  


**Example:**
```lua
mbs.copy("src/","dest/")
```

## glob_files

glob_files(path string...) *File

Globs files with the given path.

**Example:**
```lua
myfiles = mbs.glob_files("src/*.c","/home/masa/somefolder/somefile.c")
```


## glob_dirs

glob_dirs(path string...) *Directory

Globs directories with the given path.

**Example:**
```lua
myfiles = mbs.glob_dirs("headers","includes")
```

## glob_packages

glob_packages(pkg_name string...) *Package

Globs packages with the given name using pkg-config utility.


**Example:**
```lua
mypackages = mbs.glob_packages("sdl2","ffreetype2")
```

## sources

sources(project *Project, sources *Files ...) void

Binds sources to project.

**Example:**
```lua
mbs.sources(project,myfiles)
```

## headers

headers(project *Project , headers *Directory ...) void

Binds headers to project.

**Example:**
```lua
mbs.headers(project,mydirs)
```

## cflags

cflags(project *Project,flag string...) void

Binds cflags to project.

**Example:**
```lua
mbs.cflags(project,"-myflag","-käpytikka")
```

## lflags

lflags(project *Project,flag string...) void

Binds lflags to project.

**Example:**
```lua
mbs.lflags(project,"-myflag","-käpytikka")
```

## asmflags

asmflags(project *Project,flag string...) void

Binds asm flags to project.

**Example:**
```lua
mbs.asmflags(project,"-myflag","-käpytikka")
```

## linkerflags

linkerflags(project *Project,flag string...) void

Binds linker flags to project.

**Example:**
```lua
mbs.linkerflags(project,"-myflag","-käpytikka")
```


## packages

packages(project *Project,package *Package ...) void

Binds packages to project.

**Example:**
```lua
mbs.packages(project,packages)
```

## compiler

compiler(project *Project,compiler *string)void

Binds compiler to project. 

**Example:**
```lua
mbs.compiler(project,"clang")
```

## assembler

assembler(project *Project,assembler *string)void

Binds assembler to project.

**Example:**
```lua
mbs.assembler(project,"nasm")
```


## </> Command-line interface

## build (CLI)

build <build_file_path>

Builds the project. Takes the build file path as a parameter

**Example**

```
mbs build ..
```

## version (CLI)

mbs version 

Prints the version of your mbs

**Example**

```
mbs version
```

## configure (CLI)

mbs confirue

Creates a build directory and a build file

**Example**

```
mbs configure
```

## ⚑  Command-line flags

## --builddir

Lets you bypass the build directory path set in the build.lua file

**Example**

```
mbs build --builddir myownbuilddir/ ..
```


## ✍️ Examples 

Examples of building C/C++ projects with mbs.

## Example 1

### Working directory tree

```
My C Project
├── build
├── build.lua
├── headers
│   └── header.h
├── resources
│   └── img.png
└── src
    └── main.c
```

### build.lua file contents

```lua

-- Add the mbs to your build.lua file
mbs = require("mbs")

-- Enforce the version. 
-- You can check your mbs version with mbs version
mbs.version(1.0)

-- Glob all the needed resources
local project = mbs.project("CProject")
local headers = mbs.glob_dirs("headers")
local sources = mbs.glob_files("src/*.c")
local packages = mbs.glob_packages("sdl2")

-- Copy the whole resources folder to your build directory
mbs.copy("resources","build")

-- Bind everything to your project
mbs.compiler(project,"gcc")
mbs.sources(project,sources)
mbs.headers(project,headers)
mbs.packages(project,packages)

-- List everything bound to the project
mbs.debug(project)

-- Build the project
mbs.build(project)
```