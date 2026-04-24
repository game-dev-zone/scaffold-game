// cmd/scaffold-game 根據內嵌模板產生一個「獨立 repo」的新遊戲服務骨架。
//
// 外部協作者不需要 clone 整個 monorepo，只要：
//
//	go install github.com/club8/scaffold-game@latest
//	scaffold-game --game-id=niuniu --module=github.com/acme/club-game-niuniu --out-dir=./club-game-niuniu
//
// 或由 monorepo 開發者用 `make scaffold-game GAME_ID=...`。
//
// 產出內容：
//   - go.mod（pinned pkg-proto / pkg-game-framework tag 版本）
//   - cmd/game-<id>/main.go （呼叫 framework.Run 的 20 行骨架）
//   - internal/logic/logic.go （GameLogic stub，全部方法 TODO）
//   - internal/logic/logic_test.go
//   - Makefile / Dockerfile / README.md / deploy/env.example
//   - .github/workflows/ci.yml
//
// 產出後自動跑 `go mod tidy && go build ./...` 驗證。
package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed all:templates
var tmplFS embed.FS

type params struct {
	GameID           string
	ModulePath       string
	ProtoVersion     string
	FrameworkVersion string
	GoVersion        string
}

func main() {
	var (
		gameID           = flag.String("game-id", "", "lowercase id, matches ^[a-z][a-z0-9_]{1,15}$ (required)")
		modulePath       = flag.String("module", "", "Go module path for the new repo, e.g. github.com/acme/club-game-niuniu (required)")
		outDir           = flag.String("out-dir", "", "output directory; must not exist (required)")
		protoVersion     = flag.String("proto-version", "v1.0.0", "pkg-proto tag to pin")
		frameworkVersion = flag.String("framework-version", "v0.1.0", "pkg-game-framework tag to pin")
		goVersion        = flag.String("go-version", "1.25.0", "Go version for go.mod")
		skipBuild        = flag.Bool("skip-build", false, "skip go mod tidy + go build after scaffold (useful when versions not yet published)")
	)
	flag.Parse()

	if err := validate(*gameID, *modulePath, *outDir); err != nil {
		fmt.Fprintln(os.Stderr, "scaffold-game:", err)
		flag.Usage()
		os.Exit(2)
	}

	p := params{
		GameID:           *gameID,
		ModulePath:       *modulePath,
		ProtoVersion:     *protoVersion,
		FrameworkVersion: *frameworkVersion,
		GoVersion:        *goVersion,
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatal(err)
	}
	if err := render(p, *outDir); err != nil {
		fatal(err)
	}

	if *skipBuild {
		fmt.Fprintf(os.Stderr, "\n✓ scaffolded %s at %s (skipped build)\n", *gameID, *outDir)
		return
	}

	fmt.Fprintln(os.Stderr, ">> go mod tidy ...")
	if err := run(*outDir, "go", "mod", "tidy"); err != nil {
		fatal(fmt.Errorf("go mod tidy failed: %w", err))
	}
	fmt.Fprintln(os.Stderr, ">> go build ./... ...")
	if err := run(*outDir, "go", "build", "./..."); err != nil {
		fatal(fmt.Errorf("go build failed: %w", err))
	}
	fmt.Fprintf(os.Stderr, "\n✓ scaffolded %s at %s — ready to code\n", *gameID, *outDir)
}

var gameIDRe = regexp.MustCompile(`^[a-z][a-z0-9_]{1,15}$`)

func validate(gameID, modulePath, outDir string) error {
	if gameID == "" {
		return errors.New("--game-id required")
	}
	if !gameIDRe.MatchString(gameID) {
		return fmt.Errorf("--game-id %q must match %s", gameID, gameIDRe.String())
	}
	if modulePath == "" {
		return errors.New("--module required")
	}
	if outDir == "" {
		return errors.New("--out-dir required")
	}
	if _, err := os.Stat(outDir); err == nil {
		return fmt.Errorf("--out-dir %s already exists", outDir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func render(p params, outDir string) error {
	return fs.WalkDir(tmplFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// 去掉 "templates/" 前綴；".tmpl" 尾綴；把 __GAME_ID__ 替換成實際 id。
		rel := strings.TrimPrefix(path, "templates/")
		rel = strings.TrimSuffix(rel, ".tmpl")
		rel = strings.ReplaceAll(rel, "__GAME_ID__", p.GameID)
		dst := filepath.Join(outDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}

		raw, err := tmplFS.ReadFile(path)
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".tmpl") {
			t, err := template.New(rel).Parse(string(raw))
			if err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}
			f, err := os.Create(dst)
			if err != nil {
				return err
			}
			defer f.Close()
			if err := t.Execute(f, p); err != nil {
				return fmt.Errorf("exec %s: %w", path, err)
			}
			return nil
		}
		return os.WriteFile(dst, raw, 0o644)
	})
}

func run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "scaffold-game:", err)
	os.Exit(1)
}
