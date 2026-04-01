# Go Conversion Phase 0: Module Init

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Initialize the Go module, directory structure, and tooling in `go/`.

**Architecture:** Go module at `go/` with `internal/` for all packages. Module name matches the existing npm package path pattern.

**Tech Stack:** Go 1.23+, golangci-lint, Makefile

---

### Task 1: Initialize Go Module

**Files:**
- Create: `go/go.mod`
- Create: `go/go.sum` (generated)
- Create: `go/Makefile`

- [ ] **Step 1: Create go/ directory and initialize module**

```bash
mkdir -p /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go mod init github.com/anneschuth/claude-threads
```

Expected output: `go: creating new go.mod: module github.com/anneschuth/claude-threads`

- [ ] **Step 2: Add all required dependencies**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go get github.com/stretchr/testify@v1.10.0
go get github.com/spf13/cobra@v1.9.1
go get gopkg.in/yaml.v3@v3.0.1
go get github.com/gorilla/websocket@v1.5.3
go get github.com/charmbracelet/bubbletea@v1.3.4
go get github.com/charmbracelet/lipgloss@v1.1.0
go get github.com/charmbracelet/huh@v0.6.0
go get github.com/Masterminds/semver/v3@v3.3.1
go get github.com/sergi/go-diff@v1.3.1
go get github.com/go-playground/validator/v10@v10.26.0
go get golang.org/x/time@v0.11.0
```

Expected: Each dependency added to go.mod

- [ ] **Step 3: Create directory scaffold**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
mkdir -p cmd/claude-threads
mkdir -p internal/utils
mkdir -p internal/config
mkdir -p internal/platform/mattermost
mkdir -p internal/platform/slack
mkdir -p internal/session
mkdir -p internal/operations/executors
mkdir -p internal/claude
mkdir -p internal/mcp
mkdir -p internal/persistence
mkdir -p internal/git
mkdir -p internal/commands
mkdir -p internal/autoupdate
mkdir -p internal/ui/components
mkdir -p internal/ui/layouts
mkdir -p internal/ui/styles
mkdir -p internal/statusline
mkdir -p internal/messagehandler
mkdir -p internal/onboarding
mkdir -p internal/version
mkdir -p tests/integration/suites
mkdir -p tests/integration/fixtures
mkdir -p tests/unit
```

- [ ] **Step 4: Create placeholder main.go**

Create `go/cmd/claude-threads/main.go`:

```go
package main

import "fmt"

func main() {
	fmt.Println("claude-threads (Go)")
}
```

- [ ] **Step 5: Verify build**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go build ./...
```

Expected: No output (success)

- [ ] **Step 6: Create Makefile**

Create `go/Makefile`:

```makefile
.PHONY: build test lint integration clean

build:
	go build -o bin/claude-threads ./cmd/claude-threads

test:
	go test ./... -v

test-short:
	go test ./... -short -v

lint:
	golangci-lint run ./...

integration:
	go test ./tests/integration/... -v -timeout 10m

clean:
	rm -rf bin/

vet:
	go vet ./...
```

- [ ] **Step 7: Create .gitignore additions**

Append to `go/.gitignore`:

```
bin/
*.test
```

Create `go/.gitignore`:

```
bin/
*.test
```

- [ ] **Step 8: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/
git commit -m "chore: initialize Go module scaffold"
```

---

### Task 2: Verify Go Toolchain

**Files:**
- No new files

- [ ] **Step 1: Verify Go version**

```bash
go version
```

Expected: `go version go1.23.x` or newer

- [ ] **Step 2: Run go vet on scaffold**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go vet ./...
```

Expected: No output (success)

- [ ] **Step 3: Run empty test suite**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./...
```

Expected: `? github.com/anneschuth/claude-threads/cmd/claude-threads [no test files]`
