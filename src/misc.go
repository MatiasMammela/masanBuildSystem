package src

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (p *Project) Debug() {
    fmt.Println("=== Project Debug ===")
    fmt.Println("Name:", p.Name)
    fmt.Println("Build Directory Path:", p.Build_dir_path)
	fmt.Println("Build file Path:", p.Build_file_path)
	fmt.Println("Compiler:", p.Compiler)
	fmt.Println("Assembler:", p.Assembler);
    fmt.Println("Headers:")
    for _, h := range p.Headers {
        fmt.Println("  -", h.Path)
    }

    fmt.Println("Sources:")
    for _, s := range p.Sources {
        fmt.Println("  -", s.Cwd)
    }

    fmt.Println("CFlags:")
    for _, f := range p.CFlags {
        fmt.Println("  -", f)
    }
    fmt.Println("LFlags:")
    for _, f := range p.LFlags {
        fmt.Println("  -", f)
    }

	fmt.Println("ASMFlags:")
    for _, f := range p.ASMFlags {
        fmt.Println("  -", f)
    }

    fmt.Println("Libraries:")
    for _, l := range p.Libraries {
        fmt.Println("  -", l.Libraries, "  " ,l.Headers)
    }
    fmt.Println("=====================")
}
func list_bound_sources(srcs []*File,project *Project){
	for _, s := range srcs {
		msg("OK", "Bound Source to "+project.Name+"        ["+s.Name+"]")
	}
}

func list_bound_headers(headers []*Directory,project *Project){
	for _, h := range headers {
		msg("OK", "Bound Header to "+project.Name+"        ["+h.Name+"]")
	}
}
func list_bound_packages(pkgs []*Package,project *Project){
	for _, p := range pkgs {
		msg("OK", "Bound Library to "+project.Name+"       ["+p.Name+"]")
	}
}

func find_directories(patterns []string) []*Directory {
	var result []*Directory
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		if len(matches) == 0 {
			// Still record something if no matches
			msg("WARNING","Directory " + filepath.Base(pattern) + " not found!");
			result = append(result, &Directory{
				Name:  filepath.Base(pattern),
				Path:  pattern,
				Found: false,
			})
			continue
		}

		for _, match := range matches {
			info,err := os.Stat(match)
			if err != nil  {

				msg("WARNING","Cant stat " + filepath.Base(pattern) + "!");
				result = append(result, &Directory{
					Name:  filepath.Base(match),
					Path:  match,
					Found: false,
				})
				continue
			} 


			if !info.IsDir() {
				msg("WARNING","Directory " + filepath.Base(pattern) + " is a file!");
				continue
			}
			abs, _ := filepath.Abs(match)
			result = append(result, &Directory{
				Name:  filepath.Base(match),
				Path:  abs,
				Found: true,
			})
		}
	}
	return result
}
func copy_file(src string, dst string) error {
    source, err := os.Open(src)
    if err != nil {
        return err
    }
    defer source.Close()

    // Check if destination is a directory
    dstInfo, err := os.Stat(dst)
    if err == nil && dstInfo.IsDir() {
        dst = filepath.Join(dst, filepath.Base(src))
    }

    destFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destFile.Close()

    _, err = io.Copy(destFile, source)
    if err == nil {
        msg("OK", "Copied "+src+" -> "+dst)
    }
    return err
}


func copy_directory(srcDir string, dstDir string) error {
    dstDir = filepath.Join(dstDir, filepath.Base(srcDir))

    err := os.MkdirAll(dstDir, 0755)
    if err != nil {
        return fmt.Errorf("failed to create destination dir: %w", err)
    }

    entries, err := os.ReadDir(srcDir)
    if err != nil {
        return fmt.Errorf("failed to read source dir: %w", err)
    }

    for _, entry := range entries {
        srcPath := filepath.Join(srcDir, entry.Name())
        dstPath := filepath.Join(dstDir, entry.Name())

        if entry.IsDir() {
            err = copy_directory(srcPath, dstDir) 
        } else {
            err = copy_file(srcPath, dstPath)
        }
        if err != nil {
            fmt.Println("Error copying:", srcPath, "->", dstPath, ":", err)
        }
    }
    return nil
}


func find_files(patterns []string) []*File {
	var result []*File
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		if len(matches) == 0 {
			msg("WARNING","File " + filepath.Base(pattern) + " not found!");
			result = append(result, &File{
				Name:  filepath.Base(pattern),
				Type:  filepath.Ext(pattern),
				Cwd:   pattern,
				Found: false,
			})
			continue
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				msg("WARNING","Cant stat File " + filepath.Base(pattern) + "!");
				result = append(result, &File{
					Name:  filepath.Base(match),
					Type:  filepath.Ext(match),
					Cwd:   match,
					Found: false,
				})
				continue
			}

			if info.IsDir() {
				// skip directories
				msg("WARNING","File " + filepath.Base(pattern) + " is a directory!");
				continue
			}

			abs, _ := filepath.Abs(match)
			ext := filepath.Ext(match)

			result = append(result, &File{
				Name:  filepath.Base(match),
				Type:  ext,
				Cwd:   abs,
				Found: true,
			})
		}
	}
	return result
}

func find_packages(names []string) []*Package {
	var result []*Package
	for _, name := range names {
		pkg := &Package{Name: name,Found:false}

		// Check if package exists
		check := exec.Command("pkg-config", "--exists", name)
		if err := check.Run(); err != nil {
			msg("WARNING","Package "+name+" not found!")
			result = append(result, pkg)
			continue
		}

		// If found, fetch cflags and libs
		cflagsCmd := exec.Command("pkg-config", "--cflags-only-I", name)
		libsCmd := exec.Command("pkg-config", "--libs", name)

		var cflagsOut, libsOut bytes.Buffer
		cflagsCmd.Stdout = &cflagsOut
		libsCmd.Stdout = &libsOut

		_ = cflagsCmd.Run()
		_ = libsCmd.Run()

		pkg.Found = true
		pkg.Headers = strings.TrimSpace(cflagsOut.String())
		pkg.Libraries = strings.TrimSpace(libsOut.String())

		result = append(result, pkg)
	}
	return result
}