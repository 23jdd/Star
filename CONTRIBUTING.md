# 参与贡献

感谢为 Star 提交改进。请确保修改保持 Elm 更新模型、tcell Screen 抽象和现有公开
API 的一致性。

## 开发环境

```bash
git clone https://github.com/23jdd/Star.git
cd Star
go mod download
go test ./...
```

## 提交要求

1. 使用 `gofmt` 格式化所有 Go 文件。
2. 新行为必须附带测试；修复缺陷时应加入能够复现缺陷的回归测试。
3. 公开标识符应具有 GoDoc 注释。
4. 不要在 Widget 中直接向标准输出写内容，所有绘制都应通过 `tcell.Screen`。
5. `Update` 必须保持串行；异步工作放入 `Cmd`，并通过消息返回结果。
6. 修改公开 API 或用户可见行为时更新 README、对应文档和 CHANGELOG。

## 本地检查

```bash
go fmt ./...
go test -count=1 ./...
go vet ./...
go build ./...
```

在支持 CGO 的环境中还应运行：

```bash
go test -race ./...
```

## 提交信息

提交信息应简洁说明结果，例如：

```text
widgets: fix wide-character cursor placement
runtime: cancel pending command delivery on exit
```
