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
	fmt.Println("C Compiler:", p.CCompiler)
	fmt.Println("CXX Compiler:", p.CXXCompiler)
	fmt.Println("C Standard:", p.CStandard)
	fmt.Println("CXX Standard:", p.CXXStandard)
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
    fmt.Println("CXXFlags:")
    for _, f := range p.CXXFlags {
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

		var line string
		if l.Version != "" {
			line = fmt.Sprintf("  - [%s][%s][%s] %s\n", linkType, l.Name, l.Version, libs)
		} else {
			line = fmt.Sprintf("  - [%s][%s] %s\n", linkType, l.Name, libs)
		}
		fmt.Print(line)
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
	linux_library_paths=[]string{
		"/usr/lib",
        "/usr/lib/x86_64-linux-gnu",
        "/usr/local/lib",
	}
	linux_header_paths=[]string{
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


func get_pkg_libs(name string, static bool) string {
    resolved := pkg_config_resolve(name)
    if resolved == "" {
        return ""
    }
    var libsCmd *exec.Cmd
    if static {
        libsCmd = exec.Command("pkg-config", "--libs", "--static", resolved)
    } else {
        libsCmd = exec.Command("pkg-config", "--libs", resolved)
    }
    var libsOut bytes.Buffer
    libsCmd.Stdout = &libsOut
    _ = libsCmd.Run()
    return strings.TrimSpace(libsOut.String())
}

func get_pkg_cflags(name string) string {
    resolved := pkg_config_resolve(name)
    if resolved == "" {
        return ""
    }
    cflagsCmd := exec.Command("pkg-config", "--cflags", resolved)
    var cFlagsOut bytes.Buffer
    cflagsCmd.Stdout = &cFlagsOut
    _ = cflagsCmd.Run()
    return strings.TrimSpace(cFlagsOut.String())
}

func pkg_config_resolve(name string) string {
    candidates := []string{
        name,
        strings.ToLower(name),
        strings.NewReplacer("lib", "").Replace(strings.ToLower(name)),
        strings.NewReplacer("_", "-").Replace(strings.ToLower(name)),
        "lib" + strings.ToLower(name),
        "lib" + strings.NewReplacer("_", "-").Replace(strings.ToLower(name)),
    }

    lower := strings.ToLower(name)

	stripped := strings.NewReplacer("-", "", "_", "").Replace(lower)
    if stripped != lower {
        candidates = append(candidates, stripped, "lib"+stripped)
    }
	
    i := len(lower)
    for i > 0 && lower[i-1] >= '0' && lower[i-1] <= '9' {
        i--
    }
    if i > 0 && i < len(lower) {
        base := strings.TrimRight(lower[:i], "-_") 
        num := lower[i:]
        candidates = append(candidates,
            base+"-"+num,
            base+"_"+num,
            "lib"+base+"-"+num,
            "lib"+base+"_"+num,
            base+"+-"+num+".0",
            "lib"+base+"+-"+num+".0",
        )
    }
	
    for _, c := range candidates {
        if exec.Command("pkg-config", "--exists", c).Run() == nil {
            return c
        }
    }

    matches := find_ambiguous_matches(name)
    if len(matches) == 1 {
        return matches[0]
    }

    return ""
}

func get_lib_search_paths(resolved string) []string {
    seen := make(map[string]bool)
    var dirs []string

    add := func(d string) {
        d = strings.TrimSpace(d)
        if d != "" && !seen[d] {
            seen[d] = true
            dirs = append(dirs, d)
        }
    }

    cmd := exec.Command("pkg-config", "--static", "--libs-only-L", resolved)
    var out bytes.Buffer
    cmd.Stdout = &out
    if cmd.Run() == nil {
        for _, tok := range strings.Fields(out.String()) {
            add(strings.TrimPrefix(tok, "-L"))
        }
    }

    libdirCmd := exec.Command("pkg-config", "--variable=libdir", resolved)
    var libdirOut bytes.Buffer
    libdirCmd.Stdout = &libdirOut
    if libdirCmd.Run() == nil {
        add(strings.TrimSpace(libdirOut.String()))
    }

    return dirs
}

func get_all_lib_names(resolved string) []string {
    cmd := exec.Command("pkg-config", "--static", "--libs-only-l", resolved)
    var out bytes.Buffer
    cmd.Stdout = &out
    if cmd.Run() != nil {
        return nil
    }
    return strings.Fields(out.String())
}

func has_static_lib(resolved string) bool {
    searchDirs := get_lib_search_paths(resolved)
    if len(searchDirs) == 0 {
        return false
    }

    libFlags := get_all_lib_names(resolved)
    if len(libFlags) == 0 {
        return false
    }

    for _, l := range libFlags {
        name := strings.TrimPrefix(l, "-l")
        found := false
        for _, dir := range searchDirs {
            staticPath := filepath.Join(dir, "lib"+name+".a")
            if _, err := os.Stat(staticPath); err == nil {
                found = true
                break
            }
        }
        if !found {
            return false
        }
    }
    return true
}


func ensure_static_package(name string) (*PackageInfo, error) {
    // Ensure the package exists and get its info.
    info, err := ensure_dynamic_package(name)
    if err != nil {
        return nil, err
    }

    resolved := pkg_config_resolve(name)
    if resolved == "" {
        return nil, fmt.Errorf("no pkg-config entry for '%s'", name)
    }

	if has_static_lib(resolved) {
        info.Libraries = get_pkg_libs(resolved, true)
        return info, nil
    }

    msg("WARNING", fmt.Sprintf("Static libraries not available for '%s'", name))
    fmt.Printf("'%s' has no static pkg-config support. Install static package? [Y/n]: ", name)

    var response string
    fmt.Scanln(&response)
    response = strings.TrimSpace(strings.ToLower(response))

    if response != "" && response != "y" && response != "yes" {
        return nil, fmt.Errorf("user declined static package for '%s'", name)
    }

    if err := download_static_package(name); err != nil {
        return nil, err
    }

    // Re-check after installation.
    if !has_static_lib(resolved) {
        return nil, fmt.Errorf("static archive still missing for '%s' after install", name)
    }

    info.Libraries = get_pkg_libs(resolved, true)
    return info, nil
}

type PackageInfo struct {
	Headers   string
	Libraries string
    Version   string
}

func get_pkg_version(name string) string {
    resolved := pkg_config_resolve(name)
    if resolved == "" {
        return ""
    }

    out, err := exec.Command("pkg-config", "--modversion", resolved).Output()
    if err != nil {
        return ""
    }

    return strings.TrimSpace(string(out))
}

func find_dynamic_package(name string) (*PackageInfo, error) {
	if resolved := pkg_config_resolve(name); resolved != "" {
		return &PackageInfo{
			Headers:   get_pkg_cflags(resolved),
			Libraries: get_pkg_libs(resolved, false),
			Version: get_pkg_version(resolved),
		}, nil
	}
	
	headers, libs, err := cmake_find_package(name)
	if err == nil {
		return &PackageInfo{
			Headers:   headers,
			Libraries: libs,
			Version: "",
		}, nil
	}

	return nil, fmt.Errorf("package '%s' not found", name)
}

func compareVersions(a, b string) int {
    parse := func(v string) []int {
        parts := strings.Split(v, ".")
        nums := make([]int, len(parts))

        for i, p := range parts {
            n := 0
            for _, c := range p {
                if c < '0' || c > '9' {
                    break
                }
                n = n*10 + int(c-'0')
            }
            nums[i] = n
        }

        return nums
    }

    va := parse(a)
    vb := parse(b)

    n := len(va)
    if len(vb) > n {
        n = len(vb)
    }

    for i := 0; i < n; i++ {
        ai, bi := 0, 0
        if i < len(va) {
            ai = va[i]
        }
        if i < len(vb) {
            bi = vb[i]
        }

        switch {
        case ai < bi:
            return -1
        case ai > bi:
            return 1
        }
    }

    return 0
}

func checkVersion(installed, op, required string) error {
    if op == "" || required == "" || installed == "" {
        return nil
    }

    cmp := compareVersions(installed, required)

    ok := false

    switch op {
    case "==":
        ok = cmp == 0
    case "!=":
        ok = cmp != 0
    case ">":
        ok = cmp > 0
    case ">=":
        ok = cmp >= 0
    case "<":
        ok = cmp < 0
    case "<=":
        ok = cmp <= 0
    default:
        return fmt.Errorf("unknown version operator '%s'", op)
    }

    if !ok {
        return fmt.Errorf(
            "package version mismatch: installed %s, required %s %s",
            installed, op, required,
        )
    }

    return nil
}


func find_ambiguous_matches(name string) []string {
    cmd := exec.Command("pkg-config", "--list-all")
    var out bytes.Buffer
    cmd.Stdout = &out
    if cmd.Run() != nil {
        return nil
    }

    lower := strings.ToLower(name)
    var matches []string
    for _, line := range strings.Split(out.String(), "\n") {
        fields := strings.Fields(line)
        if len(fields) == 0 {
            continue
        }
        pkgName := fields[0]
        if strings.Contains(strings.ToLower(pkgName), lower) {
            matches = append(matches, pkgName)
        }
    }
    return matches
}


func ensure_dynamic_package(name string) (*PackageInfo, error) {

    if info, err := find_dynamic_package(name); err == nil {
        return info, nil
    }

    pm := detect_package_manager()

	switch pm {
	case "pacman":
		if exec.Command("pacman", "-Q", name).Run() == nil {
			if matches := find_ambiguous_matches(name); len(matches) > 1 {
				return nil, fmt.Errorf("'%s' is ambiguous - matches multiple pkg-config entries:\n%s \nPlease specify the exact name in your build file", name, strings.Join(matches,"\n"))
			}
			return nil, fmt.Errorf("'%s' is installed but might not have pkg-config or CMake support with the name %s. Use lflags/cflags instead", name, name)
		}

	case "apt":
		cmd := exec.Command("dpkg-query", "-W", "-f=${Status}", name)
		out, _ := cmd.Output()
		if strings.Contains(string(out), "install ok installed") {
			if matches := find_ambiguous_matches(name); len(matches) > 1 {
				return nil, fmt.Errorf("'%s' is ambiguous - matches multiple pkg-config entries:\n %s \nPlease specify the exact name in your build file", name, strings.Join(matches, "\n"))
			}
			return nil, fmt.Errorf("'%s' is installed but might not have pkg-config or CMake support with name %s. Use lflags/cflags instead", name, name)
		}
	}

    msg("WARNING", "Package "+name+" not found!")
    fmt.Printf("Try to install '%s'? [Y/n]: ", name)

    var response string
    fmt.Scanln(&response)
    response = strings.TrimSpace(strings.ToLower(response))

    if response != "" && response != "y" && response != "yes" {
        return nil, fmt.Errorf("user declined to install package '%s'", name)
    }

    if _, err := download_packages(name); err != nil {
        return nil, err
    }

    info, err := find_dynamic_package(name)
    if err != nil {
        return nil, fmt.Errorf("'%s' was installed but could not be found via pkg-config or CMake", name)
    }

    return info, nil
}

func cmake_find_package(name string) (headers string, libraries string, err error) {
	compileCmd := exec.Command(
		"cmake",
		"--find-package",
		"-DNAME="+name,
		"-DCOMPILER_ID=GNU",
		"-DLANGUAGE=CXX",
		"-DMODE=COMPILE",
	)

	compileOut, err := compileCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("package '%s' not found", name)
	}

	linkCmd := exec.Command(
		"cmake",
		"--find-package",
		"-DNAME="+name,
		"-DCOMPILER_ID=GNU",
		"-DLANGUAGE=CXX",
		"-DMODE=LINK",
	)

	linkOut, err := linkCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("package '%s' found but link information unavailable", name)
	}

	return strings.TrimSpace(string(compileOut)),
		strings.TrimSpace(string(linkOut)),
		nil
}




func find_packages(requests []PackageRequest, static bool) []*Package {
    var result []*Package

    for _, req := range requests {
        pkg := &Package{
            Name:   req.Name,
            Found:  false,
            Static: static,
        }

        var (
            info *PackageInfo
            err  error
        )

        if static {
            info, err = ensure_static_package(req.Name)
        } else {
            info, err = ensure_dynamic_package(req.Name)
        }

        if err != nil {
            msg("ERROR", err.Error())
            result = append(result, pkg)
            continue
        }

        if err := checkVersion(info.Version, req.Operator, req.Version); err != nil {
            msg("ERROR", err.Error() + " For " + req.Name)
            result = append(result, pkg)
            continue
        }

        pkg.Found = true
        pkg.Headers = info.Headers
        pkg.Libraries = info.Libraries
		pkg.Version = info.Version
        result = append(result, pkg)
    }

    return result
}

func parsePackage(s string) PackageRequest {
    operators := []string{">=", "<=", "==", ">", "<"}

    for _, op := range operators {
        if idx := strings.Index(s, op); idx != -1 {
            return PackageRequest{
                Name:     strings.TrimSpace(s[:idx]),
                Operator: op,
                Version:  strings.TrimSpace(s[idx+len(op):]),
            }
        }
    }

    return PackageRequest{
        Name: strings.TrimSpace(s),
    }
}

func make_package_manual(name, headers, libraries, version string, static bool) *Package {
    pkg := &Package{
        Name:      name,
        Found:     false,
        Static:    static,
        Headers:   headers,
        Libraries: libraries,
        Version:   version,
    }

    if name == "" {
        msg("ERROR", "manual package: 'name' is required")
        return pkg
    }
    if libraries == "" {
        msg("ERROR", fmt.Sprintf("manual package '%s': 'libraries' is required", name))
        return pkg
    }

    // Sanity check header include paths (-I flags)
    for _, tok := range strings.Fields(headers) {
        if !strings.HasPrefix(tok, "-I") {
            continue
        }
        dir := strings.TrimPrefix(tok, "-I")
        if dir == "" {
            continue
        }
        if info, err := os.Stat(dir); err != nil {
            msg("ERROR", fmt.Sprintf("manual package '%s': header path '%s' does not exist", name, dir))
            return pkg
        } else if !info.IsDir() {
            msg("ERROR", fmt.Sprintf("manual package '%s': header path '%s' is not a directory", name, dir))
            return pkg
        }
    }

    // Sanity check library search paths (-L flags), if any were given
    var libDirs []string
    for _, tok := range strings.Fields(libraries) {
        if !strings.HasPrefix(tok, "-L") {
            continue
        }
        dir := strings.TrimPrefix(tok, "-L")
        if dir == "" {
            continue
        }
        if info, err := os.Stat(dir); err != nil {
            msg("ERROR", fmt.Sprintf("manual package '%s': library path '%s' does not exist", name, dir))
            return pkg
        } else if !info.IsDir() {
            msg("ERROR", fmt.Sprintf("manual package '%s': library path '%s' is not a directory", name, dir))
            return pkg
        }
        libDirs = append(libDirs, dir)
    }

    searchDirs := append([]string{}, libDirs...)
    searchDirs = append(searchDirs, linux_library_paths...)

	for _, tok := range strings.Fields(libraries) {
		if !strings.HasPrefix(tok, "-l") {
			continue
		}
		libName := strings.TrimPrefix(tok, "-l")
		found := false
		for _, dir := range searchDirs {
			if static {
				if _, err := os.Stat(filepath.Join(dir, "lib"+libName+".a")); err == nil {
					found = true
					break
				}
			} else {
				if _, err := os.Stat(filepath.Join(dir, "lib"+libName+".so")); err == nil {
					found = true
					break
				}
				entries, err := os.ReadDir(dir)
				if err != nil {
					continue
				}
				prefix := "lib" + libName + ".so"
				for _, e := range entries {
					if strings.HasPrefix(e.Name(), prefix) {
						rest := strings.TrimPrefix(e.Name(), prefix)
						if rest == "" || strings.HasPrefix(rest, ".") {
							found = true
							break
						}
					}
				}
			}
			if found {
				break
			}
		}
		if !found {
			msg("ERROR", fmt.Sprintf(
				"manual package '%s': could not find 'lib%s%s' in any known library path (checked: %s)",
				name, libName, map[bool]string{true: ".a", false: ".so"}[static], strings.Join(searchDirs, ", "),
			))
			return pkg
		}
	}

    pkg.Found = true
    return pkg
}





