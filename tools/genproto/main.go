// Command genproto 为项目中每个 .proto 文件生成 Go protobuf 绑定。
// 每个源目录会成为独立的 Go 包：
//
//	protobuf/*.proto             -> ./protobuf
//	api/<name>/protobuf/*.proto  -> ./api/<name>/protobuf
//
// 这些 .proto 文件没有声明 go_package 选项，并且互相以裸文件名 import，
// 因此需要 -I include 路径以及完整的 M 映射。
//
// 用法:
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
	srcDir  string   // 存放 .proto 文件的目录
	outDir  string   // 接收 .pb.go 文件的目录
	goPkg   string   // protoc-gen-go 分配的 Go import path
	include []string // 额外的 -I include 目录
	files   []string // proto 文件名（裸名）
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

	// 收集每个单元的 .proto 文件，并以裸文件名为键构建 M 映射
	// （项目内文件名唯一）。
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
