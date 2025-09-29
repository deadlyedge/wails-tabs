# photoTidyGo

## 项目简介
photoTidyGo 是一个基于 Wails + React + Go 的跨平台桌面应用，目标是帮助用户整理海量照片与视频资料。采用 SQLite 持久化、结构化日志与现代前端栈，为后续的媒体扫描、规划与执行流程打下基础。

## 计划中
- **配置服务**：在启动时解析 `settings.toml`，支持 HOME/DATA 目录的环境变量覆盖，统一输出 UI 所需的展平配置。
- **SQLite 初始化**：定义媒体清单、计划项与操作日志三大表结构，schema 版本元数据。
- **核心工具库**：路径归一化、哈希计算（MD5/BLAKE3）、时间戳格式化与目录遍历等工具函数，供后续扫描/规划逻辑复用。
- **事件与日志**：配置 `tracing` 日志订阅器，预留应用内事件常量，确保前后端间的可观测性。
- **前端配置总览**：UI 在启动与事件广播时拉取配置快照，展示数据库位置、输入/输出目录、重复文件目录及可扫描的扩展名。
- **开发文档**：新增 `docs/setup.md` 指南，并在 README 中集中链接项目文档以便新人快速上手。
- **媒体扫描管线**：实现 `scan_media` 工作线程，递归枚举媒体文件、增量更新 SQLite，并结合哈希缓存与 EXIF 元数据提取。
- **规划与执行引擎**：`excute`，并构建复制/移动执行流程、操作日志与回滚能力。
- **前端工作流界面**：
  - 选择源文件夹/目标文件夹：扫描媒体数量；目标文件夹可用空间
  - 重建配置初始化、设置整理规范
  - 在monitor中显示计划审查、执行结果等关键页面，实时呈现事件进度与重复文件提醒。
- **打包发布**：完成 Windows/macOS/Linux 打包、权限校验、版本与签名流程，并筹备 Beta 发布。

## Building

To build a redistributable, production mode package, use `wails build` or:

```bash
wails build -ldflags="-s -w" -upx -upxflags="--best"
```
