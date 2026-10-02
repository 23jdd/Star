# Star

Star 是一个基于 [tcell](https://github.com/gdamore/tcell) 的 Elm 风格 Go
终端 UI 框架。应用状态只在串行的 `Update` 中更新，异步任务通过消息返回结果：

```text
Message -> Update -> Model -> View -> Measure / Arrange / Render
                `-> Cmd -> Message
```

## 文档导航

- [安装与最小应用](#安装)
- [运行时](#运行时)
- [Widget 与布局](#widget-与布局)
- [焦点与输入路由](#焦点与输入路由)
- [Program 配置](#program-配置)
- [示例](#示例)
- [测试](#测试)
- [完整 API 手册](docs/api.md)
- [架构与生命周期](docs/architecture.md)
- [自定义 Widget](docs/custom-widget.md)
- [测试指南](docs/testing.md)
- [常见问题](docs/faq.md)

## 安装

```bash
go get github.com/23jdd/Star
```

## 最小应用

```go
package main

import (
    "fmt"

    star "github.com/23jdd/Star"
    "github.com/23jdd/Star/widgets"
    "github.com/gdamore/tcell/v2"
)

type model struct{ count int }

func (m model) Init() star.Cmd { return nil }

func (m model) Update(message star.Message) (star.Model, star.Cmd) {
    switch message := message.(type) {
    case star.KeyMsg:
        switch {
        case message.Rune == '+':
            m.count++
        case message.Rune == '-':
            m.count--
        case message.Rune == 'q' || message.Key == tcell.KeyEscape:
            return m, star.QuitCmd()
        }
    }
    return m, nil
}

func (m model) View() widgets.Widget {
    return widgets.NewVStack(
        widgets.NewText("计数器"),
        widgets.NewButton(fmt.Sprintf(" %d ", m.count)),
        widgets.NewText("按 +/- 修改，按 q 退出"),
    )
}

func main() {
    if err := star.NewProgram().Run(model{}); err != nil {
        panic(err)
    }
}
```

## 运行时

- `KeyMsg`、`MouseMsg`、`WindowSizeMsg`、`PasteMsg` 和 `FocusMsg` 将终端输入
  转换为模型消息，不会把并发泄漏到 `Update` 中。
- `QuitCmd`、`After`、`Tick`、`Batch`、`Sequence` 和 `WithContext` 用于退出、
  定时、组合异步任务以及绑定任务生命周期。
- `RunContext` 支持取消；`Send`、`SendContext` 和 `TrySend` 可以从其他
  goroutine 安全地向程序发送消息。
- 模型或命令发生 panic 时，Star 会先恢复终端，再返回 `*star.PanicError`。
- `WithScreen` 可以注入 tcell 模拟屏幕，便于编写确定性的自动化测试。

程序发送给 `Update` 的第一条消息始终是 `WindowSizeMsg`。随后 Star 会启动
`Init` 返回的命令，并立即绘制首帧。

## Widget 与布局

每次绘制都会依次调用 `Measure`、`Arrange` 和 `Render`。根 Widget 会占满整个
终端区域。`VStack` 和 `HStack` 支持固定尺寸与弹性尺寸：

```go
row := widgets.NewHStack()
row.Gap = 1
row.Add(sidebar, 24, 0) // 固定为 24 列
row.Add(content, 0, 1)  // 占用全部剩余空间
```

`WidgetBase` 保存布局边界、可见性、父节点和事件监听器。`HitTest` 用于查找
坐标下最上层的 Widget，`Dispatch` 按照 Capture → Target → Bubble 的顺序分发
事件，并支持停止传播和取消默认行为。文本绘制能够识别 Unicode 字素宽度，且会
裁剪到组件的布局边界内。

内置组件包括：

- `Text`：Unicode 裁剪、自动换行和水平对齐
- `Button`、单行 `Input`、多行 `TextArea`、可选择的 `List` 和 `ScrollView`
- `Table`、`Tree`、`Tabs`、`Modal` 和带校验的 `Form`
- `VStack`、`HStack`、`Box`、`Overlay` 和 `Spacer`
- `Progress` 进度条、真彩色半块渲染的 `Image` 和用于自定义绘制的 `Canvas`
- `Theme`：统一管理语义样式，同时允许单个组件覆盖样式

从 PNG、JPEG 或 GIF 路径创建图片组件：

```go
picture, err := widgets.NewImageFromPath("assets/logo.png")
if err != nil {
    return err
}
picture.Fit = widgets.ImageFitContain
```

## 焦点与输入路由

`Router` 负责焦点顺序、Tab/Shift-Tab 导航、鼠标捕获、滚动以及 Click 事件合成。
需要交互状态的 Widget 和 Router 应保存在模型中，然后在 `Update` 中转发输入：

```go
type model struct {
    input  *widgets.Input
    save   *widgets.Button
    root   widgets.Widget
    router *widgets.Router
}

func (m *model) Update(message star.Message) (star.Model, star.Cmd) {
    star.RouteInput(m.router, message)
    return m, nil
}
```

使用 `WidgetBase.On` 注册捕获或冒泡监听器。文字编辑、列表移动、Tab 导航和点击
等默认行为会在事件传播结束后执行；监听器可以调用 `event.PreventDefault()`
取消默认行为。

## Program 配置

```go
program := star.NewProgram(
    star.WithMouse(true),
    star.WithPaste(true),
    star.WithFocus(true),
    star.WithQueueSize(512),
)
```

可用配置：

- `WithMouse`：启用鼠标事件
- `WithPaste`：启用括号粘贴边界事件
- `WithFocus`：启用终端焦点变化事件
- `WithQueueSize`：设置内部消息队列容量
- `WithScreen`：注入自定义或模拟的 tcell Screen

## 示例

- [`examples/main.go`](examples/main.go)：最小定时计数器
- [`examples/form/main.go`](examples/form/main.go)：输入框、列表、焦点、鼠标、边框和事件回调
- [`examples/dashboard/main.go`](examples/dashboard/main.go)：异步命令、响应式固定/弹性布局、
  主题、进度条、服务编辑以及滚动活动日志组成的完整仪表盘
- [`examples/components/main.go`](examples/components/main.go)：Table、Tree、Tabs、Modal、
  Form 校验和键盘/鼠标交互组成的高级组件展示
- [`examples/showcase/main.go`](examples/showcase/main.go)：完整控制中心，包含树导航、
  实时指标、字符图表、服务表格、日志、编辑器、设置表单、帮助弹窗和命令面板
- [`examples/image/main.go`](examples/image/main.go)：透明背景、真彩色和四种缩放模式的
  Image 组件交互展示

运行仪表盘示例：

```bash
go run ./examples/dashboard
go run ./examples/components
go run ./examples/showcase
go run ./examples/image
go run ./examples/image ./assets/logo.png
```

## 测试

```bash
go test ./...
go vet ./...
```

测试使用 `tcell.NewSimulationScreen`，无需真实终端即可验证输入转换、生命周期、
Unicode 绘制、布局、焦点导航和事件传播。

## 兼容性与稳定性

- 当前 `go.mod` 使用 Go 1.27.1。
- 终端能力由 tcell 决定，支持 Windows Terminal、常见 Linux/macOS 终端以及
  tcell 模拟屏幕。
- Star 在 `Update` 中串行处理消息；`Cmd` 可以并发执行，但不得直接修改模型。
- 当前 API 处于首个发布阶段。破坏兼容性的变更会记录在 [CHANGELOG.md](CHANGELOG.md)。

## 参与贡献

提交代码前请阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。安全问题请按照
[SECURITY.md](SECURITY.md) 中的方式报告。

## 许可证

Star 使用 [MIT License](LICENSE)。
