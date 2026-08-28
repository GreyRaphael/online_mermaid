export interface MermaidTemplate {
  id: string
  name: string
  category: string
  code: string
}

export const MERMAID_TEMPLATES: MermaidTemplate[] = [
  {
    id: 'flowchart',
    name: '流程图 (Flowchart)',
    category: '常用',
    code: `flowchart TD
    Start([开始]) --> Process[处理步骤]
    Process --> Cond{是否满足条件?}
    Cond -->|是| Success[成功完成]
    Cond -->|否| Retry[重试]
    Retry --> Process
    Success --> End([结束])
`,
  },
  {
    id: 'sequence',
    name: '时序图 (Sequence)',
    category: '常用',
    code: `sequenceDiagram
    autonumber
    actor User as 用户
    participant Browser as 浏览器
    participant Server as 后端服务
    participant DB as SQLite 数据库

    User->>Browser: 点击保存图表 (Ctrl+S)
    Browser->>Server: POST /api/diagrams/:id
    Server->>DB: UPDATE diagrams SET code=...
    DB-->>Server: OK (Affected 1)
    Server-->>Browser: 200 OK (Diagram)
    Browser-->>User: 提示 "✓ 已保存"
`,
  },
  {
    id: 'class',
    name: '类图 (Class Diagram)',
    category: '结构',
    code: `classDiagram
    class Diagram {
        +String id
        +String title
        +String code
        +DateTime createdAt
        +DateTime updatedAt
        +render()
        +exportPNG()
    }
    class Server {
        +Config config
        +DB database
        +handleRequest()
    }
    Server --> Diagram : manages
`,
  },
  {
    id: 'state',
    name: '状态图 (State Diagram)',
    category: '行为',
    code: `stateDiagram-v2
    [*] --> Idle: 初始化
    Idle --> Editing: 键盘输入
    Editing --> Rendering: 触发防抖 (200ms)
    Rendering --> Rendered: 渲染成功
    Rendering --> Error: 语法错误
    Error --> Editing: 修改源码
    Rendered --> Saving: 按下 Ctrl+S
    Saving --> Rendered: 保存成功
    Rendered --> [*]: 关闭页面
`,
  },
  {
    id: 'er',
    name: '实体关系图 (ER Diagram)',
    category: '数据',
    code: `erDiagram
    DIAGRAMS {
        string id PK
        string title
        string code
        datetime created_at
        datetime updated_at
        int sort_order
    }
    USERS {
        string username PK
        string password_hash
    }
    USERS ||--o{ DIAGRAMS : owns
`,
  },
  {
    id: 'gantt',
    name: '甘特图 (Gantt Chart)',
    category: '项目',
    code: `gantt
    title 项目开发进度甘特图
    dateFormat  YYYY-MM-DD
    section 后端架构
    SQLite 存储驱动   :done, 2026-08-20, 2d
    REST API 与鉴权   :done, 2026-08-22, 2d
    section 前端实现
    编辑器与分屏联动  :active, 2026-08-25, 3d
    图表全屏与导出    : 2026-08-28, 2d
`,
  },
  {
    id: 'pie',
    name: '饼图 (Pie Chart)',
    category: '统计',
    code: `pie title 常用 Mermaid 图表类型占比
    "流程图 Flowchart" : 45
    "时序图 Sequence" : 25
    "类图 Class" : 12
    "状态图 State" : 10
    "其他 Others" : 8
`,
  },
  {
    id: 'mindmap',
    name: '思维导图 (Mindmap)',
    category: '思维',
    code: `mindmap
  root((Online Mermaid))
    核心特性
      登录认证
      三模式切换
        预览
        编辑
        分屏
      左栏管理
        自动递增命名
        实时搜索
    数据存储
      SQLite3 无依赖
      毫秒级读写
    渲染交互
      PanZoom 缩放
      白底/透明导出
      全屏旋转
`,
  },
  {
    id: 'gitGraph',
    name: 'Git 分支图 (GitGraph)',
    category: '开发',
    code: `gitGraph
    commit
    commit
    branch develop
    checkout develop
    commit id: "feat: add sqlite"
    commit id: "feat: view modes"
    checkout main
    merge develop
    commit id: "v1.0.0" tag: "v1.0.0"
`,
  },
]
