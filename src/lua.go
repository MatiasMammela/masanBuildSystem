package src

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

	lua "github.com/yuin/gopher-lua"
)
func RegisterStructType[T any](L *lua.LState, name string) {
    mt := L.NewTypeMetatable(name)

    L.SetField(mt, "__index", L.NewFunction(func(L *lua.LState) int {
        ud := L.CheckUserData(1)
        val := reflect.ValueOf(ud.Value)
        if val.Kind() == reflect.Ptr {
            val = val.Elem()
        }

        key := L.CheckString(2)
        field := val.FieldByName(key)
        if !field.IsValid() {
            L.Push(lua.LNil)
            return 1
        }

        switch field.Kind() {
        case reflect.Bool:
            L.Push(lua.LBool(field.Bool()))
        case reflect.String:
            L.Push(lua.LString(field.String()))
        case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
            L.Push(lua.LNumber(field.Int()))
        case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
            L.Push(lua.LNumber(int64(field.Uint())))
        case reflect.Float32, reflect.Float64:
            L.Push(lua.LNumber(field.Float()))
        default:
            L.Push(lua.LString(fmt.Sprintf("%v", field.Interface())))
        }
        return 1
    }))

    L.SetField(mt, "__newindex", L.NewFunction(func(L *lua.LState) int {
        ud := L.CheckUserData(1)
        val := reflect.ValueOf(ud.Value)
        if val.Kind() == reflect.Ptr {
            val = val.Elem()
        }

        key := L.CheckString(2)
        field := val.FieldByName(key)
        if !field.IsValid() || !field.CanSet() {
            return 0
        }

        newVal := L.CheckAny(3)

        switch field.Kind() {
        case reflect.Bool:
            if b, ok := newVal.(lua.LBool); ok {
                field.SetBool(bool(b))
            }
        case reflect.String:
            if s, ok := newVal.(lua.LString); ok {
                field.SetString(string(s))
            }
        case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
            if n, ok := newVal.(lua.LNumber); ok {
                field.SetInt(int64(n))
            }
        case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
            if n, ok := newVal.(lua.LNumber); ok {
                field.SetUint(uint64(n))
            }
        case reflect.Float32, reflect.Float64:
            if n, ok := newVal.(lua.LNumber); ok {
                field.SetFloat(float64(n))
            }
        }
        return 0
    }))
}



func lua_glob_dirs(L *lua.LState) int {
	var patterns []string
	luaFileDir := L.GetGlobal("__lua_file_dir").String()
	top := L.GetTop()
	for i := 1; i <= top; i++ {
		val := L.Get(i)
		if str, ok := val.(lua.LString); ok {
			pattern := filepath.Join(luaFileDir, string(str))
			patterns = append(patterns, pattern)
		} else {
			L.ArgError(i, "expected string")
		}
	}

	dirs := find_directories(patterns)

	if len(dirs) == 0 {
		L.RaiseError("No matching directories found.")
		L.Push(lua.LNil)
		return 1
	}
    
	resultTbl := L.NewTable()
	for _, d := range dirs {
		ud := L.NewUserData()
		ud.Value = d
		L.SetMetatable(ud, L.GetTypeMetatable("Directory"))
		resultTbl.RawSetString(d.Name, ud)
	}
	L.Push(resultTbl)

	return 1
}


func lua_glob_files(L *lua.LState) int {
	var patterns []string
	luaFileDir := L.GetGlobal("__lua_file_dir").String()
	top := L.GetTop() 
	for i := 1; i <= top; i++ {
		val := L.Get(i)
		if str, ok := val.(lua.LString); ok {
			pattern := filepath.Join(luaFileDir, string(str))
			patterns = append(patterns, pattern)
		} else {
			L.ArgError(i, "expected string")
		}
	}
	files := find_files(patterns);

	for _, f := range files {
		if !f.Found {
			L.RaiseError("file not found: %s", f.Name)
			return 0
		}
	}

	resultTbl := L.NewTable()
	for _, f := range files {
		ud := L.NewUserData()
		ud.Value = f
		L.SetMetatable(ud, L.GetTypeMetatable("File"))
		resultTbl.RawSetString(f.Name, ud)
	}
	L.Push(resultTbl)
	return 1;
}

func lua_version(L *lua.LState) int {
    required := float64(L.CheckNumber(1)) 

    if required*10 != float64(int(required*10)) {
        msg("ERROR", fmt.Sprintf("Invalid version number %.6g (only one decimal place allowed)", required))
        return 0
    }

    if float64(required) > Version {
        msg("ERROR",fmt.Sprintf("Error build.lua requires MBS version %.1f or newer, but you are using %.1f",required,Version))
    }
    return 0
}


func lua_project(L *lua.LState) int {

    if L.GetTop() < 1 {
        L.RaiseError("project(name[, build_dir]) requires at least 1 argument: name (string)")
    }

    nameArg := L.Get(1)
    name, ok := nameArg.(lua.LString)
    if !ok {
        L.RaiseError("project(name[, build_dir]) first argument must be a string, got %s", nameArg.Type().String())
    }


	cwd, err := os.Getwd()
    if err != nil {
        L.RaiseError("failed to get current working directory: %v", err)
    }


    // Optional build directory
	baseDir := L.GetGlobal("__lua_file_dir").String()
    buildDir := filepath.Join(baseDir,"build")
    if L.GetTop() >= 2 {
        dirArg := L.Get(2)
        if dirStr, ok := dirArg.(lua.LString); ok {
            if filepath.IsAbs(string(dirStr)) {
                buildDir = string(dirStr)
            } else {
                buildDir = filepath.Join(baseDir, string(dirStr))
            }
        } else {
            L.RaiseError("project(name[, build_dir]) second argument must be a string, got %s", dirArg.Type().String())
        }
    }


    //Overwrite if we have flags
    if GlobalFlags.builddir != "" {
        if filepath.IsAbs(GlobalFlags.builddir) {
            buildDir = GlobalFlags.builddir
        } else {
            buildDir = filepath.Join(baseDir, GlobalFlags.builddir)
        }
    }


	abs, err := filepath.Abs(buildDir)
    if err != nil {
        L.ArgError(2, "could not resolve absolute build directory path")
        return 0
    }
	

	info, err := os.Stat(buildDir)
	if err != nil {
		L.RaiseError("build directory does not exist or cannot be accessed: %v", err)
	}
	if !info.IsDir() {
		L.RaiseError("build directory exists but is not a directory: %s", buildDir)
	}

	
	luaFile := L.GetGlobal("__lua_file_path").String()
    
	absfile, err := filepath.Abs(luaFile)
    if err != nil {
        L.ArgError(2, "could not resolve absolute build directory path")
        return 0
    }
    OS := runtime.GOOS;
    if OS == ""{
        return 0
    }
	project := &Project{
        Name:          string(name),
        Build_dir_path: abs,
		Build_file_path: absfile,
        Buildr_file_dir_path: baseDir,
		Cwd:            cwd,
        AutoConfigure: true,
        OS: OS,
    }

    Projects = append(Projects, project)

	ud := L.NewUserData()
	ud.Value = project
	L.SetMetatable(ud, L.GetTypeMetatable("Project"))
	L.Push(ud)

    return 1
}
func lua_sources(L *lua.LState) int{
	if L.GetTop() < 2 {
        L.ArgError(1, "expected atleast 2 arguments: sources and project")
        return 0
    }
	ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)
    if !ok {
        L.ArgError(2, "expected Project userdata")
        return 0
    }

	for i := 2; i <= L.GetTop(); i++ {
		filesTable := L.CheckTable(i)
		files := []*File{}

		filesTable.ForEach(func(_ lua.LValue, value lua.LValue) {
			if fUd, ok := value.(*lua.LUserData); ok {
				if f, ok := fUd.Value.(*File); ok && f.Found {
					files = append(files, f)
				}
			}
		})

		project.Sources = append(project.Sources, files...)
	}
	list_bound_sources(project.Sources,project)
	return 0;
}

func lua_debug(L *lua.LState) int {
    ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)
    if !ok {
        L.ArgError(2, "expected Project userdata")
        return 0
    }
    project.Debug();
    return 1;
}

func lua_compiler(L *lua.LState)int {
	if L.GetTop() < 2 {
        L.ArgError(1, "expected atleast 2 arguments: headers and project")
        return 0
    }
	ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)
    if !ok {
        L.ArgError(2, "expected Project userdata")
        return 0
    }

	compiler := L.CheckString(2);
	if compiler == "" {
  		L.ArgError(3, "expected Compiler")
	}

	project.Compiler = compiler;
	return 0;
}

func lua_set_cflags(L *lua.LState) int {
    if L.GetTop() < 2 {
        L.ArgError(1, "expected at least 2 arguments: project and flags")
        return 0
    }

    ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)
    if !ok {
        L.ArgError(1, "expected Project userdata")
        return 0
    }

    for i := 2; i <= L.GetTop(); i++ {
        if str, ok := L.Get(i).(lua.LString); ok {
            project.CFlags = append(project.CFlags, string(str))
        } else {
            L.ArgError(i, "expected string flag")
        }
    }
    return 0
}

func lua_set_asmflags(L *lua.LState) int {
    if L.GetTop() < 2 {
        L.ArgError(1, "expected at least 2 arguments: project and flags")
        return 0
    }

    ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)
    if !ok {
        L.ArgError(1, "expected Project userdata")
        return 0
    }

    for i := 2; i <= L.GetTop(); i++ {
        if str, ok := L.Get(i).(lua.LString); ok {
            project.ASMFlags = append(project.ASMFlags, string(str))
        } else {
            L.ArgError(i, "expected string flag")
        }
    }
    return 0
}

func lua_set_linkerflags(L *lua.LState) int {
    if L.GetTop() < 2 {
        L.ArgError(1, "expected at least 2 arguments: project and flags")
        return 0
    }

    ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)
    if !ok {
        L.ArgError(1, "expected Project userdata")
        return 0
    }

    for i := 2; i <= L.GetTop(); i++ {
        if str, ok := L.Get(i).(lua.LString); ok {
            project.LinkerFlags = append(project.LinkerFlags, string(str))
        } else {
            L.ArgError(i, "expected string flag")
        }
    }
    return 0
}

func lua_set_lflags(L *lua.LState) int {
    if L.GetTop() < 2 {
        L.ArgError(1, "expected at least 2 arguments: project and flags")
        return 0
    }

    ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)
    if !ok {
        L.ArgError(1, "expected Project userdata")
        return 0
    }

    for i := 2; i <= L.GetTop(); i++ {
        if str, ok := L.Get(i).(lua.LString); ok {
            project.LFlags = append(project.LFlags, string(str))
        } else {
            L.ArgError(i, "expected string flag")
        }
    }
    return 0
}


func lua_assembler(L *lua.LState)int {
	if L.GetTop() < 2 {
        L.ArgError(1, "expected atleast 2 arguments: headers and project")
        return 0
    }
	ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)
    if !ok {
        L.ArgError(2, "expected Project userdata")
        return 0
    }

	assembler := L.CheckString(2);
	if assembler == "" {
  		L.ArgError(3, "expected Assembler")
	}

	project.Assembler = assembler;
	return 0;
}

func lua_headers(L *lua.LState) int{
	if L.GetTop() < 2 {
        L.ArgError(1, "expected atleast 2 arguments: headers and project")
        return 0
    }
	ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)
    if !ok {
        L.ArgError(2, "expected Project userdata")
        return 0
    }

	for i := 2; i <= L.GetTop(); i++ {
		headerTable := L.CheckTable(i)
		headers := []*Directory{}

		headerTable.ForEach(func(_ lua.LValue, value lua.LValue) {
			if fUd, ok := value.(*lua.LUserData); ok {
				if f, ok := fUd.Value.(*Directory); ok && f.Found {
					headers = append(headers, f)
				}
			}
		})

		project.Headers = append(project.Headers, headers...)
	}
	list_bound_headers(project.Headers,project)
	return 0;
}
func lua_packages(L *lua.LState) int{
	if L.GetTop() < 2 {
        L.ArgError(1, "expected atleast 2 arguments: packages and project")
        return 0
    }
	ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)
    if !ok {
        L.ArgError(2, "expected Project userdata")
        return 0
    }

	for i := 2; i <= L.GetTop(); i++ {
		packagesTable := L.CheckTable(i)
		packages := []*Package{}

		packagesTable.ForEach(func(_ lua.LValue, value lua.LValue) {
			if fUd, ok := value.(*lua.LUserData); ok {
				if f, ok := fUd.Value.(*Package); ok && f.Found {
					packages = append(packages, f)
				}
			}
		})

		project.Libraries = append(project.Libraries, packages...)
	}

	list_bound_packages(project.Libraries,project)

	return 0;
}

func lua_glob_packages(L *lua.LState) int {
	var names []string

	top := L.GetTop()
	for i := 1; i <= top; i++ {
		val := L.Get(i)
		if str, ok := val.(lua.LString); ok {
			names = append(names, string(str))
		} else {
			L.ArgError(i, "expected string")
		}
	}

	pkgs := find_packages(names)

	resultTbl := L.NewTable()
	for _, p := range pkgs {
		ud := L.NewUserData()
		ud.Value = p
		L.SetMetatable(ud, L.GetTypeMetatable("Package"))
		resultTbl.RawSetString(p.Name, ud)
	}
	L.Push(resultTbl)

	return 1
}

func lua_build(L *lua.LState)int{
    ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)


    if !ok {
        L.ArgError(1, "expected Project userdata")
        return 0
    }
    fmt.Println("Building ",project.Name, "..")
    build_path := project.Build_dir_path

    info, err := os.Stat(build_path)
    if os.IsNotExist(err) || !info.IsDir() {
        fmt.Println("Build directory does not exist or is not a directory:", build_path)
        return 0
    }

    if project.AutoConfigure {
        auto_configure_project(project)
    }
    Generate_ninja(project)
	fmt.Println("Building finished!");
    return 1
}


func lua_copy(L *lua.LState) int {
    
    if L.GetTop() < 2 {
        L.ArgError(1, "expected atleast 2 arguments: files / directories (string or table) and project")
        return 0
    }

	dest := L.CheckString(L.GetTop())

  	baseDir := L.GetGlobal("__lua_file_dir").String()
    if !filepath.IsAbs(dest) {
        dest = filepath.Join(baseDir, dest)
    }

	var sources[]string;
  	for i := 1; i < L.GetTop(); i++ {
        val := L.Get(i)
        switch val.Type() {
        case lua.LTString:
			src := val.String()
            if !filepath.IsAbs(src) {
                src = filepath.Join(baseDir, src)
            }
            sources = append(sources, src)
        default:
            L.ArgError(i, "expected string or table")
        }
    }

	for _, src := range sources {
		info, err := os.Stat(src)
        if err != nil {
            L.RaiseError("failed to stat %s: %v", src, err)
            return 0
        }
		if info.IsDir() {
            err = copy_directory(src, dest)
        } else {
            err = copy_file(src, dest)
        }
		
		if err != nil {
			L.RaiseError("failed to copy %s -> %s: %v", src, dest, err)
			return 0
		}
	}
    return 0
}

func lua_set_autoconfigure(L *lua.LState)int {
    ud := L.CheckUserData(1)
    project, ok := ud.Value.(*Project)
    if !ok {
        L.ArgError(1, "expected Project userdata")
        return 0
    }

    autoConfigure := L.CheckBool(2)
    project.AutoConfigure = autoConfigure;
    return 0;
}

func mbs_loader(L *lua.LState)int{
    mod := L.SetFuncs(L.NewTable(),map[string]lua.LGFunction {
        "project":lua_project,
		"glob_files":lua_glob_files,
		"glob_dirs":lua_glob_dirs,
		"glob_packages":lua_glob_packages,
		"sources":lua_sources,
		"headers":lua_headers,
		"debug":lua_debug,
		"packages":lua_packages,
		"build":lua_build,
		"compiler":lua_compiler,
		"assembler":lua_assembler,
		"cflags":lua_set_cflags,
		"lflags":lua_set_lflags,
		"asmflags":lua_set_asmflags,
        "linkerflags":lua_set_linkerflags,
		"copy":lua_copy,
		"version": lua_version,
        "autoconfigure":lua_set_autoconfigure,
    })
    L.Push(mod);
    return 1;
}

func Init(){

	fmt.Println("Moi!")
}
