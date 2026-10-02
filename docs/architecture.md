# 架构与生命周期

## 设计目标

Star 将应用状态、异步副作用和终端绘制分开：

```text
终端事件 ─┐
Cmd 结果 ─┼─> 单一消息队列 ─> Update ─> 新 Model ─> View ─> Screen
外部 Send ┘                         └─> Cmd ───────────────┘
```

只有事件循环调用 `Update`，所以模型更新天然串行。`Cmd` 可以并发执行，但只能返回
消息，不能直接修改模型。

## 启动顺序

1. 检查 Model、Context 和 Program 运行状态。
2. 创建或使用注入的 tcell Screen，并调用 `Init`。
3. 根据 Option 启用鼠标、粘贴和焦点事件。
4. 启动独立的 Screen 事件读取循环。
5. 把当前 Screen 尺寸作为第一条 `WindowSizeMsg` 交给 `Update`。
6. 执行 `View → Measure → Arrange → Render`，绘制首帧。
7. 启动首次 `Update` 和 `Init` 返回的命令。
8. 进入消息循环。

先发送尺寸消息可以保证首帧已经拥有真实终端尺寸。

## 消息循环

消息有三个来源：tcell 事件、Cmd 结果和 Program.Send。任何来源最终都会在同一个
goroutine 中调用 `Update`。每次有效消息处理后都会重新构建或取得 View，随后清屏、
布局并调用 tcell `Show`。

窗口尺寸变化时 Star 会先调用 Screen.Sync，再发送 `WindowSizeMsg`。

## 命令模型

普通 Cmd 在独立 goroutine 中执行。`Batch` 同时启动多个命令；`Sequence` 在一个
goroutine 中按顺序执行。命令完成后向内部队列发送结果。程序退出后，普通 Cmd 可以
继续完成自身工作，但结果投递会被丢弃，不会永久阻塞 goroutine。`WithContext` 命令
会收到取消信号，因此更适合网络、数据库和长时间任务。

长时间运行或访问外部资源的命令应使用 `WithContext` 接受 Program 生命周期 Context：

```go
func fetch() star.Cmd {
    return star.WithContext(func(runCtx context.Context) star.Message {
        result, err := client.Load(runCtx)
        return loadedMsg{Result: result, Err: err}
    })
}
```

## 退出和错误

以下情况会结束运行：

- 收到 `Quit()` 或 `QuitCmd()` 的结果
- `RunContext` 的 Context 被取消
- Screen 返回 EventError 或事件流关闭
- Model 返回 nil
- Model、View、Widget 或 Cmd 发生 panic

退出时 Screen.Fini 一定先于错误返回执行。Cmd panic 被封装为 `PanicError` 并通过
事件循环返回；同步 panic 同样会在恢复终端后转换为 `PanicError`。

## 布局

布局分为三步：

1. `Measure`：Widget 根据 Constraints 返回期望 Size。
2. `Arrange`：父组件把绝对 Rect 分配给子组件。
3. `Render`：组件只在自己的 Rect 中绘制。

Program 会以终端完整尺寸对根 Widget 调用 Tight constraints，并把根节点安排到
`Rect{0,0,width,height}`。

## Widget 事件树

WidgetBase 保存父节点，容器通过 `EventChildren` 暴露子节点。一次 Dispatch 会先固定
目标到根节点的路径，因此监听器在处理期间修改组件树不会改变本次传播路线。

```text
Root capture -> Panel capture -> Target -> Panel bubble -> Root bubble
```

默认行为只属于目标 Widget，并在传播结束后执行。`StopPropagation` 不会隐式取消
默认行为；必须显式调用 `PreventDefault`。

## 并发约束

- `Update`、`View` 和 Widget 渲染在事件循环内串行执行。
- Cmd 会并发运行，不能读写正在使用的可变模型或 Widget。
- 外部 goroutine 使用 Send/SendContext/TrySend 传递数据。
- Event handler 由 RouteInput 在 Update 调用链中执行，可以安全修改模型内状态。
- 不要从 Event handler 或 Update 中执行阻塞 I/O；把工作封装成 Cmd。
