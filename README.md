# Online Mermaid

一个简洁优雅、安全高效的在线 Mermaid 编辑器与渲染平台。基于 Go + SQLite3 + Vue 3 + Mermaid 11 构建，单二进制文件分发，无需额外依赖。

## 🌟 核心特性

1. **🔐 登录与会话认证**
   - 基于 Bcrypt 强哈希密码校验。
   - 安全 HttpOnly Session Cookie、同源保护与暴力破解频率限制（Rate Limiter）。
   - 提供 `hash-password` 命令行工具快速生成密码哈希。

2. **📑 三种视图模式**
   - 👁 **预览 (Preview)**：大画布沉浸式查看与缩放拖拽。
   - ✏️ **编辑 (Edit)**：带行号统计、语法模版快速插入的代码编辑器。
   - 📑 **分屏 (Split)**：左侧编辑、右侧防抖实时渲染。
   - 快捷保存（`Ctrl+S` / `Cmd+S`）与未保存修改状态提示。
   - LocalStorage 本地草稿自动恢复机制，防止断电或误关丢失。

3. **📂 左栏图表管理**
   - 快速搜索过滤图表。
   - **自动递增命名**：新建图表自动识别序号（`graph1`, `graph2`, `graph3`...）。
   - 支持自定义重命名、图表复制（副本）、安全删除确认。
   - 支持侧边栏折叠/展开。

4. **💾 纯 Go SQLite3 存储**
   - 采用 `modernc.org/sqlite`，纯 Go 实现，零 CGO 编译依赖。
   - 数据持久化在 `mermaid.db`，支持 WAL 高性能模式。
   - 初次启动自动初始化示例图表 `graph1`。

5. **🎨 完整的图表交互能力**
   - **平移与缩放**：内置 `panzoom` 手势支持，鼠标滚轮缩放、双击与拖拽。
   - **图片导出与复制**：
     - 📋 复制 Mermaid 源码
     - 🖼️ 复制白底 PNG 图片到剪切板
     - 💾 导出透明背景高清 PNG
     - 📐 导出 SVG 矢量图
   - **全屏模态框**：沉浸式全屏图表查看，支持 90° 旋转与自适应适配。
   - **精准错误捕获**：解析报错时高亮错误行号（如 `第 3 行解析错误`），不破坏界面布局。
   - **内置丰富模版**：Flowchart 流程图、Sequence 时序图、Class 类图、State 状态图、ER 实体关系图、Gantt 甘特图、Pie 饼图、Mindmap 思维导图、GitGraph 分支图等。

6. **🌗 Day / Night 主题切换**
   - 支持 `Day`（浅色）、`Night`（深色）与 `System`（跟随系统）三种模式。
   - Mermaid 引擎与界面主题实时联动（Day -> default, Night -> dark）。

---

## 🚀 快速开始

### 1. 编译构建

```bash
# 构建前端并编译生成可执行二进制文件 bin/online-mermaid
make build
```

### 2. 运行服务

```bash
# 启动服务（默认监听 0.0.0.0:8850，默认用户名 admin，默认密码 admin123）
./bin/online-mermaid
```

打开浏览器访问 `http://localhost:8850` 即可使用。

---

## ⚙️ 配置参数

可通过命令行参数或环境变量进行配置：

| 命令行参数 | 环境变量 | 默认值 | 描述 |
| :--- | :--- | :--- | :--- |
| `-addr` | `ONLINE_MERMAID_ADDR` | `0.0.0.0:8850` | HTTP 服务监听地址 |
| `-db-path` | `ONLINE_MERMAID_DB` | `~/.config/online-mermaid/mermaid.db` | SQLite 数据库文件路径 |
| `-admin-user` | `ONLINE_MERMAID_ADMIN_USERNAME` | `admin` | 管理员用户名 |
| `-password-hash`| `ONLINE_MERMAID_ADMIN_PASSWORD_HASH` | 默认 `admin123` 对应哈希 | Bcrypt 密码哈希 |
| `-session-ttl` | `ONLINE_MERMAID_SESSION_TTL` | `72h` | 登录会话有效期 |
| `-secure-cookie`| `ONLINE_MERMAID_SECURE_COOKIE` | `false` | 是否启用 Secure Cookie (HTTPS 环境开启) |

### 生成自定义密码哈希

```bash
./bin/online-mermaid hash-password
# 输入自定义密码后回车即可输出 Bcrypt 哈希字符串
```

示例使用自定义密码启动：
```bash
./bin/online-mermaid -admin-user myuser -password-hash '$2a$10$ZV1d9Joh7H9OAGZ/iJP15u0hYkGfpfsB6rCY9WXuaMYNNmtd1jLQm'
```

---

## 🛠️ 项目目录结构

```
online_mermaid/
├── cmd/
│   └── online-mermaid/main.go        # 服务入口与 CLI
├── internal/
│   ├── auth/                        # 鉴权与 Session 管理
│   ├── config/                      # 配置解析
│   ├── db/                          # SQLite3 数据库驱动与图表 CRUD
│   ├── server/                      # HTTP 路由、中间件与 SPA 服务
│   └── webui/                       # 前端嵌入文件 (go:embed)
├── web/                             # Vue 3 + TypeScript 前端源码
│   ├── src/
│   │   ├── api/                     # 后端 API 客户端
│   │   ├── components/              # 编辑器、预览画布、侧边栏、全屏弹窗
│   │   ├── composables/             # useTheme 主题控制
│   │   ├── styles/                  # 设计系统 CSS Token
│   │   ├── utils/                   # 图表模版库、导出与图标工具
│   │   └── views/                   # 登录页与主编辑页
│   └── package.json
├── Makefile                         # 构建工具
├── go.mod
└── go.sum
```
