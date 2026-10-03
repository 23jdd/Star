# Star 入门到生产实践教程

本教程通过一个小型服务控制台介绍 Star 的完整开发流程。完成后，你将掌握：

- Elm 风格的 `Model → Update → View` 数据流；
- 固定尺寸、弹性尺寸和嵌套布局；
- 键盘、鼠标、焦点及组件事件；
- 定时器与可取消异步任务；
- 表格、表单、弹窗和图片组件；
- 使用模拟终端编写自动化测试；
- 将 TUI 应用整理为可长期维护的生产代码。

如果只需要查字段和函数签名，请直接阅读 [API 手册](api.md)。本教程更关注组件如何
协作，以及为什么要这样组织应用。

## 1. 创建项目

初始化 Go 模块并安装 Star：

```bash
mkdir star-tutorial
cd star-tutorial
go mod init example.com/star-tutorial
go get github.com/23jdd/Star
```

创建 `main.go`：

```go
package main

import (
    "fmt"

    star "github.com/23jdd/Star"
    "github.com/23jdd/Star/widgets"
    "github.com/gdamore/tcell/v2"
)

type model struct {
    count int
}

func (m model) Init() star.Cmd { return nil }

func (m model) Update(message star.Message) (star.Model, star.Cmd) {
    if key, ok := message.(star.KeyMsg); ok {
        switch {
        case key.Rune == '+':
            m.count++
        case key.Rune == '-':
            m.count--
        case key.Rune == 'q' || key.Key == tcell.KeyEscape:
            return m, star.QuitCmd()
        }
    }
    return m, nil
}

func (m model) View() widgets.Widget {
    return widgets.NewVStack(
        widgets.NewText("Star 计数器"),
        widgets.NewText(fmt.Sprintf("当前值：%d", m.count)),
        widgets.NewText("按 +/- 修改，按 q 或 Esc 退出"),
    )
}

func main() {
    if err := star.NewProgram().Run(model{}); err != nil {
        panic(err)
    }
}
```

运行：

```bash
go run .
```

这个程序已经包含 Elm 架构的三个部分：

1. `Model` 保存唯一可信的应用状态。
2. `Update` 串行处理消息，并返回新 Model 与可选的 `Cmd`。
3. `View` 根据当前状态返回 Widget 树。

Star 会在启动时发送一条 `WindowSizeMsg`，之后每次处理有效消息都会重新执行布局和
绘制。终端初始化、恢复和异常清理由 `Program` 负责。

## 2. 理解 Cmd：副作用返回消息

`Update` 不应该等待网络、磁盘或长时间计算。耗时工作放进 `Cmd`，完成后把结果作为
消息送回 `Update`：

```go
type loadedMsg struct {
    value string
    err   error
}

func loadService() star.Cmd {
    return func() star.Message {
        value, err := readServiceFromDisk()
        return loadedMsg{value: value, err: err}
    }
}
```

处理结果：

```go
case loadedMsg:
    if message.err != nil {
        m.status = "加载失败：" + message.err.Error()
    } else {
        m.status = "已加载：" + message.value
    }
```

多个任务可以使用 `star.Batch` 并发执行，或使用 `star.Sequence` 保证命令启动和结果
顺序。需要响应程序退出的任务应使用 `star.WithContext`：

```go
func watchService() star.Cmd {
    return star.WithContext(func(ctx context.Context) star.Message {
        select {
        case result := <-waitForService():
            return result
        case <-ctx.Done():
            return nil
        }
    })
}
```

不要从 Cmd goroutine 直接修改 Model 或 Widget。Cmd 只返回消息，真正的状态变化仍在
`Update` 中发生。

## 3. 保留有状态 Widget

简单的 `Text` 可以在每次 `View` 中重建，但输入框、列表、滚动区、表格和 Router
包含光标、选择或滚动状态，应保存在 Model 中：

```go
type dashboard struct {
    theme    widgets.Theme
    input    *widgets.Input
    services *widgets.List
    progress *widgets.Progress
    logs     *widgets.ScrollView
    deploy   *widgets.Button
    status   *widgets.Text
    root     widgets.Widget
    router   *widgets.Router
}
```

在构造函数中一次性创建组件树：

```go
func newDashboard() *dashboard {
    theme := widgets.DefaultTheme()
    app := &dashboard{
        theme:    theme,
        input:    theme.NewInput(""),
        services: theme.NewList("api", "worker", "billing"),
        progress: widgets.NewProgress(0.35),
        logs:     theme.NewScrollView("[ready] application started"),
        deploy:   theme.NewButton(" Deploy "),
        status:   theme.NewText("Ready"),
    }
    app.input.Placeholder = "Service name"
    app.progress.ShowPercentage = true
    app.logs.Wrap = true
    app.logs.FollowEnd = true
    app.logs.ShowScrollbar = true

    app.root = app.buildLayout()
    app.router = widgets.NewRouter(app.root)
    app.router.FocusNext()
    app.bindEvents()
    return app
}
```

`View` 只返回已经保存的树：

```go
func (a *dashboard) View() widgets.Widget {
    a.router.SetRoot(a.root)
    return a.root
}
```

这样焦点、输入光标和滚动位置不会因为重绘丢失。

## 4. 构建响应式布局

`VStack` 从上到下排列，`HStack` 从左到右排列。`Add(widget, basis, flex)` 中：

- `basis > 0`：主轴固定占用多少终端单元格；
- `basis == 0`：按 `flex` 权重瓜分剩余空间；
- `Gap`：相邻组件之间的空隙。

下面创建左侧服务列表、右侧进度和日志：

```go
func (a *dashboard) buildLayout() widgets.Widget {
    inputBox := a.theme.NewBox(a.input)
    inputBox.Border = &widgets.RoundedBorder
    inputBox.Title = " Add service "

    servicesBox := a.theme.NewBox(a.services)
    servicesBox.Border = &widgets.RoundedBorder
    servicesBox.Title = " Services "

    logsBox := a.theme.NewBox(a.logs)
    logsBox.Border = &widgets.RoundedBorder
    logsBox.Title = " Activity "

    left := widgets.NewVStack()
    left.Gap = 1
    left.Add(inputBox, 3, 0)
    left.Add(servicesBox, 0, 1)

    right := widgets.NewVStack()
    right.Gap = 1
    right.Add(a.progress, 1, 0)
    right.Add(logsBox, 0, 1)
    right.Add(a.deploy, 1, 0)

    body := widgets.NewHStack()
    body.Gap = 1
    body.Add(left, 28, 0)
    body.Add(right, 0, 1)

    title := a.theme.NewText(" Star Service Console ")
    title.Style = a.theme.Primary.Bold(true)

    root := widgets.NewVStack()
    root.Add(title, 1, 0)
    root.Add(body, 0, 1)
    root.Add(a.status, 1, 0)
    return root
}
```

每帧的布局流程是 `Measure → Arrange → Render`。组件只应在 `Arrange` 得到的边界内
绘制；窗口变化时无需手工计算整个页面。

## 5. 路由键盘、鼠标和焦点

### Button 事件的最小完整写法

Button 不要求调用者手工判断 Enter、空格或鼠标按键。Router 会把这些输入统一转换为
`widgets.Click`：

```go
type buttonApp struct {
    save   *widgets.Button
    status *widgets.Text
    root   widgets.Widget
    router *widgets.Router
}

func newButtonApp() *buttonApp {
    theme := widgets.DefaultTheme()
    app := &buttonApp{
        save:   theme.NewButton(" Save "),
        status: theme.NewText("Not saved"),
    }

    // On 只注册一次，不要放进 View。
    app.save.On(widgets.Click, false, func(event *widgets.Event) {
        app.status.Content = "Saved"
    })

    app.root = widgets.NewVStack(app.save, app.status)
    app.router = widgets.NewRouter(app.root)
    app.router.FocusNext() // 让第一个可聚焦组件获得初始焦点
    return app
}

func (a *buttonApp) Init() star.Cmd { return nil }

func (a *buttonApp) Update(message star.Message) (star.Model, star.Cmd) {
    if key, ok := message.(star.KeyMsg); ok &&
        (key.Key == tcell.KeyEscape || key.Key == tcell.KeyCtrlC) {
        return a, star.QuitCmd()
    }
    star.RouteInput(a.router, message)
    return a, nil
}

func (a *buttonApp) View() widgets.Widget {
    a.router.SetRoot(a.root)
    return a.root
}

func main() {
    program := star.NewProgram(star.WithMouse(true))
    if err := program.Run(newButtonApp()); err != nil {
        panic(err)
    }
}
```

用户可以通过以下方式触发同一个回调：

- Tab 把焦点移到 Button 后按 Enter；
- Tab 把焦点移到 Button 后按空格；
- 启用 `WithMouse(true)` 后用鼠标点击 Button。

Button 文字默认在其布局边界内水平居中，也可以修改：

```go
button.Align = widgets.AlignLeft
button.Align = widgets.AlignCenter
button.Align = widgets.AlignRight
```

`Align` 只控制按钮内部文字的位置，不控制 Button 在父布局中的位置和高度。需要单行
按钮时，应让 VStack 明确分配一行，而不是把按钮作为弹性子项：

```go
root := widgets.NewVStack()
root.Add(text, 1, 0)
root.Add(button, 1, 0)
```

如果希望 Button 在一块较大的 Box 内保持自然尺寸并居中，使用 Box 的内容适配：

```go
box := widgets.NewBox(button)
box.Border = &widgets.RoundedBorder
box.FitContent = true
box.Align = widgets.AlignCenter
box.VerticalAlign = widgets.AlignMiddle
```

`FitContent=false` 是默认行为，此时 Child 填满 Box 内部区域，对齐属性没有可见效果。
这个默认值保证 List、Table、ScrollView 等需要占满面板的组件仍然正常伸展。

### 通用注册方式

所有内置组件都嵌入 `WidgetBase`，因此使用同一种注册方法：

```go
widget.On(eventKind, capture, handler)
```

- `eventKind` 是要监听的事件，例如 `Click` 或 `Changed`；
- `capture=false` 是普通用法，在目标或冒泡阶段执行；
- `capture=true` 用于父容器在事件到达目标前拦截；
- `handler` 在 `Update → RouteInput` 的调用过程中同步执行。

常用组件事件：

| 组件 | 常用事件 | 触发时机 |
|---|---|---|
| `Button` | `Click` | 鼠标点击，或获得焦点后按 Enter/空格 |
| `Input` | `Changed`、`Submitted` | 内容变化、按 Enter |
| `TextArea` | `Changed` | 多行内容变化 |
| `List` | `Changed`、`Submitted` | 选择变化、Enter 或项目点击 |
| `Table` | `Changed`、`Submitted` | 行选择变化、Enter 或行点击 |
| `Tree` | `Changed`、`Submitted` | 节点选择/展开变化、激活节点 |
| `Tabs` | `Changed` | 活动页发生变化 |
| `Form` | `Submitted` | 点击提交且所有字段校验成功 |
| `Modal` | `Dismissed` | Esc 或点击面板外部 |
| 可聚焦组件 | `Focused`、`Blurred` | Router 改变焦点 |
| 任意命中组件 | `PointerDown`、`PointerMove`、`PointerUp`、`Click` | 鼠标操作 |

例如监听输入框和列表：

```go
input.On(widgets.Changed, false, func(*widgets.Event) {
    status.Content = "正在输入：" + input.Value
})

input.On(widgets.Submitted, false, func(*widgets.Event) {
    status.Content = "提交：" + input.Value
})

list.On(widgets.Changed, false, func(*widgets.Event) {
    status.Content = "选择：" + list.Items[list.Selected]
})
```

`Progress`、`Text`、`Image` 和 `Canvas` 是显示组件，没有内置的值变化行为，但鼠标命中
后仍可以监听 `Click` 等指针事件。

### 从事件启动异步 Cmd

`On` 回调没有返回值，所以不能直接从回调返回 `Cmd`。需要让回调记录一个同步请求，
再由当前这次 `Update` 返回命令：

```go
type deployApp struct {
    deployRequested bool
    deploy          *widgets.Button
    router          *widgets.Router
    root            widgets.Widget
}

func newDeployApp() *deployApp {
    app := &deployApp{deploy: widgets.NewButton(" Deploy ")}
    app.root = widgets.NewVStack(app.deploy)
    app.router = widgets.NewRouter(app.root)
    app.router.FocusNext()
    app.deploy.On(widgets.Click, false, func(*widgets.Event) {
        app.deployRequested = true
    })
    return app
}

func (a *deployApp) Update(message star.Message) (star.Model, star.Cmd) {
    star.RouteInput(a.router, message)
    if a.deployRequested {
        a.deployRequested = false
        return a, startDeploymentCmd()
    }
    return a, nil
}
```

这种写法确保副作用仍由 Elm 运行时管理，回调本身不会启动失控的 goroutine。

### 把终端消息交给 Router

终端输入先以 `star.KeyMsg` 或 `star.MouseMsg` 到达 `Update`。调用 `RouteInput` 后，
Router 会完成命中测试、Tab 焦点移动、鼠标捕获和 Click 合成：

```go
func (a *dashboard) Update(message star.Message) (star.Model, star.Cmd) {
    switch message := message.(type) {
    case star.KeyMsg:
        if message.Key == tcell.KeyEscape || message.Key == tcell.KeyCtrlC {
            return a, star.QuitCmd()
        }
        star.RouteInput(a.router, message)
    case star.MouseMsg:
        star.RouteInput(a.router, message)
    case star.WindowSizeMsg:
        a.status.Content = fmt.Sprintf("terminal: %dx%d", message.Width, message.Height)
    }
    return a, nil
}
```

启用鼠标、粘贴和终端焦点报告：

```go
program := star.NewProgram(
    star.WithMouse(true),
    star.WithPaste(true),
    star.WithFocus(true),
)
```

组件事件通过 `On(kind, capture, handler)` 注册。回调发生在 `RouteInput` 的调用栈内，
因此这里修改 Model 中保存的组件仍然是串行的：

```go
func (a *dashboard) bindEvents() {
    a.input.On(widgets.Submitted, false, func(*widgets.Event) {
        name := strings.TrimSpace(a.input.Value)
        if name == "" {
            a.status.Content = "Service name is required"
            return
        }
        a.services.Items = append(a.services.Items, name)
        a.services.SetSelected(len(a.services.Items) - 1)
        a.input.SetValue("")
    })

    a.deploy.On(widgets.Click, false, func(*widgets.Event) {
        a.progress.Value = 0
        a.status.Content = "Deployment started"
    })
}
```

事件传播顺序为 Capture → Target → Bubble。需要拦截默认行为时调用
`event.PreventDefault()`；需要阻止继续传播时调用 `StopPropagation()`。

不要在 `View()` 中调用 `On`。`View` 会重复执行，在其中注册会让同一个事件触发越来越
多次。事件绑定应和组件创建一起放在 `newApp`、`newDashboard` 等构造函数中。

## 6. 周期更新进度和日志

`Tick` 只触发一次。周期任务应在收到消息后安排下一次 Tick，让 Model 随时可以决定
停止：

```go
type pulseMsg time.Time

func nextPulse() star.Cmd {
    return star.Tick(100*time.Millisecond, func(now time.Time) star.Message {
        return pulseMsg(now)
    })
}

func (a *dashboard) Init() star.Cmd { return nextPulse() }
```

在 `Update` 增加：

```go
case pulseMsg:
    a.progress.Value += 0.02
    if a.progress.Value >= 1 {
        a.progress.Value = 0
        a.logs.SetContent(a.logs.Content + "\n[ok] deployment completed")
    }
    return a, nextPulse()
```

`Progress` 会把值钳制到 `[0,1]`。未填充区域不绘制背景，能够继承父容器颜色；百分比
文字位于填充区域时使用填充色背景，位于未填充区域时保留父背景。

## 7. 表格、表单与弹窗

### Table

```go
table := theme.NewTable(
    widgets.TableColumn{Title: "Service"},
    widgets.TableColumn{Title: "Region", Width: 12},
    widgets.TableColumn{Title: "State", Width: 10},
)
table.Rows = [][]string{
    {"api", "us-east", "healthy"},
    {"worker", "eu-west", "degraded"},
}
table.On(widgets.Submitted, false, func(*widgets.Event) {
    row := table.Rows[table.Selected]
    status.Content = "Selected: " + strings.Join(row, " / ")
})
```

### Form

```go
name := theme.NewInput("")
form := theme.NewForm(widgets.FormField{
    Label: "Name",
    Input: name,
    Validate: func(value string) error {
        if strings.TrimSpace(value) == "" {
            return errors.New("required")
        }
        return nil
    },
})
form.On(widgets.Submitted, false, func(*widgets.Event) {
    status.Content = "Saved: " + form.Values()[0]
})
```

### Modal

弹窗打开时把 Modal 临时设为 Router Root，焦点和鼠标命中就会被限制在弹窗内：

```go
modal := theme.NewModal(theme.NewText("Press Esc to close"))
modal.Title = " Help "
modal.Width, modal.Height = 48, 10
modal.On(widgets.Dismissed, false, func(*widgets.Event) {
    modal = nil
    router.SetRoot(base)
    router.FocusNext()
})
router.SetRoot(modal)
```

应用的 `View` 应在弹窗存在时返回 `modal`，否则返回基础页面。不要只把 Modal 画在基础
页面之上而继续让 Router 指向旧根节点，否则用户仍可能操作弹窗后面的控件。

## 8. 从文件显示图片

`Image` 使用真彩色半块字符，一个终端单元格表达上下两个图片像素。它不依赖 Kitty、
Sixel 或 iTerm 私有协议。

从 PNG、JPEG 或 GIF 路径加载：

```go
picture, err := theme.NewImageFromPath("assets/logo.png")
if err != nil {
    return nil, fmt.Errorf("load logo: %w", err)
}
picture.Fit = widgets.ImageFitContain
picture.Align = widgets.AlignCenter
picture.VerticalAlign = widgets.ImageAlignMiddle
```

运行时替换：

```go
if err := picture.SetImagePath("assets/new-logo.jpg"); err != nil {
    status.Content = err.Error()
}
```

`SetImagePath` 失败时会保留原图。透明像素保留父组件背景，半透明像素与已有背景进行
混合。四种缩放模式为：

| 模式 | 行为 |
|---|---|
| `ImageFitContain` | 保持比例，完整显示，可能留白 |
| `ImageFitCover` | 保持比例，填满区域，可能裁剪 |
| `ImageFitFill` | 拉伸到整个区域，不保持比例 |
| `ImageFitNone` | 一个源像素对应一个终端半单元格 |

运行仓库内的交互示例：

```bash
go run ./examples/image ./assets/logo.png
```

## 9. 使用 Theme

`DefaultTheme` 提供 Text、Surface、Border、Primary、Focused、Selected、Disabled、
Scrollbar、Progress、Error 和 Backdrop 等语义样式。推荐从 Theme 创建组件，再对少数
组件进行局部覆盖：

```go
theme := widgets.DefaultTheme()
theme.Primary = tcell.StyleDefault.
    Foreground(tcell.ColorWhite).
    Background(tcell.ColorDarkCyan)

save := theme.NewButton(" Save ")
panel := theme.NewBox(save)
panel.Border = &widgets.DoubleBorder
```

不要在业务代码各处复制颜色常量。统一 Theme 可以让暗色、亮色和品牌主题更容易维护。

## 10. 自动化测试

### 直接测试 Update

`Update` 是普通 Go 方法，可以直接构造消息：

```go
func TestEscapeQuits(t *testing.T) {
    app := newDashboard()
    _, command := app.Update(star.KeyMsg{Key: tcell.KeyEscape})
    if command == nil {
        t.Fatal("Escape did not return a quit command")
    }
}
```

### 测试 Widget 绘制

```go
screen := tcell.NewSimulationScreen("UTF-8")
if err := screen.Init(); err != nil {
    t.Fatal(err)
}
defer screen.Fini()
screen.SetSize(20, 4)

text := widgets.NewText("你好")
text.Measure(widgets.Tight(20, 4))
text.Arrange(widgets.NewRect(0, 0, 20, 4))
text.Render(screen)

main, combining, style, width := screen.GetContent(0, 0)
_, _, _, _ = main, combining, style, width
```

### 测试完整 Program

```go
screen := tcell.NewSimulationScreen("UTF-8")
screen.SetSize(80, 24)
program := star.NewProgram(star.WithScreen(screen))

done := make(chan error, 1)
go func() { done <- program.Run(newDashboard()) }()

screen.PostEventWait(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))
if err := <-done; err != nil {
    t.Fatal(err)
}
```

Program 负责模拟 Screen 的 `Init` 和 `Fini`，测试不要重复调用。异步测试必须设置超时，
确保失败时不会永久挂起。

## 11. 生产化检查清单

在发布 TUI 前，建议逐项确认：

- Model 与有状态 Widget 不被多个 goroutine 直接修改；
- 耗时操作全部放入 Cmd，可长期运行的任务使用 `WithContext`；
- `Update` 处理 `Esc`、`Ctrl+C` 和业务退出流程；
- 交互组件保存在 Model 中，而不是每帧在 `View` 重建；
- 所有键盘和鼠标输入都经过同一个 Router；
- Modal 打开时切换 Router Root，关闭时恢复根节点与焦点；
- 窄窗口、零尺寸和 resize 场景经过 SimulationScreen 测试；
- 文件和网络错误显示给用户，不在组件回调中 panic；
- 日志视图设置容量策略，避免无限累积字符串；
- 发布前执行以下检查。

```bash
go test -count=1 ./...
go test -cover ./...
go vet ./...
go build ./...
```

支持 CGO 的环境还应执行：

```bash
go test -race ./...
```

## 12. 下一步

- 查看 [`examples/dashboard`](../examples/dashboard/main.go) 获取本教程控制台的完整实现。
- 查看 [`examples/showcase`](../examples/showcase/main.go) 了解 Tree、Tabs、Table、
  TextArea、Form、Modal 和实时图表的组合方式。
- 查看 [`examples/image`](../examples/image/main.go) 体验图片缩放和透明混合。
- 阅读 [架构与生命周期](architecture.md) 理解事件循环和终端恢复策略。
- 阅读 [自定义 Widget](custom-widget.md) 编写自己的绘制或交互组件。
- 阅读 [测试指南](testing.md) 获取模糊测试和事件测试细节。
