package src

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)



func ninja_pathcompat(path string) string {
    if len(path) >= 3 && ((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')) && path[1] == ':' && path[2] == '\\' {
        path = string(path[0]) + "$:" + path[2:]
    }
    return path
}

func Generate_headers(proj *Project, file *os.File) {
    for _, header := range proj.Headers {
        proj.CFlags = append(proj.CFlags, "-I"+header.Path)
    }
}

func Generate_packages(proj *Project, file *os.File) {
    for _, pkg := range proj.Libraries {
        if pkg.Found {
            if pkg.Headers != "" {
                for _, flag := range strings.Fields(pkg.Headers){
                    proj.CFlags = append_unique(proj.CFlags, flag)
                }
            }
            if pkg.Libraries != "" {
                var libs string
                if pkg.Static {
                    libs = "-Wl,-Bstatic " + pkg.Libraries + " -Wl,-Bdynamic"
					proj.LFlags = append_unique(proj.LFlags, libs)
                } else {
                   for _, flag := range strings.Fields(pkg.Libraries) {
						proj.LFlags = append_unique(proj.LFlags, flag)
					}
                }
            }
        }
    }
}

func Generate_sources(proj *Project, file *os.File) {
    objDir := filepath.Join(proj.Build_dir_path, "obj")
    os.MkdirAll(objDir, 0755)
    for _, src := range proj.Sources {
        if !src.Found {
            continue
        }
        baseName := filepath.Base(src.Cwd)
        objName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + "_" + strings.TrimPrefix(src.Type, ".") + ".o"
        objPath := filepath.Join(objDir, objName)
        proj.ObjFiles = append(proj.ObjFiles, objPath)
        depFile := objPath + ".d"
        switch src.Type {
        case ".c", ".cpp":
            fmt.Fprintf(file, "build %s: cc %s\n  CFLAGS = %s -MMD -MF %s\n",
                objPath, src.Cwd, strings.Join(proj.CFlags, " "), depFile)
            fmt.Fprintf(file, "  depfile = %s\n  deps = gcc\n", depFile)
        case ".asm":
            fmt.Fprintf(file, "build %s: asm %s\n  ASMFLAGS = %s\n",
                objPath, src.Cwd, strings.Join(proj.ASMFlags, " "))
        }
    }
}

func Generate_link(proj *Project, file *os.File) {
    binDir := filepath.Join(proj.Build_dir_path, "bin")
    os.MkdirAll(binDir, 0755)
    var output string
    switch proj.Target_type {
    case "static_lib":
        output = filepath.Join(binDir, "lib"+proj.Name+".a")
    case "dynamic_lib":
        output = filepath.Join(binDir, "lib"+proj.Name+".so")
    default:
        output = filepath.Join(binDir, proj.Name)
    }
    fmt.Fprintf(file, "build %s: link %s\n  LFLAGS = %s\n  LINKFLAGS = %s\n",
        output,
        strings.Join(proj.ObjFiles, " "),
        strings.Join(proj.LFlags, " "),
        strings.Join(proj.LinkerFlags, " "))
}

func windows_compatibility(project *Project) {

	project.Build_dir_path = ninja_pathcompat(project.Build_dir_path)
	project.Build_file_path = ninja_pathcompat(project.Build_file_path)
	project.Build_file_dir_path = ninja_pathcompat(project.Build_file_dir_path)

	for _, src := range project.Sources {
		src.Cwd = ninja_pathcompat(src.Cwd)
	}

	for _, hdr := range project.Headers {
		hdr.Path = ninja_pathcompat(hdr.Path)
	}

	for _, pkg := range project.Libraries {
		if pkg.Headers != "" {
			pkg.Headers = ninja_pathcompat(pkg.Headers)
		}
		if pkg.Libraries != "" {
			pkg.Libraries = ninja_pathcompat(pkg.Libraries)
		}
	}
}



func Generate_rules(proj *Project, file *os.File) {
	fmt.Fprintf(
		file,
		"rule cc\n  command = %s $CFLAGS -c $in -o $out\n  description = CC $in\n\n",
		proj.Compiler,
	)

	fmt.Fprintf(
		file,
		"rule asm\n  command = %s $ASMFLAGS $in -o $out\n  description = ASM $in\n\n",
		proj.Assembler,
	)

	switch proj.Target_type {
	case "static_lib":
		fmt.Fprintf(
			file,
			"rule link\n  command = ar rcs $out $in\n  description = AR $out\n\n",
		)

	case "dynamic_lib":
		fmt.Fprintf(
			file,
			"rule link\n  command = %s -shared $LINKFLAGS $in $LFLAGS -o $out\n  description = LINK $out\n\n",
			proj.Linker,
		)

	default:
		fmt.Fprintf(
			file,
			"rule link\n  command = %s $LINKFLAGS $in $LFLAGS -o $out\n  description = LINK $out\n\n",
			proj.Linker,
		)
	}
}


func write_header(file *os.File) {
	fmt.Fprintln(file, "# Auto-generated build.ninja")
	fmt.Fprintln(file)
}

func Generate_ninja(proj *Project) {
    ninjaPath := filepath.Join(proj.Build_dir_path, "build.ninja")
    file, err := os.Create(ninjaPath)
    if err != nil {
        fmt.Println("Failed to create build.ninja:", err)
        return
    }
    defer file.Close()

    if proj.OS == "windows" {
        windows_compatibility(proj)
    }

	write_header(file);
	Generate_rules(proj,file);
	
    Generate_headers(proj, file)
    Generate_packages(proj, file)
    Generate_sources(proj, file)
    Generate_link(proj, file)
}



