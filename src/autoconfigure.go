package src

import (
	"fmt"
	"os/exec"
)

var (
	c_auto_flags          = []string{"-Wall", "-Wextra", "-O2"}
	cpp_auto_flags        = []string{"-Wall", "-Wextra", "-O2"}
	asm_auto_flags        = []string{"-f", "elf64"}
	assembler_candidates  = []string{"nasm", "as"}
	c_compiler_candidates = []string{"gcc", "clang", "cc"}
	cpp_compiler_candidates = []string{"g++", "clang++", "c++"}
	cpp_standard = "c++17"
	c_standard = "c11"
)

func detect_languages(proj *Project) (hasC bool, hasCpp bool, hasAsm bool) {
	for _, src := range proj.Sources {
		switch src.Type {
		case ".cpp", ".cxx", ".cc":
			hasCpp = true
		case ".c":
			hasC = true
		case ".asm", ".s":
			hasAsm = true
		}
	}
	return
}

func default_flags(lang string) []string {
	switch lang {
	case "c":
		return c_auto_flags;
	case "cpp":
		return cpp_auto_flags;
	case "asm":
		return asm_auto_flags;
	default:
		return []string{}
	}
}

func default_standard(lang string) string {
	switch lang {
	case "c":
		return c_standard;
	case "cpp":
		return cpp_standard;
	default:
		return ""
	}
}


func append_unique(flags []string, new_flags ...string) []string {
    existing := make(map[string]bool)
    for _, f := range flags {
        existing[f] = true
    }
    for _, f := range new_flags {
        if !existing[f] {
            flags = append(flags, f)
            existing[f] = true
        }
    }
    return flags
}

func detect_assembler() string {
	for _, asm := range assembler_candidates {
		if _, err := exec.LookPath(asm); err == nil {
			return asm
		}
	}
	return "" 
}

func gnu_source_needed(proj *Project, compiler string) bool {
	if proj.OS != "linux" {
		return false
	}
	switch compiler {
	case "gcc", "clang", "cc", "g++", "clang++", "c++":
		return true
	default:
		return false
	}
}
func auto_configure_project(proj *Project) {
	hasC, hasCpp, hasAsm := detect_languages(proj)
	proj.HasC = hasC;
	proj.HasCpp = hasCpp;
	if !hasC && !hasCpp {
		fmt.Println("Warning: no C or C++ source files found to detect language")
		return
	}

	if hasC {
		if proj.CCompiler == "" {
			proj.CCompiler = detect_compiler("c")
		}
		proj.CFlags = append_unique(proj.CFlags, default_flags("c")...)
 
		if proj.CStandard == "" {
			proj.CStandard = default_standard("c")
		}
		proj.CFlags = append_unique(proj.CFlags, "-std="+proj.CStandard)
 
		if gnu_source_needed(proj, proj.CCompiler) {
			proj.CFlags = append_unique(proj.CFlags, "-D_GNU_SOURCE")
		}
	}
	
	if hasCpp {
		if proj.CXXCompiler == "" {
			proj.CXXCompiler = detect_compiler("cpp")
		}
		proj.CXXFlags = append_unique(proj.CXXFlags, default_flags("cpp")...)
 
		if proj.CXXStandard == "" {
			proj.CXXStandard = default_standard("cpp")
		}
		proj.CXXFlags = append_unique(proj.CXXFlags, "-std="+proj.CXXStandard)
 
		if gnu_source_needed(proj, proj.CXXCompiler) {
			proj.CXXFlags = append_unique(proj.CXXFlags, "-D_GNU_SOURCE")
		}
	}

	if hasAsm {
		proj.Assembler = detect_assembler()
		proj.ASMFlags = append(proj.ASMFlags,default_flags("asm")...);
		proj.LFlags = append(proj.LFlags, "-no-pie")
	}
    
	driver := proj.CCompiler
	if hasCpp {
		driver = proj.CXXCompiler
	}
 
	if proj.Linker != "" && proj.Linker != driver {
		proj.LinkerFlags = append([]string{"-fuse-ld=" + proj.Linker}, proj.LinkerFlags...)
		proj.Linker = driver
	} else {
		proj.Linker = driver
	}
}

func detect_compiler(lang string) string {
	if lang == "c" {
		for _, c := range c_compiler_candidates {
			if _, err := exec.LookPath(c); err == nil {
				return c
			}
		}
	} else if lang == "cpp" {
		for _, c := range cpp_compiler_candidates {
			if _, err := exec.LookPath(c); err == nil {
				return c
			}
		}
	}
	return ""
}
