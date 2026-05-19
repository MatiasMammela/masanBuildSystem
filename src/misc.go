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
    fmt.Println("Target type:", p.Target_type)
    fmt.Println("Build Directory Path:", p.Build_dir_path)
	fmt.Println("Build file Path:", p.Build_file_path)
	fmt.Println("Build file dir Path:",p.Build_file_dir_path)
	fmt.Println("Bin dir Path:",p.Bin_dir_path)
	fmt.Println("CWD:",p.Cwd)
	fmt.Println("Compiler:", p.Compiler)
	fmt.Println("Standard:", p.Standard)
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
		libs := l.Libraries
		if libs == "" {
			libs = "(header only)"
		}
		fmt.Printf("  - [%s][%s] %s\n", linkType, l.Name, libs)
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


func normalize_package_name(name string) string {
    // Replace _ with [-_] to match both dash and underscore variants
    return strings.ReplaceAll(strings.ToLower(name), "_", "[-_]")
}

func get_package_options(pm, name string) []string {
    switch pm {
    case "apt":
        return get_apt_package_options(normalize_package_name(name))
    case "pacman":
        return get_pacman_package_options(normalize_package_name(name))
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
	linux_library_paths[]string{
		"/usr/lib",
        "/usr/lib/x86_64-linux-gnu",
        "/usr/local/lib"
	}
	linux_header_paths[]string{
		"/usr/include/",
	}

)

func detect_package_manager()string{
	for _, pkgmngr := range package_managers_candidates {
		if _, err := exec.LookPath(pkgmngr); err == nil {
			return pkgmngr
		}
	}
	return "" 
}

func download_packages(name string) (string, error) {
    pm := detect_package_manager()
    if pm == "" {
        return "", fmt.Errorf("no supported package manager found")
    }
    options := get_package_options(pm, name)
    if len(options) == 0 {
        return "", fmt.Errorf("package '%s' not found in %s repositories", name, pm)
    }
    pkgname := prompt_package_selection(name, options)
    if pkgname == "" {
        return "", fmt.Errorf("no package selected for '%s'", name)
    }
    fmt.Printf("Install '%s' using %s? [Y/n]: ", pkgname, pm)
    var response string
    fmt.Scanln(&response)
    response = strings.TrimSpace(strings.ToLower(response))
    if response != "" && response != "y" && response != "yes" {
        return "", fmt.Errorf("user declined to install package '%s'", name)
    }
    var cmd *exec.Cmd
    switch pm {
    case "apt":
        cmd = exec.Command("sudo", "apt", "install", "-y", pkgname)
    case "pacman":
        cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", pkgname)
    default:
        return "", fmt.Errorf("unsupported package manager: %s", pm)
    }
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return pkgname, cmd.Run()
}


func download_static_package(name string) error {
    pm := detect_package_manager()
    if pm == "" {
        return fmt.Errorf("no supported package manager found")
    }

    staticName := name + "-static"
    options := get_package_options(pm, staticName)
    if len(options) == 0 {
        return fmt.Errorf("no static package found for '%s' in %s repositories", name, pm)
    }

    pkgname := prompt_package_selection(staticName, options)
    if pkgname == "" {
        return fmt.Errorf("no package selected for '%s'", name)
    }

    fmt.Printf("Install '%s' using %s? [Y/n]: ", pkgname, pm)
    var response string
    fmt.Scanln(&response)
    response = strings.TrimSpace(strings.ToLower(response))
    if response != "" && response != "y" && response != "yes" {
        return fmt.Errorf("user declined to install static package '%s'", name)
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
    parts := strings.Fields(libs)
    var missing []string
    for _, part := range parts {
        if strings.HasPrefix(part, "-l") {
            libname := "lib" + strings.TrimPrefix(part, "-l") + ".a"
            found := false
            for _, path := range linux_library_paths {
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


func get_pkg_libs(name string, static bool) string {
    var libsCmd *exec.Cmd
    if static {
        libsCmd = exec.Command("pkg-config", "--libs", "--static", name)
    } else {
        libsCmd = exec.Command("pkg-config", "--libs", name)
    }
    var libsOut bytes.Buffer
    libsCmd.Stdout = &libsOut
    _ = libsCmd.Run()
    return strings.TrimSpace(libsOut.String())
}


func get_pkg_cflags(name string) string {
    cflagsCmd := exec.Command("pkg-config", "--cflags", name)
    var cFlagsOut bytes.Buffer
    cflagsCmd.Stdout = &cFlagsOut
    _ = cflagsCmd.Run()
    return strings.TrimSpace(cFlagsOut.String())
}

func ensure_static_package(name string) error {
    // First ensure dynamic package exists
    if err := ensure_dynamic_package(name); err != nil {
        return err
    }
    // Now check if static libs are bundled in the dynamic package
    if ok, _ := check_static_libs_available(get_pkg_libs(name, true)); ok {
        return nil
    }
    // Static libs not found. ask user to install static package
    msg("WARNING", fmt.Sprintf("Static libraries not found for '%s'", name))
    fmt.Printf("'%s' is installed but has no static libraries. Try to install them? [Y/n]: ", name)
    var response string
    fmt.Scanln(&response)
    response = strings.TrimSpace(strings.ToLower(response))
    if response != "" && response != "y" && response != "yes" {
        return fmt.Errorf("user declined to install static libraries for '%s'", name)
    }
    if err := download_static_package(name); err != nil {
        return err
    }
    if ok, missing := check_static_libs_available(get_pkg_libs(name, true)); !ok {
        return fmt.Errorf("static libraries still not found: %s", missing)
    }
    return nil
}

func ensure_dynamic_package(name string) error {
    if exec.Command("pkg-config", "--exists", name).Run() == nil {
        return nil
    }
    pm := detect_package_manager()
    switch pm {
    case "pacman":
        if exec.Command("pacman", "-Q", name).Run() == nil {
            return fmt.Errorf("'%s' is installed but has no pkg-config support. Use lflags/cflags instead", name)
        }
    case "apt":
        cmd := exec.Command("dpkg-query", "-W", "-f=${Status}", name)
        out, _ := cmd.Output()
        if strings.Contains(string(out), "install ok installed") {
            return fmt.Errorf("'%s' is installed but has no pkg-config support. Use lflags/cflags instead", name)
        }
    }
    msg("WARNING", "Package "+name+" not found!")
    fmt.Printf("Try to install '%s'? [Y/n]: ", name)
    var response string
    fmt.Scanln(&response)
    response = strings.TrimSpace(strings.ToLower(response))
    if response != "" && response != "y" && response != "yes" {
        return fmt.Errorf("user declined to install package '%s'", name)
    }
    installed, err := download_packages(name)
    if err != nil {
        return err
    }
    if exec.Command("pkg-config", "--exists", name).Run() != nil {
        return fmt.Errorf("'%s' was installed but has no pkg-config support. Use lflags/cflags instead", installed)
    }
    return nil
}

func find_packages(names []string, static bool) []*Package {
    var result []*Package
    for _, name := range names {
        pkg := &Package{Name: name,Found:false,Static:static}

        var err error
        if static {
            err = ensure_static_package(name)
        } else {
            err = ensure_dynamic_package(name)
        }

        if err != nil {
            msg("ERROR", err.Error())
            result = append(result, pkg)
            continue
        }
		
        pkg.Found = true
        pkg.Headers = get_pkg_cflags(name);
        pkg.Libraries = get_pkg_libs(name,static);
        result = append(result, pkg)
    }
    return result
}

func find_library_file(name string, static bool) bool {
	if static {
		for _, path := range linux_library_paths {
			if _, err := os.Stat(filepath.Join(path, "lib"+name+".a")); err == nil {
				return true
			}
		}

		return false
	}

	cmd := exec.Command("ldconfig", "-p")

	out, err := cmd.Output()
	if err != nil {
		return false
	}


	for _, line := range strings.Split(string(out), "\n") {
        if strings.Contains(strings.ToLower(line), strings.ToLower("lib"+name)) {
            return true
        }
    }
    return false
}

func find_header_path(name string) string {
    cmd := exec.Command("find", "/usr/include", "-name", name+".h", "-o", "-name", name+".hpp")
    out, err := cmd.Output()
    if err != nil || len(out) == 0 {
        return ""
    }
    // Return the directory of the first found header
    headerPath := strings.TrimSpace(strings.Split(string(out), "\n")[0])
    return "-I" + filepath.Dir(headerPath)
}

func glob_libraries(names []string, static bool) []*Package {
	var result []*Package

	for _, name := range names {
		pkg := &Package{
			Name:   name,
			Found:  false,
			Static: static,
		}

		if !find_library_file(name, static) {
			if static {
				msg(
					"ERROR",
					fmt.Sprintf(
						"Static library '%s' not found in system paths",
						name,
					),
				)
			} else {
				msg(
					"ERROR",
					fmt.Sprintf(
						"Dynamic library '%s' not found in system paths",
						name,
					),
				)
			}

			result = append(result, pkg)
			continue
		}

		pkg.Found = true
		pkg.Libraries = "-l" + name
		pkg.Headers = find_header_path(name)

		result = append(result, pkg)
	}

	return result
}
