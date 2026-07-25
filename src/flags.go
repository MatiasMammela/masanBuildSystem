package src

import (
	"fmt"
	"strings"
)

type flagSpec struct {
	Name   string
	IsBool bool
	Usage  string
}

var allFlags = []flagSpec{
	{Name: "builddir", IsBool: false, Usage: "build directory path for the build"},
	{Name: "installdir", IsBool: false, Usage: "installation directory path"},
	{Name: "generate_compdb", IsBool: true, Usage: "generate compile_commands.json"},
}

var commandFlags = map[string][]string{
	"configure": {},
	"build":     {"builddir"},
	"run":       {"builddir", "generate_compdb"},
	"install":   {"builddir", "installdir"},
	"version":   {},
	"help":      {},
}

func flagLookup(name string) (flagSpec, bool) {
	for _, f := range allFlags {
		if f.Name == name {
			return f, true
		}
	}
	return flagSpec{}, false
}

func isAllowed(command, name string) bool {
	allowed, ok := commandFlags[command]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == name {
			return true
		}
	}
	return false
}

func applyFlag(name, value string) error {
	switch name {
	case "builddir":
		GlobalFlags.builddir = value
	case "installdir":
		GlobalFlags.installdir = value
	case "generate_compdb":
		GlobalFlags.generate_compdb = true
	default:
		return fmt.Errorf("internal error: no handler registered for flag %q", name)
	}
	return nil
}

func ParseFlags(command string, args []string) ([]string, error) {
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}

		name := strings.TrimLeft(arg, "-")
		if name == "" {
			return nil, fmt.Errorf("invalid flag: %q", arg)
		}

		value := ""
		hasValue := false
		if eq := strings.Index(name, "="); eq != -1 {
			value = name[eq+1:]
			name = name[:eq]
			hasValue = true
		}

		spec, known := flagLookup(name)
		if !known {
			return nil, fmt.Errorf("unknown flag: --%s", name)
		}
		if !isAllowed(command, name) {
			return nil, fmt.Errorf("flag --%s is not valid for command %q", name, command)
		}

		if spec.IsBool {
			if hasValue {
				return nil, fmt.Errorf("flag --%s does not take a value", name)
			}
			value = "true"
		} else if !hasValue {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag --%s requires a value", name)
			}
			i++
			value = args[i]
		}

		if err := applyFlag(name, value); err != nil {
			return nil, err
		}
	}

	return positional, nil
}
