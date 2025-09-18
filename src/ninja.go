package src

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)



type Ninja struct {
	file *os.File
	LFlags []string
	CFlags []string
	ASMFlags []string
	LinkerFlags []string
	ObjFiles []string

}


func ninja_pathcompat(path string) string {
    p := strings.ReplaceAll(path, "\\", "/")
    if len(p) >= 3 && ((p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z')) && p[1] == ':' && p[2] == '/' {
        p = string(p[0]) + "$" + p[2:]
    }
    return p
}

func Generate_packages(proj *Project, ninja *Ninja) {
	for _, pkg := range proj.Libraries {
		if pkg.Found {
			if pkg.Headers != "" {
				ninja.CFlags = append(ninja.CFlags, pkg.Headers)
			}
			if pkg.Libraries != "" {
				ninja.LFlags = append(ninja.LFlags, pkg.Libraries)
			}
		}
	}
}

func Generate_headers(Project *Project ,Ninja *Ninja){
	for _, header := range Project.Headers {
        Ninja.CFlags = append(Ninja.CFlags, "-I"+header.Path)
    }
}

func Generate_sources(proj *Project, ninja *Ninja) {
	for _, src := range proj.Sources {
		if !src.Found {
			continue
		}

		baseName := filepath.Base(src.Cwd)
		objName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + "_" + strings.TrimPrefix(src.Type, ".") + ".o"
		objPath := filepath.Join(proj.Build_dir_path, objName)
		ninja.ObjFiles = append(ninja.ObjFiles, objPath)

		depFile := objPath + ".d" 

		switch src.Type {
		case ".c", ".cpp":
			fmt.Fprintf(ninja.file,
				"build %s: cc %s\n  CFLAGS = %s -MMD -MF %s\n",
				objPath, src.Cwd, strings.Join(ninja.CFlags, " "), depFile)
			fmt.Fprintf(ninja.file, "  depfile = %s\n  deps = gcc\n", depFile)
		case ".asm":
			fmt.Fprintf(ninja.file,
				"build %s: asm %s\n  ASMFLAGS = %s\n",
				objPath, src.Cwd, strings.Join(ninja.ASMFlags, " "))
		}
	}
}


func Generate_link(proj *Project, ninja *Ninja) {
	output := filepath.Join(proj.Build_dir_path, proj.Name)
	fmt.Fprintf(ninja.file,
    "build %s: link %s\n  LFLAGS = %s\n  LINKFLAGS = %s\n",
    output,
    strings.Join(ninja.ObjFiles, " "),
    strings.Join(ninja.LFlags, " "),
    strings.Join(ninja.LinkerFlags, " "))
}

func windows_compatibility(project *Project) {

	project.Build_dir_path = ninja_pathcompat(project.Build_dir_path)
	project.Build_file_path = ninja_pathcompat(project.Build_file_path)
	project.Buildr_file_dir_path = ninja_pathcompat(project.Buildr_file_dir_path)

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

func Generate_ninja(Project *Project){
	ninjaPath := Project.Build_dir_path+"/build.ninja"
	file , err := os.Create(ninjaPath)
	if err != nil {
		fmt.Println("Failed to create build.ninja:",err)
	}
	defer file.Close()
          
	if Project.OS == "windows" {
		windows_compatibility(Project)
	}

	fmt.Fprintln(file, "# This is an auto-generated build.ninja file")
	fmt.Fprintln(file)
	fmt.Fprintf(file, "rule cc\n  command = %s $CFLAGS -c $in -o $out\n  description = CC $in\n\n", Project.Compiler)
	fmt.Fprintf(file, "rule asm\n  command = %s $ASMFLAGS $in -o $out\n  description = ASM $in\n\n", Project.Assembler)
	fmt.Fprintf(file, "rule link\n  command = %s $LINKFLAGS $in $LFLAGS -o $out\n  description = LINK $out\n\n", Project.Compiler)

	Ninja  := &Ninja{file:file,
		CFlags: Project.CFlags,
		LFlags: Project.LFlags,
		ASMFlags: Project.ASMFlags,
		LinkerFlags: Project.LinkerFlags,
	}
	Generate_headers(Project, Ninja)
	Generate_packages(Project, Ninja)
	Generate_sources(Project, Ninja)
	Generate_link(Project, Ninja)
}