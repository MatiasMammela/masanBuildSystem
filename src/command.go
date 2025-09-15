package src

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)


func configure(args []string) error{
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if _,err := os.Stat(cwd + "/build");!os.IsNotExist(err){
		return  fmt.Errorf("build directory already exists on /build!");
	}


	var buildDirpath = cwd + "/build";
	err = os.Mkdir(buildDirpath,0755);
	if err != nil {
		return err;
	}

	
	buildFilePath := cwd + "/build.lua"
	file, err := os.Create(buildFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Println("Project configured at:", cwd)

	return nil
}


func build(args []string) error {

    if len(args) == 0 {
        return fmt.Errorf("location of the build.lua file needed")
    } 
	
	
	argumentDir := args[0]
    info, err := os.Stat(argumentDir)
    if os.IsNotExist(err) {
        return fmt.Errorf("path does not exist: %s", argumentDir)
    }
    if err != nil {
        return err
    }


	buildFilePath := args[0]
	if info.IsDir() {
		buildFilePath = filepath.Join(argumentDir, "build.lua")
		if _, err := os.Stat(buildFilePath); os.IsNotExist(err) {
			return fmt.Errorf("no build.lua found in directory: %s", argumentDir)
		}
	}

	buildFilePath, err = filepath.Abs(buildFilePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}	

	L = lua.NewState()
    defer L.Close()
	luaFileDir := filepath.Dir(buildFilePath)
	
	L.SetGlobal("__lua_file_dir", lua.LString(luaFileDir))
	L.SetGlobal("__lua_file_path", lua.LString(buildFilePath))

	if err := os.Chdir(luaFileDir); err != nil {
    	return fmt.Errorf("failed to chdir: %v", err)
	}


	L.PreloadModule("mbs", mbs_loader)
  	
	
	chunk, err := L.LoadFile(buildFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", buildFilePath, err)
	}

	L.Push(chunk)
	RegisterStructType[*Project](L, "Project")
	RegisterStructType[*File](L,"File")
	RegisterStructType[*Directory](L,"Directory")
	RegisterStructType[*Package](L,"Package")

	errHandler := L.NewFunction(func(L *lua.LState) int {
		msg := L.ToString(1)
		L.Push(lua.LString(msg)) 
		return 1
	})

	if err := L.PCall(0, lua.MultRet, errHandler); err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

func Init_command(args []string) error{

	command := args[0]
	commandArgs := args[1:]

	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.StringVar(&GlobalFlags.builddir, "builddir", "", "build directory path for the build")

	if err := fs.Parse(commandArgs); err != nil {
		return err
	}

	switch command {
	case "configure":
		return configure(fs.Args())
	case "build":
		return build(fs.Args())
	case "version":
		return fmt.Errorf("%.1f" ,Version)
	default:
		return fmt.Errorf("unkown command: %s", command)
	}
}