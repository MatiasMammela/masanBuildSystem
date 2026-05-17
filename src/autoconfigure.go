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

func detect_language(proj *Project) (string, bool) {
	lang := ""
	hasAsm := false
	for _, src := range proj.Sources {
		switch src.Type {
		case ".cpp", ".cxx", ".cc":
			lang = "cpp"
		case ".c":
			if lang == "" { 
				lang = "c"
			}
		case ".asm", ".s":
			hasAsm = true
		}
	}
	return lang, hasAsm
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


func auto_configure_project(proj *Project) {
	lang, hasAsm := detect_language(proj)

	if lang == "" {
		fmt.Println("Warning: no source files found to detect language")
		return
	}

	if proj.Compiler == "" {
        proj.Compiler = detect_compiler(lang)
    }

	proj.CFlags = append_unique(proj.CFlags,default_flags(lang)...);

	
	if proj.Standard == "" {
		proj.CFlags = append(proj.CFlags,"-std="+default_standard(lang));
		proj.Standard = default_standard(lang);
	}else{
		proj.CFlags=append(proj.CFlags,"-std="+proj.Standard);
	}

    if proj.OS == "linux" && 
       (proj.Compiler == "gcc" || proj.Compiler == "g++" || 
        proj.Compiler == "clang" || proj.Compiler == "clang++") {
        proj.CFlags = append_unique(proj.CFlags, "-D_GNU_SOURCE")
    }
	
	if hasAsm {
		proj.Assembler = detect_assembler()
		proj.ASMFlags = append(proj.ASMFlags,default_flags("asm")...);
		proj.LFlags = append(proj.LFlags, "-no-pie")
	}
    
    if proj.Linker != "" && proj.Linker != proj.Compiler {
        proj.LinkerFlags = append([]string{"-fuse-ld=" + proj.Linker}, proj.LinkerFlags...)
        proj.Linker = proj.Compiler 
    } else {
        proj.Linker = proj.Compiler
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
