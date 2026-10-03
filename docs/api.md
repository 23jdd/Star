# Star API 手册

本文覆盖 `star` 和 `widgets` 包的公开 API。可执行示例位于 `examples/`。

## star 包

### 核心类型

```go
type Message = any
type Cmd func() Message

type Model interface {
    Init() Cmd
    Update(Message) (Model, Cmd)
    View() widgets.Widget
}
```

- `Message` 可以是任意 Go 值。建议使用小型结构体类型区分消息。
- `Update` 只在程序事件循环中串行调用，不会并发调用。
- `Cmd` 在独立 goroutine 中执行，返回值会重新进入消息队列。返回 `nil` 表示不发送消息。
- `View` 返回当前模型对应的 Widget 树。状态型 Widget 应保存在模型中，不应每帧重建。

### Program

```go
program := star.NewProgram(options...)
err := program.Run(model)
err := program.RunContext(ctx, model)
```

`Run` 使用 `context.Background()`；`RunContext` 在 Context 取消时恢复终端并返回
`ctx.Err()`。同一个 Program 不能并发运行，否则返回 `star.ErrRunning`。

Program 始终负责初始化和终止 Screen。即使模型、命令或渲染发生 panic，也会先调用
`Screen.Fini`，随后返回 `*star.PanicError`。

外部消息注入：

| 方法 | 行为 |
|---|---|
| `Send(msg)` | 等待消息队列可写；未运行时返回 `ErrNotRunning` |
| `SendContext(ctx, msg)` | 队列可写、程序结束或 Context 取消时返回 |
| `TrySend(msg)` | 从不阻塞；写入成功返回 `true` |

### Program Option

| Option | 默认值 | 说明 |
|---|---:|---|
| `WithQueueSize(n)` | 256 | 内部消息缓冲容量；小于 1 的值被忽略 |
| `WithMouse(bool)` | false | 是否让 tcell 上报鼠标事件 |
| `WithPaste(bool)` | false | 是否上报括号粘贴开始/结束事件 |
| `WithFocus(bool)` | false | 是否上报终端焦点变化 |
| `WithScreen(screen)` | 自动创建 | 注入 Screen；Program 仍负责 `Init` 和 `Fini` |

### 内置消息

| 类型 | 字段 | 说明 |
|---|---|---|
| `KeyMsg` | `Key`、`Rune`、`Modifiers` | 键盘输入；`String()` 返回 tcell 键名 |
| `MouseMsg` | `X`、`Y`、`Buttons`、`Modifiers` | 鼠标状态；`Position()` 返回坐标 |
| `WindowSizeMsg` | `Width`、`Height` | 启动时首先发送，resize 时再次发送 |
| `PasteMsg` | `Start` | 括号粘贴开始或结束 |
| `FocusMsg` | `Focused` | 终端获得或失去焦点 |
| `RawEventMsg` | `Event` | Star 尚未专门转换的 tcell 事件 |

### 命令

| 函数 | 说明 |
|---|---|
| `Quit()` | 直接产生退出消息，适合外部 `Send` |
| `QuitCmd()` | 返回产生退出消息的命令 |
| `After(d, msg)` | 等待指定时长后返回消息 |
| `Tick(d, fn)` | 等待后以实际触发时间调用 `fn`，只触发一次 |
| `Batch(cmds...)` | 并发启动所有非 nil 命令，结果按完成顺序到达 |
| `Sequence(cmds...)` | 顺序执行所有非 nil 命令，结果保持顺序 |
| `WithContext(fn)` | 让命令接收 Program 生命周期 Context，退出时自动取消 |

需要周期任务时，应在处理 Tick 消息后返回下一次 `Tick`，这样模型可以决定是否继续：

```go
type tickMsg time.Time

func tick() star.Cmd {
    return star.Tick(time.Second, func(t time.Time) star.Message {
        return tickMsg(t)
    })
}

func (m model) Update(msg star.Message) (star.Model, star.Cmd) {
    if _, ok := msg.(tickMsg); ok {
        m.count++
        return m, tick()
    }
    return m, nil
}
```

### RouteInput

`RouteInput(router, message)` 把 `KeyMsg` 和 `MouseMsg` 转交给 Widget Router；其他
消息返回 `false`。启用鼠标时还需要给 Program 添加 `WithMouse(true)`。

## widgets 包

### Widget 接口

```go
type Widget interface {
    Render(screen tcell.Screen)
    Measure(constraints Constraints) Size
    Arrange(bounds Rect)
    HandlerEvent(event *Event)
}
```

渲染顺序固定为 `Measure → Arrange → Render`。`Arrange` 接收绝对屏幕坐标；绘制和
命中测试必须使用同一份边界。

### 几何与约束

- `Size{W,H}`：期望或测量出的尺寸。
- `Rect{X,Y,W,H}`：绝对布局区域；提供 `Contains` 和 `Empty`。
- `Constraints{MinW,MinH,MaxW,MaxH}`：布局约束。
- `Tight(w,h)`：最小值和最大值相同。
- `Loose(w,h)`：最小值为 0。
- `Constraints.Constrain(size)`：把尺寸限制到合法范围。

所有负尺寸都会安全地归一化为 0。

### WidgetBase

自定义 Widget 可以嵌入 `WidgetBase` 获得：

- `Bounds` / `SetBounds`
- `Parent` / `SetParent`
- `Hidden` / `SetHidden`
- `Parent`
- `On(kind, capture, handler)`
- `EventBase`
- `DefaultAction`

`DefaultAction` 在事件传播结束后执行；调用 `PreventDefault` 可以取消它。

### 布局组件

#### Stack、VStack、HStack

```go
stack.Add(child, basis, flex)
```

- `basis > 0`：主轴固定尺寸。
- `basis <= 0`：按 `flex` 比例分配剩余空间；非正 flex 视为 1。
- `Gap`：相邻非 nil 子组件之间的间隔。
- `Background`：Stack 覆盖范围的背景样式。

#### Box

单子节点装饰容器。支持 `Padding`、`Border`、`Title`、`Style` 和
`BorderStyle`。内置 `RoundedBorder`、`SquareBorder` 和 `DoubleBorder`。

默认情况下 Child 填满扣除边框和 Padding 后的内部区域。设置 `FitContent=true` 后，
Box 会使用 Child 的 `Measure` 尺寸，并通过 `AlignLeft` / `AlignCenter` /
`AlignRight` 及 `AlignTop` / `AlignMiddle` / `AlignBottom` 定位：

```go
box := widgets.NewBox(button)
box.Border = &widgets.RoundedBorder
box.Padding = widgets.UniformInsets(1)
box.FitContent = true
box.Align = widgets.AlignCenter
box.VerticalAlign = widgets.AlignMiddle
```

#### Overlay

所有子组件获得相同边界，后加入的组件绘制在上层，命中测试也优先选择后加入的组件。

#### Spacer

请求最小宽高的空组件，常用于布局留白。

### 显示组件

#### Text

字段：`Content`、`Style`、`Wrap`、`Align`。对齐方式为 `AlignLeft`、
`AlignCenter`、`AlignRight`。换行和裁剪按 Unicode 字素宽度计算。

#### Progress

`Value` 的有效范围为 0 到 1，渲染时自动钳制。`ShowPercentage` 控制百分比文本，
`Style` 和 `FilledStyle` 分别控制未完成和已完成区域。

#### Canvas

保留模式的终端 Cell 缓冲区：

- `Resize(w,h)` 调整缓冲区并保留重叠区域。
- `SetCell(x,y,cell)` 设置单元格，越界返回 `false`。
- `Cell(x,y)` 返回单元格副本和是否存在。
- `Cell` 支持主 rune、组合 rune 和独立 tcell Style。

#### Image

`NewImage(source image.Image)` 使用 `▀`、`▄` 半块字符渲染真彩色图片，因此一个终端
单元格能够表达上下两个图片像素，并且不依赖 Kitty、Sixel 或 iTerm 私有协议。
也可以直接从文件创建：

```go
picture, err := widgets.NewImageFromPath("assets/logo.png")
if err != nil {
    return err
}
picture.Fit = widgets.ImageFitContain
```

- `Fit` 支持 `ImageFitContain`、`ImageFitCover`、`ImageFitFill` 和 `ImageFitNone`。
- `Align` 使用 `AlignLeft`、`AlignCenter`、`AlignRight` 控制水平位置。
- `VerticalAlign` 使用 `ImageAlignTop`、`ImageAlignMiddle`、`ImageAlignBottom`。
- 全透明像素不会覆盖已有单元格；半透明像素会与父组件背景合成。
- `FallbackBackground` 用于终端默认背景无法解析时的 alpha 合成；主题构造器
  `theme.NewImage` 会自动使用主题的 Surface 背景。
- `NewImageFromPath` 和 `theme.NewImageFromPath` 支持 PNG、JPEG 和 GIF（GIF 第一帧）。
- `SetImage` 或 `SetImagePath` 可以在运行期间替换图片；`SetImage(nil)` 会停止绘制，
  `SetImagePath` 加载失败时会保留原图片。

### 交互组件

#### Button

支持普通、焦点和禁用样式。实现 `Focusable`，Enter 或空格会由 Router 合成
`Click`。通过 `On(Click, false, handler)` 处理激活。`Align` 使用 `AlignLeft`、
`AlignCenter` 或 `AlignRight` 控制按钮边界内的文字位置，`NewButton` 默认居中；
`Padding` 控制左右内边距。

#### Input

单行输入框，支持：

- Rune 插入
- 左右方向键、Home、End、Ctrl+A、Ctrl+E
- Backspace、Delete、Ctrl+D
- Enter 产生 `Submitted`
- 内容变化产生 `Changed`

`Cursor` 是 rune 索引。使用内置编辑状态时必须在多次 `View` 之间保留同一个 Input
实例。`SetValue` 会同步钳制光标位置。

#### TextArea

保留状态的多行编辑器，支持 Rune 插入、换行、跨行 Backspace/Delete、方向键、
Home/End、PageUp/PageDown、Ctrl+A/Ctrl+E，以及纵向和横向滚动。内容变化产生
`Changed`。`CursorRow` 和 `CursorCol` 分别是行索引和 rune 索引。

#### List

`Items` 保存字符串项目，`Selected` 和 `Offset` 保存选择及滚动位置。支持方向键、
Home、End、PageUp、PageDown、j/k、鼠标滚轮和点击。选择变化产生 `Changed`，
Enter 或点击产生 `Submitted`。

#### ScrollView

支持方向键、PageUp/PageDown、Home/End 和鼠标滚轮。`Wrap` 控制换行，
`FollowEnd` 用于日志追尾，`ShowScrollbar` 显示滚动条。`SetContent` 和
`ScrollTo` 会自动校正 Offset。

#### Table

`TableColumn` 使用 `Title` 和 `Width` 描述列；`Width <= 0` 的列平均分配剩余宽度。
Table 固定显示表头，支持方向键、Home/End、PageUp/PageDown、滚轮、点击和 Enter。
选择变化产生 `Changed`，激活行产生 `Submitted`。

#### Tree

`TreeNode` 保存 Label、Children 和 Expanded。Tree 使用 Left/Right 折叠或展开，Left
也可以返回父节点；Up/Down、Home/End、滚轮和点击用于选择。`SelectedNode` 返回当前
可见节点。

#### Tabs

每个 `Tab` 包含 Title 和 Content。Tabs 只测量、布局和暴露当前活动 Content，支持
Left/Right、Home/End 和点击切换。活动页变化产生 `Changed`。

#### Modal

Modal 在 Backdrop 上居中显示单个 Child，并支持标题、边框、padding、固定或内容驱动
尺寸。将 Modal 临时设为 Router Root 可以限制命中测试与焦点范围。Esc 或点击面板外部
会产生 `Dismissed`，可分别通过配置关闭。

#### Form

`FormField` 关联 Label、Input 和 `Validator`。Form 自动排列字段和 Submit Button，
`Validate` 将错误写入 `Errors`；只有全部字段有效时才产生 `Submitted`。`Values` 按
显示顺序返回字段值。

### Theme

`DefaultTheme()` 返回语义化默认样式，包括 Text、Surface、Border、Primary、
Focused、Selected、Disabled、Scrollbar、Progress、Error 和 Backdrop。Theme 为基础
组件和高级组件提供一致样式的便捷构造器。

### Router 与事件

`Router` 保存 Root、Focused、PointerCaptured、PressedTarget 和 DragThreshold。

- `SetRoot`：替换根节点；旧焦点不在新树中时自动清除。
- `SetFocus`：切换焦点并发送 `Blurred` / `Focused`。
- `FocusNext` / `FocusPrevious`：按深度优先顺序循环移动焦点。
- `HandleKey`：发送 KeyDown，处理 Tab、Shift-Tab 和键盘激活。
- `HandleMouse`：发送移动、按下、释放、滚动事件，并合成 Click。

事件种类：`KeyDown`、`PointerDown`、`PointerMove`、`PointerUp`、`Click`、
`PointerScroll`、`Focused`、`Blurred`、`Changed`、`Submitted`、`Dismissed`。

传播控制：

| 方法 | 效果 |
|---|---|
| `StopPropagation` | 当前组件的其余监听器继续，停止后续节点传播 |
| `StopImmediatePropagation` | 同时停止当前组件剩余监听器和后续传播 |
| `PreventDefault` | 保留传播，但取消目标组件默认行为 |

`HitTest(root,x,y)` 从后向前检查子节点，因此后绘制的组件优先。
