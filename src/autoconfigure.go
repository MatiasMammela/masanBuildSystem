package src

import (
	"fmt"
	"os/exec"
)

var (
	c_auto_flags          = []string{"-Wall", "-Wextra", "-O2"}
	cpp_auto_flags        = []string{"-Wall", "-Wextra", "-O2", "-std=c++17"}
	asm_auto_flags        = []string{"-f", "elf64"}
	assembler_candidates  = []string{"nasm", "as"}
	c_compiler_candidates = []string{"gcc", "clang", "cc"}
	cpp_compiler_candidates = []string{"g++", "clang++", "c++"}
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

	// Set compiler
	proj.Compiler = detect_compiler(lang)

	proj.CFlags = append(proj.CFlags,default_flags(lang)...); 
	

	if hasAsm {
		proj.Assembler = detect_assembler()
		proj.ASMFlags = append(proj.ASMFlags,default_flags("asm")...);
		proj.LFlags = append(proj.LFlags, "-no-pie")
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