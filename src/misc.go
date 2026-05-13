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
	fmt.Println("OS:", p.OS)
    fmt.Println("Build Directory Path:", p.Build_dir_path)
	fmt.Println("Build file Path:", p.Build_file_path)
	fmt.Println("Build file dir Path:",p.Buildr_file_dir_path)
	fmt.Println("CWD:",p.Cwd)
	fmt.Println("Compiler:", p.Compiler)
	fmt.Println("Assembler:", p.Assembler);
	fmt.Println("Linker:", p.Linker);
	fmt.Println("Linking:", p.Linking);
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

	fmt.Println("LinkerFlags:")
    for _, f := range p.LinkerFlags {
        fmt.Println("  -", f)
    }

	fmt.Println("ASMFlags:")
    for _, f := range p.ASMFlags {
        fmt.Println("  -", f)
    }

    fmt.Println("Libraries:")
    for _, l := range p.Libraries {
		
		linkType := "\033[32mdynamic\033[0m"
		if l.Static {
			linkType = "\033[33mstatic\033[0m"
		}
		
		fmt.Printf("  - [%s] %s  %s\n", linkType, l.Libraries, l.Headers)
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



func get_pacman_package_options(name string) []string {
    cmd := exec.Command("pacman", "-Ss", name)
    out, err := cmd.Output()
    if err != nil || len(out) == 0 {
        return nil
    }
    var options []string
    lines := strings.Split(string(out), "\n")
    for _, line := range lines {
        // Package lines start with "repo/name version"
        if len(line) > 0 && !strings.HasPrefix(line, " ") {
            parts := strings.Fields(line)
            if len(parts) > 0 {
                full := parts[0]
                if idx := strings.Index(full, "/"); idx >= 0 {
                    options = append(options, full[idx+1:])
                }
            }
        }
    }
    return options
}

func get_apt_package_options(name string) []string {
    cmd := exec.Command("apt-cache", "search","--names-only", name)
    out, err := cmd.Output()
    if err != nil || len(out) == 0 {
        return nil
    }
    var options []string
    lines := strings.Split(string(out), "\n")
    for _, line := range lines {
        if len(line) > 0 {
            parts := strings.Fields(line)
            if len(parts) > 0 {
                options = append(options, parts[0])
            }
        }
    }
    return options
}

func get_package_options(pm, name string) []string {
    switch pm {
    case "apt":
        return get_apt_package_options(name)
    case "pacman":
        return get_pacman_package_options(name)
    default:
        return nil
    }
}


func prompt_package_selection(name string, options []string) string {
    if len(options) == 0 {
        return ""
    }
    if len(options) == 1 {
        return options[0]
    }
    fmt.Printf("Multiple packages found for '%s':\n", name)
    for i, opt := range options {
        fmt.Printf("  [%d] %s\n", i+1, opt)
    }
    fmt.Printf("Select package [1-%d] or 0 to skip: ", len(options))
    var choice int
    fmt.Scanln(&choice)
    if choice < 1 || choice > len(options) {
        return ""
    }
    return options[choice-1]
}

var (
	package_managers_candidates=[]string{"apt","pacman"}
)

func detect_package_manager()string{
	for _, pkgmngr := range package_managers_candidates {
		if _, err := exec.LookPath(pkgmngr); err == nil {
			return pkgmngr
		}
	}
	return "" 
}

func download_packages(name string) error {
    pm := detect_package_manager()
    if pm == "" {
        return fmt.Errorf("no supported package manager found")
    }

    options := get_package_options(pm, name)
    if len(options) == 0 {
        return fmt.Errorf("package '%s' not found in %s repositories", name, pm)
    }

    pkgname := prompt_package_selection(name, options)
    if pkgname == "" {
        return fmt.Errorf("no package selected for '%s'", name)
    }

    fmt.Printf("Install '%s' using %s? [Y/n]: ", pkgname, pm)
    var response string
    fmt.Scanln(&response)
    response = strings.TrimSpace(strings.ToLower(response))
    if response != "" && response != "y" && response != "yes" {
        return fmt.Errorf("user declined to install package '%s'", name)
    }

    var cmd *exec.Cmd
    switch pm {
    case "apt":
        cmd = exec.Command("sudo", "apt", "install", "-y", pkgname)
    case "pacman":
        cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", pkgname)
    default:
        return fmt.Errorf("unsupported package manager: %s", pm)
    }
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}



func check_static_libs_available(libs string) (bool, string) {
    // Parse library names from flags like "-lSDL2 -lncurses"
    parts := strings.Fields(libs)
    var missing []string
    for _, part := range parts {
        if strings.HasPrefix(part, "-l") {
            libname := "lib" + strings.TrimPrefix(part, "-l") + ".a"
            // Search common static lib locations
            searchPaths := []string{
                "/usr/lib",
                "/usr/lib/x86_64-linux-gnu",
                "/usr/local/lib",
            }
            found := false
            for _, path := range searchPaths {
                if _, err := os.Stat(filepath.Join(path, libname)); err == nil {
                    found = true
                    break
                }
            }
            if !found {
                missing = append(missing, libname)
            }
        }
    }
    if len(missing) > 0 {
        return false, strings.Join(missing, ", ")
    }
    return true, ""
}

func find_packages(names []string, static bool) []*Package {
    var result []*Package
    for _, name := range names {
        pkg := &Package{Name: name,Found:false,Static:static}
        // Check if package exists
        check := exec.Command("pkg-config", "--exists", name)
        if err := check.Run(); err != nil {
            msg("WARNING", "Package "+name+" not found!")
            if err := download_packages(name); err != nil {
                msg("ERROR", "Failed to install package "+name+": "+err.Error())
                result = append(result, pkg)
                continue
            }
            // Retry pkg-config after install
            if err := exec.Command("pkg-config", "--exists", name).Run(); err != nil {
                msg("ERROR", "Package "+name+" still not found after installation")
                result = append(result, pkg)
                continue
            }
        }
        // If found, fetch cflags and libs
        cflagsCmd := exec.Command("pkg-config", "--cflags-only-I", name)
        var libsCmd *exec.Cmd
        if static {
            libsCmd = exec.Command("pkg-config", "--libs", "--static", name)
        } else {
            libsCmd = exec.Command("pkg-config", "--libs", name)
        }

        var cflagsOut, libsOut bytes.Buffer
        cflagsCmd.Stdout = &cflagsOut
        libsCmd.Stdout = &libsOut
        _ = cflagsCmd.Run()
        _ = libsCmd.Run()
        libsStr := strings.TrimSpace(libsOut.String())

        if static {
            ok, missing := check_static_libs_available(libsStr)
            if !ok {
                msg("ERROR", fmt.Sprintf("Static libraries not found for '%s': %s", name, missing))
                result = append(result, pkg)
                continue
            }
        }
        pkg.Found = true
        pkg.Headers = strings.TrimSpace(cflagsOut.String())
        pkg.Libraries = libsStr
        result = append(result, pkg)
    }
    return result
}
