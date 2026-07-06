package src

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	_ "embed"
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


func run(args []string) error {
    // Run build first
    if err := build(args); err != nil {
        return err
    }

    if len(Projects) == 0 {
        return fmt.Errorf("no project found")
    }
    project := Projects[0]

    // Run ninja
    ninja := exec.Command("ninja", "-C", project.Build_dir_path)
    ninja.Stdout = os.Stdout
    ninja.Stderr = os.Stderr
    if err := ninja.Run(); err != nil {
        return fmt.Errorf("ninja failed: %v", err)
    }

    if GlobalFlags.generate_compdb {
        compdb := exec.Command("ninja", "-C", project.Build_dir_path, "-t", "compdb")
        outFile, err := os.Create(filepath.Join(project.Build_file_dir_path, "compile_commands.json"))
        if err != nil {
            return fmt.Errorf("failed to create compile_commands.json: %v", err)
        }
        defer outFile.Close()
        compdb.Stdout = outFile
        compdb.Stderr = os.Stderr
        if err := compdb.Run(); err != nil {
            return fmt.Errorf("failed to generate compile_commands.json: %v", err)
        }
        msg("OK", "Generated compile_commands.json")
    }

    binary := filepath.Join(project.Bin_dir_path, project.Name)
    cmd := exec.Command(binary)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    return cmd.Run()
}

func install(args []string) error {

    if err := build(args); err != nil {
        return err
    }

    if len(Projects) == 0 {
        return fmt.Errorf("no project found")
    }
    project := Projects[0]

    ninja := exec.Command("ninja", "-C", project.Build_dir_path)
    ninja.Stdout = os.Stdout
    ninja.Stderr = os.Stderr
    if err := ninja.Run(); err != nil {
        return fmt.Errorf("ninja failed: %v", err)
    }

    installPath := "/usr/local/bin"
    if GlobalFlags.installdir != "" {
        installPath = GlobalFlags.installdir
    }

    if err := os.MkdirAll(installPath, 0755); err != nil {
        return fmt.Errorf("failed to create install directory: %v", err)
    }

    binary := filepath.Join(project.Bin_dir_path, project.Name)
    dest := filepath.Join(installPath, project.Name)

    if err := copy_file(binary, dest); err != nil {
        return fmt.Errorf("failed to install binary: %v", err)
    }

    if err := os.Chmod(dest, 0755); err != nil {
        return fmt.Errorf("failed to set executable permission: %v", err)
    }

    msg("OK", fmt.Sprintf("Installed %s to %s", project.Name, dest))
    return nil
}

//go:embed banner.txt
var Banner string

func help() error {

	fmt.Println(Banner)
	fmt.Println("Usage: mbs <command> [options] [build_file_path]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  configure    Creates a build directory and a build file")
	fmt.Println("  build        Builds the project from the given build file path")
	fmt.Println("  run          Builds and runs the project executable")
	fmt.Println("  install      Installs the built project")
	fmt.Println("  version      Prints the current mbs version")
	fmt.Println("  help         Shows this help message")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --builddir <path>       Bypass the build directory path set in the build.lua file")
	fmt.Println("  --installdir <path>     Set the installation directory path (used with install)")
	fmt.Println("  --generate_compdb       Generate a compile_commands.json for the project")
	fmt.Println()
	fmt.Println("Documentation: https://github.com/MatiasMammela/masanBuildSystem/blob/main/document.md")

	return nil
}

func Init_command(args []string) error{

	if len(args) == 0 {
		return help()
	}

	command := args[0]
	commandArgs := args[1:]

	positional, err := ParseFlags(command, commandArgs)
	if err != nil {
		return err
	}

	switch command {
	case "configure":
		return configure(positional)
	case "build":
		return build(positional)
	case "version":
		return fmt.Errorf("%.1f", Version)
	case "run":
		return run(positional)
	case "install":
		return install(positional)
	case "help":
		return help()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}
