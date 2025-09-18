> Easy to use build system running on lua


---

## 📚 Table of contents

<summary><a href="#🔍-overview">Overview</a></summary>
<details>
<summary><a href="#✨-functions">Functions</a></summary>

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
- [autoconfigure](#autoconfigure)

</details>

<details>
<summary><a href="#command-line-interface">Command-line interface</a></summary>

- [build (CLI)](#build-cli)
- [version (CLI)](#version-cli)
- [configure (CLI)](#configure-cli)

</details>

<details>
<summary><a href="#⚑-command-line-flags">Command-line flags</a></summary>

- [--builddir](#--builddir)

</details>

<details>
<summary><a href="#🛠-api">Lua exposed API</a></summary>

- [objects overview](#objects)
- [Project obj](#project-obj)
- [File obj](#file-obj)
- [Directory obj](#directory-obj)
- [Package obj](#package-obj)

</details>

<details>
<summary><a href="#✍️-examples">Examples</a></summary>

- [Basics](#example-1--basics)
- [Multiple projects](#example-2--multiple-projects)
- [Using lua](#example-3--using-lua)
</details>



## 🔍 overview 
<details> 
<summary></summary>
MBS (Masan build system) is an easy to use build system running on top of the lua-interpreter. The build system is basically a lua library written in go that can be included to any lua file. The power of the build system running on top of lua is the fact that it retains all the portability and power of the lua language and all the build system functions can be mixed with the lua syntax to create really flexible build files.
</details>

## ✨ functions
<details>
<summary></summary>
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

If the package is not found from the users system the function tries to install them through a suitable package manager that the user might have.

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

## autoconfigure 

autoconfigure(project *Project , enabled bool) void

Sets autoconfigure on or off for the current project. 
If autoconfigure is enabled it's run with the build function. 
Autoconfigure tries to find suitable compilers , assemblers and flags for your project.

Autoconfigure is enabled by default by every mbs project.

**Example:**

```lua 
mbs.autoconfigure(project,false)
```
</details>

## </> Command-line interface
<details>
<summary></summary>

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
</details>

## ⚑  Command-line flags
<details>
<summary></summary>

## --builddir

Lets you bypass the build directory path set in the build.lua file

**Example**

```
mbs build --builddir myownbuilddir/ ..
```
</details>

## 🛠 API
<details>
<summary></summary>
The real power of this build system is the api that is exposed to lua.
It lets you read / write to the objects you create with the build system, break them down and play with them.

There is nothing you cant do with this system as you have a full-fledged programming language in your hands.

Examples on how this works in practise are found from the [Examples](#✍️-examples) section of this documentation.

## Objects

All build system objects are fully exposed to Lua.
While it’s possible to modify them directly, this is not recommended. Doing so can clutter your build files and make them harder to understand.

## Project obj

```go
type Project struct {
	Name string
	Cwd string
	Build_dir_path string
	Build_file_path string
	Sources []*File
	Headers []*Directory
	Libraries []*Package
	Compiler string 
	CFlags []string
	LFlags []string
	ASMFlags []string
	LinkerFlags []string
	Assembler string
	AutoConfigure bool
	OS string
}
```

## File obj

```go
type File struct {
	Name string
	Type string
	Cwd string
	Found bool
}
```

## Directory obj

```go
type Directory struct {
	Name string
	Path string
	Found bool
}
```

## Package obj

```go
type Package struct {
	Name string
	Headers string
	Libraries string
	Found bool
}
```
</details>

## ✍️ Examples 
<details>
<summary></summary>
Examples of building C/C++ projects with mbs.

## Example 1 / Basics

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

## Example 2 / Multiple projects

### Working directory tree

```
My C Project
.
├── build
│   ├── build.ninja
│   ├── cppProject
│   └── main_cpp.o
├── build.lua
├── build2
│   ├── build.ninja
│   ├── cppProject2
│   └── main_cpp.o
├── headers
│   └── test.h
├── resources
└── src
    └── main.cpp
```

### build.lua file contents

```lua

-- This example builds 2 projects from the same build.lua file

-- This can be achieved by simply giving the second project a different build directory

local mbs = require("mbs")
 -- We dont appoint any specific build directory for the first project so it will just use build/
local project = mbs.project("cppProject")
local sources = mbs.glob_files("src/*")
local headers = mbs.glob_dirs("headers")

mbs.sources(project,sources)
mbs.headers(project,headers)
mbs.build(project)

 -- For the second project we appoint a different build directory so it wont overwrite the first project on the build/ directory
local project2 = mbs.project("cppProject2","build2/")
local sources2 = mbs.glob_files("src/*")
local headers2 = mbs.glob_dirs("headers")

mbs.sources(project2,sources)
mbs.headers(project2,headers)
mbs.build(project2)

```

## Example 3 / Using lua

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

mbs = require("mbs")
mbs.version(1.0)

local project = mbs.project("CProject")

-- headers2 does not exist
local headers = mbs.glob_dirs("headers","headers2") 
local sources = mbs.glob_files("src/*.cpp","src/*.asm")

-- gtk10 does not exist
local packages = mbs.glob_packages("sdl2","gtk10")


-- Using the lua exposed data from *File , *Directory , *Package and *Project.
-- We can mix the exposed data with normal lua to create virtually anything

if not headers["headers2"].Found then
    print("⚠️ headers2 not found")
end

if not packages["gtk10"].Found then
    print("⚠️ gtk10 not found")
end

-- Access to this data works both ways. We can also modify the data.
packages["gtk10"].Found = true 
project.Name = "BetterName"

mbs.sources(project,sources)
mbs.headers(project,headers)
mbs.packages(project,packages)

mbs.build(project)
mbs.debug(project)
```
</details>


