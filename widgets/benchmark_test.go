package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func BenchmarkTextLayout(b *testing.B) {
	text := NewText("Star 支持 Unicode 字素、自动换行和终端宽度裁剪。")
	text.Wrap = true
	constraints := Loose(40, 20)
	b.ReportAllocs()
	for b.Loop() {
		text.Measure(constraints)
		text.Arrange(NewRect(0, 0, 40, 20))
	}
}

func BenchmarkDashboardLikeRender(b *testing.B) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		b.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(120, 40)
	list := NewList("api", "billing", "search", "notifications")
	logs := NewScrollView("ready\nloading\nhealthy")
	root := NewHStack(NewBox(list), NewBox(logs))
	root.Arrange(NewRect(0, 0, 120, 40))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		root.Render(screen)
	}
}
