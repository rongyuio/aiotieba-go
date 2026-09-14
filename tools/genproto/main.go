// Command genproto generates the Go protobuf bindings for every .proto file of
// the project. Each source directory becomes its own Go package:
//
//	protobuf/*.proto             -> ./protobuf
//	api/<name>/protobuf/*.proto  -> ./api/<name>/protobuf
//
// The .proto files declare no go_package option and import each other by bare
// file name, so a -I include path plus a full set of M mappings is used.
//
// Usage:
//
//	go run ./tools/genproto
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const modulePath = "github.com/rongyuio/aiotieba-go"

type unit struct {
	srcDir  string   // directory that holds the .proto files
	outDir  string   // directory that receives the .pb.go files
	goPkg   string   // Go import path assigned by protoc-gen-go
	include []string // extra -I include directories
	files   []string // proto file names (bare)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "genproto:", err)
		os.Exit(1)
	}
}

func run() error {
	commonDir := "protobuf"

	units := []unit{{
		srcDir: commonDir,
		outDir: commonDir,
		goPkg:  modulePath + "/protobuf",
	}}

	entries, err := os.ReadDir("api")
	if err != nil {
		return fmt.Errorf("reading api: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join("api", e.Name(), "protobuf")
		if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
			continue
		}
		units = append(units, unit{
			srcDir:  dir,
			outDir:  filepath.Join("api", e.Name(), "protobuf"),
			goPkg:   modulePath + "/api/" + e.Name() + "/protobuf",
			include: []string{commonDir},
		})
	}

	// Collect the .proto files of every unit and build the M mappings keyed by
	// bare file name (the names are unique across the project).
	mappings := map[string]string{}
	for i := range units {
		names, err := protoFiles(units[i].srcDir)
		if err != nil {
			return err
		}
		units[i].files = names
		for _, n := range names {
			if prev, dup := mappings[n]; dup && prev != units[i].goPkg {
				return fmt.Errorf("proto file %s maps to both %s and %s", n, prev, units[i].goPkg)
			}
			mappings[n] = units[i].goPkg
		}
	}

	mapKeys := make([]string, 0, len(mappings))
	for n := range mappings {
		mapKeys = append(mapKeys, n)
	}
	sort.Strings(mapKeys)

	mOpts := make([]string, 0, len(mapKeys))
	for _, n := range mapKeys {
		mOpts = append(mOpts, "--go_opt=M"+n+"="+mappings[n])
	}

	for _, u := range units {
		if len(u.files) == 0 {
			continue
		}
		if err := os.MkdirAll(u.outDir, 0o755); err != nil {
			return fmt.Errorf("creating %s: %w", u.outDir, err)
		}

		args := []string{"-I", u.srcDir}
		for _, inc := range u.include {
			args = append(args, "-I", inc)
		}
		args = append(args, "--go_out="+u.outDir, "--go_opt=paths=source_relative")
		args = append(args, mOpts...)
		for _, f := range u.files {
			args = append(args, filepath.Join(u.srcDir, f))
		}

		cmd := exec.Command("protoc", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Printf("protoc %s -> %s\n", u.srcDir, u.outDir)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("protoc for %s: %w", u.srcDir, err)
		}
	}
	return nil
}

func protoFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".proto") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
