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

type PackageRequest struct {
    Name     string
    Operator string
    Version  string
}

type Package struct {
	Name string
	Headers string
	Libraries string
	Found bool
	Static bool
	Version string
}

type Project struct {
	Name string
	Cwd string
	Build_dir_path string
	Build_file_path string
	Build_file_dir_path string
	Bin_dir_path string
	Sources []*File
	Headers []*Directory
	Libraries []*Package
	PublicLibraries []*Package
	CCompiler string
	CXXCompiler string
	Linker string
	CFlags []string 
	CXXFlags []string 
	LFlags []string
	ASMFlags []string
	LinkerFlags []string
	Assembler string
	AutoConfigure bool
	OS string
	Target_type string
	Linking string
	ObjFiles []string
	CStandard string
	CXXStandard string
	HasCpp bool
	HasC bool
}

type Flags struct {
	builddir string
	installdir string
	generate_compdb bool
}


var GlobalFlags = &Flags{}


var Projects []*Project
var L *lua.LState
const Version = 1.4

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
