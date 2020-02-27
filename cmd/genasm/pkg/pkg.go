// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"dappui.com/cmd/genasm/asm"
	"github.com/mmcloughlin/avo/pass"
	"github.com/mmcloughlin/avo/printer"
)

var (
	indicators = []string{"blake3", "kangaroo12", "osexit"}
	registry   = map[string][]*Entry{}
)

// Entry represents the build configuration for a specific file in a package.
type Entry struct {
	Constraints string
	File        string
	Generator   func(*asm.Context)
	Stub        bool
}

// Generate will generate all the files for the given packages.
func Generate(pkgs []string) {
	root, err := findRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! ERROR: Failed to find the root source directory:\n\n\t%s\n\n", err)
		os.Exit(1)
	}
	for _, pkg := range pkgs {
		if err := gen(root, pkg); err != nil {
			fmt.Fprintf(os.Stderr, "!! ERROR: Failed to generate asm for pkg/%s:\n\n\t%s\n\n", pkg, err)
			os.Exit(1)
		}
	}
}

// Register registers the given generator function with the
func Register(pkg string, entry *Entry) {
	registry[pkg] = append(registry[pkg], entry)
}

func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		failed := false
		for _, indicator := range indicators {
			path := filepath.Join(dir, "pkg", indicator)
			_, err := os.Stat(path)
			if err != nil {
				if os.IsNotExist(err) {
					failed = true
					continue
				}
				return "", err
			}
		}
		if !failed {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not find pkg directory with %s subdirectories", indicators)
}

func gen(root string, pkg string) error {
	entries, ok := registry[pkg]
	if !ok {
		return fmt.Errorf("unable to find registry entry for %q", pkg)
	}
	for _, e := range entries {
		if err := genFile(root, pkg, e); err != nil {
			return fmt.Errorf("failed to generate %s: %s", e.File, err)
		}
		if e.Stub {
			fmt.Printf(">> Successfully wrote pkg/%s/%s.go\n", pkg, e.File)
		}
		fmt.Printf(">> Successfully wrote pkg/%s/%s.s\n", pkg, e.File)
	}
	return nil
}

func genFile(root string, pkg string, e *Entry) error {
	ctx := asm.NewContext()
	ctx.Package("dappui.com/pkg/" + pkg)
	if e.Constraints == "" {
		e.Constraints = "amd64,!gccgo"
	}
	ctx.ConstraintExpr(e.Constraints)
	e.Generator(ctx)
	f, err := ctx.Result()
	if err != nil {
		return err
	}
	pcfg := printer.Config{
		Argv: []string{"genasm", pkg},
		Pkg:  pkg,
	}
	out, err := os.Create(filepath.Join(root, "pkg", pkg, e.File+".s"))
	if err != nil {
		return err
	}
	passes := []pass.Interface{
		pass.Compile,
		&pass.Output{
			Printer: printer.NewGoAsm(pcfg),
			Writer:  out,
		},
	}
	if e.Stub {
		stub, err := os.Create(filepath.Join(root, "pkg", pkg, e.File+".go"))
		if err != nil {
			return err
		}
		passes = append(passes, &pass.Output{
			Printer: printer.NewStubs(pcfg),
			Writer:  stub,
		})
	}
	p := pass.Concat(passes...)
	if err := p.Execute(f); err != nil {
		return err
	}
	return nil
}
