# 测试指南

## 测试 Model

Model.Update 是普通 Go 方法，可以直接构造消息并断言新模型和 Cmd。Cmd 的返回值也
可以直接执行检查，但不要在单元测试中使用过长的真实定时器。

## 测试完整 Program

使用 tcell 模拟屏幕，无需占用真实终端：

```go
screen := tcell.NewSimulationScreen("UTF-8")
screen.SetSize(80, 24)
program := star.NewProgram(star.WithScreen(screen))

done := make(chan error, 1)
go func() { done <- program.Run(model) }()

screen.PostEventWait(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))
if err := <-done; err != nil {
    t.Fatal(err)
}
```

Program 会负责模拟 Screen 的 Init 和 Fini，不应在测试中重复调用。

## 测试 Widget 绘制

```go
screen := tcell.NewSimulationScreen("UTF-8")
if err := screen.Init(); err != nil { t.Fatal(err) }
defer screen.Fini()
screen.SetSize(20, 4)

widget := widgets.NewText("你好")
widget.Measure(widgets.Tight(20, 4))
widget.Arrange(widgets.NewRect(0, 0, 20, 4))
widget.Render(screen)

main, combining, style, width := screen.GetContent(0, 0)
```

可以通过 `GetContent` 断言主 rune、组合 rune、Style 和字符宽度。

## 测试事件

直接调用 `Dispatch` 可以精确验证传播顺序、停止传播和默认行为。需要测试真实焦点与
点击合成时，创建 Router 并调用 `HandleKey` / `HandleMouse`。

## 推荐命令

```bash
go test -count=1 ./...
go test -cover ./...
go vet ./...
go build ./...
```

支持 CGO 时运行竞态检测：

```bash
go test -race ./...
```

仓库还提供约束、Canvas 和 TextArea 光标模糊测试：

```bash
go test -run '^$' -fuzz '^FuzzConstraintsNeverReturnNegativeSize$' ./widgets
go test -run '^$' -fuzz '^FuzzCanvasBounds$' ./widgets
go test -run '^$' -fuzz '^FuzzTextAreaCursorBounds$' ./widgets
```

## 测试隔离

- 不要让测试访问真实全局终端。
- 为每个 Program 测试创建独立 SimulationScreen。
- 所有运行中的 Program 都必须通过 Quit、Context 或测试清理逻辑结束。
- 对异步测试使用有界 timeout，避免失败时永久挂起。
- 不依赖 Batch 命令的完成顺序；需要顺序时使用 Sequence。
