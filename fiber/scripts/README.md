# scripts - 构建与部署脚本

| 脚本 | 说明 |
|------|------|
| `build.sh` | 构建二进制，输出到 `dist/fiber-app`，支持 `GOOS`/`GOARCH` 交叉编译 |
| `run.sh` | 本地开发运行 `go run .` |
| `deploy.sh` | 构建后复制到 `deploy/`（含 config、views） |
| `clean.sh` | 清理 `dist/` |

## 用法

```bash
# 构建（在 fiber 目录下）
./scripts/build.sh
./scripts/build.sh my-app   # 自定义输出名

# 交叉编译 Linux
GOOS=linux GOARCH=amd64 ./scripts/build.sh

# 部署
./scripts/deploy.sh           # 部署到 ./deploy
./scripts/deploy.sh /opt/app  # 部署到指定目录
```

## 根目录 Makefile

在 go_utils 根目录执行 `make help` 查看所有命令。
