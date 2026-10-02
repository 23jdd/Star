# 自定义 Widget

## 最小实现

自定义 Widget 需要实现 `widgets.Widget`。推荐嵌入 `WidgetBase`，从而自动支持布局
边界、隐藏状态、父节点和事件监听器。

```go
package components

import (
    "github.com/23jdd/Star/widgets"
    "github.com/gdamore/tcell/v2"
)

type Badge struct {
    widgets.WidgetBase
    Label string
    Style tcell.Style
}

func NewBadge(label string) *Badge {
    return &Badge{Label: label, Style: tcell.StyleDefault.Bold(true)}
}

func (b *Badge) Measure(c widgets.Constraints) widgets.Size {
    return c.Constrain(widgets.Size{W: len([]rune(b.Label)) + 2, H: 1})
}

func (b *Badge) Arrange(bounds widgets.Rect) {
    b.SetBounds(bounds)
}

func (b *Badge) Render(screen tcell.Screen) {
    if b.Hidden() || b.Bounds().Empty() {
        return
    }
    x := b.Bounds().X
    y := b.Bounds().Y
    screen.SetContent(x, y, '[', nil, b.Style)
    for _, r := range []rune(b.Label) {
        x++
        if x >= b.Bounds().X+b.Bounds().W-1 {
            break
        }
        screen.SetContent(x, y, r, nil, b.Style)
    }
    if b.Bounds().W > 1 {
        screen.SetContent(b.Bounds().X+b.Bounds().W-1, y, ']', nil, b.Style)
    }
}

func (b *Badge) HandlerEvent(event *widgets.Event) {}
```

生产组件应使用 Unicode 字素宽度而不是简单 rune 数计算文本宽度；可以参考内置
`Text` 和 `Canvas`，或使用 `github.com/rivo/uniseg`。

## 自定义容器

容器还应实现：

```go
type childProvider interface {
    EventChildren() []widgets.Widget
}
```

`EventChildren` 的顺序应与渲染顺序一致，后面的子节点位于视觉上层。构造或 Arrange
期间还要连接父子事件链：

```go
func (c *Container) Arrange(bounds widgets.Rect) {
    c.SetBounds(bounds)
    for _, child := range c.Children {
        if provider, ok := child.(interface{ EventBase() *widgets.WidgetBase }); ok {
            provider.EventBase().SetParent(c)
        }
        child.Arrange(bounds)
    }
}
```

这样 Capture/Bubble 就能穿过第三方容器。没有嵌入 WidgetBase 的子组件仍可绘制，
但不会自动提供父节点和监听器存储。

## 可聚焦组件

实现 `Focusable` 即可加入 Router 的焦点顺序：

```go
func (w *Widget) CanFocus() bool { return !w.Disabled && !w.Hidden() }
func (w *Widget) SetFocus(value bool) { w.Focused = value }
```

Router 按深度优先顺序遍历。隐藏或禁用组件应从 `CanFocus` 返回 false。

## 默认行为

把组件自身行为赋给 `WidgetBase.DefaultAction`：

```go
func NewToggle() *Toggle {
    t := &Toggle{}
    t.DefaultAction = func(event *widgets.Event) {
        if event.Kind == widgets.Click {
            t.Checked = !t.Checked
        }
    }
    return t
}
```

应用可以监听事件并调用 `PreventDefault` 覆盖该行为。

## 绘制规则

- 只使用传入的 Screen，不要直接写 stdout/stderr。
- 尊重 Bounds 和 Hidden。
- SetContent 前检查屏幕及组件边界。
- 宽字符和组合字符应作为一个字素处理。
- Render 不应执行网络、文件或其他阻塞 I/O。
- Measure 和 Render 应尽量保持确定性，便于使用 SimulationScreen 测试。
