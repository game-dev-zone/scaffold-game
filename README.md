# scaffold-game

產生 Club No.8 新遊戲**獨立 repo**骨架的 Go CLI。外部協作者用這個不必 clone 主 monorepo，一行裝好、一行產 repo。

## 安裝

```bash
go install github.com/game-dev-zone/scaffold-game@latest
```

需要 Go 1.22+。

## 使用

```bash
scaffold-game \
  --game-id=niuniu \
  --module=github.com/acme/club-game-niuniu \
  --out-dir=./club-game-niuniu
```

必填：

| Flag | 說明 |
| ---- | ---- |
| `--game-id` | 小寫遊戲代號，正則 `^[a-z][a-z0-9_]{1,15}$`（如 `niuniu` / `fish`） |
| `--module` | 新 repo 的 Go module path（如 `github.com/acme/club-game-niuniu`） |
| `--out-dir` | 輸出資料夾，**必須不存在** |

選填：

| Flag | 預設 | 說明 |
| ---- | ---- | ---- |
| `--proto-version` | `v1.0.1` | 要 pin 的 `pkg-proto` tag |
| `--framework-version` | `v0.1.1` | 要 pin 的 `pkg-game-framework` tag |
| `--go-version` | `1.25.0` | 產出 `go.mod` 的 Go 版本 |
| `--skip-build` | — | 跳過產出後的自動 `go mod tidy + go build` |

## 產出結構

```
<out-dir>/
├── go.mod                       pinned pkg-proto + pkg-game-framework tag
├── README.md                    快速開始指引
├── Makefile                     tidy / lint / test / build / run / docker-build
├── Dockerfile                   alpine + static binary
├── .gitignore
├── .github/workflows/ci.yml     go vet + test + build
├── cmd/game-<id>/main.go        ~20 行：呼叫 framework.Run
├── internal/logic/
│   ├── logic.go                 GameLogic 介面 TODO stub（你要填這裡）
│   └── logic_test.go            table-test 骨架，用 fake Tx/Record 驅動
└── deploy/env.example
```

產出後自動跑 `go mod tidy && go build ./...`；全綠才成功。

## 之後怎麼開工

1. 進 `<out-dir>` → `git init -b main`
2. 在 GitHub 建你的新 repo（`gh repo create ...`）
3. `git push -u origin main`
4. 編輯 `internal/logic/logic.go`，實作 5 個 `GameLogic` hook
5. 本機跑 `make run`（需主 monorepo 的 infra 與 tx/record 服務起來）

詳細遊戲開發指南見主 monorepo 的 `docs/game-developer-guide.md`（scaffold 產出的 README 有連結）。

## 疑問

- 為什麼要獨立 repo：見主 monorepo 的 `CLAUDE.md` 第 7 條開發原則（降低耦合；每個遊戲自己的 CI / version / 部署節奏）。
- 要直接在 monorepo 內開新資料夾可不可以：不可以，會違反「新遊戲是獨立 repo」原則。`game-ddz` 是唯一例外，作為框架的 in-tree 參考實作。
- 如何升版依賴：`go get github.com/game-dev-zone/pkg-proto@vX.Y.Z && go mod tidy`。
