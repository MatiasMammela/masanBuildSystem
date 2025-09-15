package src

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

type File struct {
	Name string
	Type string
	Cwd string
	Found bool
}

type Directory struct {
	Name string
	Path string
	Found bool
}

type Package struct {
	Name string
	Headers string
	Libraries string
	Found bool
}

type Project struct {
	Name string
	Cwd string
	Build_dir_path string
	Build_file_path string
	Buildr_file_dir_path string
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
}

type Flags struct {
	builddir string
}


var GlobalFlags = &Flags{
	builddir: "",
}


var Projects []*Project
var L *lua.LState
const Version = 1.1

const  (  
    Red = "\033[31m"
    Green = "\033[32m"
    Yellow = "\033[33m"
    Reset = "\033[0m"
)
func msg(msgType string, a ...interface{}) {
    color := Reset
    switch msgType {
    case "WARNING":
        color = Yellow
    case "ERROR":
        color = Red
    case "OK":
        color = Green
        fmt.Print(a...)
        fmt.Print(color)
        fmt.Println(" [✔] ");
        fmt.Print(Reset)
        return
    }
    fmt.Print(color)
    fmt.Println(a...)
    fmt.Print(Reset)
}
