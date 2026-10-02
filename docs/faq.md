# 常见问题

## Star 与 Bubble Tea 有什么区别？

两者都使用 Elm 风格状态更新。Star 的 View 返回可测量、可布局、可命中测试的 Widget
树，并直接使用 tcell Screen；Bubble Tea 的核心 View 通常产生字符串。

## 为什么程序启动后先收到 WindowSizeMsg？

首帧布局需要真实终端尺寸。Star 在 Init 命令启动前发送尺寸消息，让模型能够立即建立
响应式布局状态。

## 为什么 Input 的内容每次按键后都会丢失？

Input 是状态型 Widget。如果在每次 View 中重新 `NewInput`，编辑状态会被重置。应把
Input 指针保存在 Model 中，并让 View 返回同一个实例。

## 为什么收不到鼠标、Paste 或焦点事件？

这些事件默认关闭，需要配置：

```go
star.NewProgram(
    star.WithMouse(true),
    star.WithPaste(true),
    star.WithFocus(true),
)
```

Widget 交互还需要在 Update 中调用 `star.RouteInput`。

## Button 为什么没有自动修改 Model？

Button 不知道应用消息类型。可以在模型构造阶段通过 `On(Click, ...)` 注册回调；回调由
Update 内的 RouteInput 同步触发，因此可以安全更新模型字段。需要异步工作时，让回调
设置状态，并由 Update 返回 Cmd，或向 Program 发送业务消息。

## 如何创建周期任务？

`Tick` 只触发一次。在处理 tick 消息时返回下一次 Tick。这样退出或业务状态变化时可以
停止调度。

## Batch 与 Sequence 应该选哪个？

- 各任务互不依赖、希望尽快完成：Batch。
- 后一个任务必须等前一个完成、结果必须有序：Sequence。

终端输入仍可能穿插在 Sequence 的结果消息之间。

## Cmd 可以直接修改 Model 吗？

不可以。Cmd 在其他 goroutine 中运行，直接修改 Model 会造成数据竞争。Cmd 只返回消息，
由 Update 串行应用结果。

## 如何在退出时取消正在运行的 Cmd？

使用 `star.WithContext` 创建命令。传入的 Context 在 Program 因 Quit、错误或外部
Context 取消而退出时都会取消。命令调用的下游 API 也应使用这个 Context。

## 如何记录日志而不破坏终端画面？

不要在运行期间使用 fmt.Println。写入文件、结构化日志目标，或者把日志作为消息发送给
模型并显示在 ScrollView 中。

## 如何退出？

在 Update 中返回 `star.QuitCmd()`，或从外部调用 `program.Send(star.Quit())`。使用
RunContext 时也可以取消 Context。

## 如何处理 panic？

Run/RunContext 会恢复终端并返回 `*star.PanicError`。应用入口应记录该错误并以非零状态
退出。不要依赖 panic 执行业务控制流。

## 支持图片协议吗？

`Image` 组件使用真彩色半块字符显示普通 Go `image.Image`，不发送 Sixel、Kitty
Graphics 或 iTerm 图片协议。它能够跨终端工作，但清晰度受终端字符网格限制；需要
像素级图片协议时，可通过自定义 Widget 集成目标终端能力。

## 为什么第三方容器的事件不能自动冒泡？

普通自定义叶子 Widget 完全支持。父节点绑定目前由内置容器维护，第三方容器应优先组合
Stack、Box 或 Overlay。这个限制不会影响自定义绘制组件。
