# 变更记录

本项目遵循[语义化版本](https://semver.org/lang/zh-CN/)。尚未发布的修改记录在
“未发布”章节中。

## 未发布

### 新增

- Elm 风格的 `Model`、`Update`、`View` 和 `Cmd` 运行时。
- 键盘、鼠标、窗口尺寸、粘贴和终端焦点消息。
- `Batch`、`Sequence`、`After`、`Tick`、`WithContext` 和安全退出命令。
- Context 取消、外部消息注入、panic 恢复和模拟 Screen 注入。
- 约束布局、垂直/水平 Stack、Box、Overlay 和 Spacer。
- Text、Button、Input、List、ScrollView、Progress 和 Canvas。
- TextArea、Table、Tree、Tabs、Modal 和带校验的 Form 高级组件。
- Capture、Target、Bubble 事件传播，焦点导航、鼠标捕获和 Click 合成。
- Unicode 字素宽度处理、主题、示例程序和自动化测试。
- 自定义容器父链 API、模糊测试、性能基准和跨平台 CI。
- 中文 API、架构、自定义组件、测试、FAQ、贡献和安全文档。
- 高级组件综合示例。
- 集成全部布局、交互和高级组件的实时控制中心 Showcase。
- 支持透明合成、四种缩放模式和半块真彩色渲染的 Image 组件及交互示例。
- Image 支持从 PNG、JPEG、GIF 文件路径创建及运行时安全替换。
- 从最小应用到布局、路由、异步、图片、测试与发布检查的完整中文教程。
- 教程新增 Button 与其他 Widget 的事件注册、事件类型和异步 Cmd 衔接说明。
- Button 新增左、中、右文字对齐，`NewButton` 默认居中。
- Box 新增可选的内容尺寸适配及水平、垂直对齐，同时保留默认填满行为。

### 兼容性说明

- 首次正式发布前 API 仍可能调整；发布后将按照语义化版本管理破坏性变更。
