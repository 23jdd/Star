# Star 事件系统

本文说明当前实现中的 Widget 事件、焦点、鼠标捕获和 Click 合成。完整 API 表参见
[`docs/api.md`](docs/api.md)。

## 两层输入模型

Star 首先把 tcell 输入转换为 Elm 消息：

- `star.KeyMsg`
- `star.MouseMsg`
- `star.WindowSizeMsg`
- `star.PasteMsg`
- `star.FocusMsg`

应用可以直接处理这些消息，也可以调用 `star.RouteInput(router, message)`，把键盘和
鼠标消息继续转换为 Widget 事件。

```go
func (m *model) Update(message star.Message) (star.Model, star.Cmd) {
    if key, ok := message.(star.KeyMsg); ok && key.Key == tcell.KeyEscape {
        return m, star.QuitCmd()
    }
    star.RouteInput(m.router, message)
    return m, nil
}
```

需要鼠标输入时必须启用：

```go
program := star.NewProgram(star.WithMouse(true))
```

## 传播阶段

`widgets.Dispatch` 先固定目标到根节点的路径，再按三个阶段传播同一个 `*Event`：

```text
Capture: Root -> Panel
Target:  Button
Bubble:  Panel -> Root
Default: Button.DefaultAction
```

注册监听器：

```go
panel.On(widgets.Click, true, func(event *widgets.Event) {
    // Capture：在目标之前观察子节点事件。
})

button.On(widgets.Click, false, func(event *widgets.Event) {
    // Target 或 Bubble。
})
```

事件中的 `Target` 始终是最初目标，`CurrentTarget` 是当前执行监听器的 Widget，
`Phase` 表示传播阶段。

## 传播控制

| 方法 | 行为 |
|---|---|
| `StopPropagation()` | 当前节点其余监听器继续，不再访问后续节点 |
| `StopImmediatePropagation()` | 当前节点剩余监听器和后续节点都停止 |
| `PreventDefault()` | 传播继续，但不执行目标的 `DefaultAction` |

停止传播不会自动取消默认行为。

## 命中测试

`widgets.HitTest(root,x,y)` 使用 Arrange 保存的绝对 Bounds。隐藏、空 Bounds 或坐标不在
范围内的 Widget 不会命中。容器的后一个子节点被视为上层，因此优先命中。

自定义容器需要实现 `EventChildren() []widgets.Widget`，并通过
`child.EventBase().SetParent(container)` 建立事件父链。

## 焦点

实现 `widgets.Focusable` 的组件会加入 Router 的深度优先焦点顺序。内置 Button、Input、
List、ScrollView、TextArea、Table、Tree、Tabs 和 Form 内部字段都支持焦点。

- Tab：下一个焦点
- Shift-Tab / Backtab：上一个焦点
- 鼠标按下：聚焦目标或最近的可聚焦祖先
- Enter / Space：对支持键盘激活的组件合成 Click

焦点切换会分别发送 `Blurred` 和 `Focused`。

## 鼠标捕获与 Click

鼠标左键按下时，Router 保存 `PressedTarget` 并设置 `PointerCaptured`。捕获期间即使
指针移出组件，PointerMove 和 PointerUp 仍发送给捕获目标。

释放时同时满足以下条件才会合成 Click：

1. 释放位置仍命中最初按下的目标；
2. 移动距离没有超过 `Router.DragThreshold`。

滚轮会生成 `PointerScroll`，方向保存在 `DeltaY`。

## 默认行为

Input、List 和 ScrollView 使用 `WidgetBase.DefaultAction` 实现内置交互。应用可以在
监听器中调用 `PreventDefault` 替换行为。例如禁止某个字符输入：

```go
input.On(widgets.KeyDown, false, func(event *widgets.Event) {
    if event.Rune == '/' {
        event.PreventDefault()
    }
})
```

事件回调在 `Update -> RouteInput` 调用链中同步执行，可以修改模型持有的 Widget 状态；
不要在回调中执行阻塞 I/O。
