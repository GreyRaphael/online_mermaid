import { expect, test } from '@playwright/test'
import { login } from './helpers'

// Sanitized generic topology: tall multi-layer diagram with nested subgraphs,
// used to assert that large diagrams render fully within the container bounds.
const testMermaidCode = `flowchart TD
    %% 1. 接入层
    subgraph AccessLayer["1. 用户接入层 (Access)"]
        Client["Web 浏览器 / 移动客户端"]
        CDN["CDN 静态加速节点"]
        WAF["Web 应用防火墙 (WAF)"]
        Client --> CDN
        CDN --> WAF
    end

    %% 2. 网关层
    subgraph GatewayLayer["2. 统一网关层 (Gateway)"]
        Gateway["API 网关<br/>• 鉴权与限流<br/>• 协议转换与路由"]
        AuthSvc["认证中心<br/>(Token / OAuth2)"]
        Gateway <--> AuthSvc
    end

    %% 3. 业务服务层
    subgraph ServiceLayer["3. 业务服务集群 (Services)"]
        direction TB
        OrderSvc["订单服务"]
        UserSvc["用户服务"]
        PaySvc["支付服务"]

        subgraph InfraPool["基础组件池"]
            Cache["分布式缓存"]
            MQ["消息队列"]
            Search["检索引擎"]
        end
        OrderSvc --> InfraPool
    end

    %% 4. 数据层
    subgraph DataLayer["4. 数据存储层 (Storage)"]
        direction TB
        PrimaryDB[("主数据库 (读写)")]
        ReplicaDB[("只读副本库")]
        PrimaryDB --> ReplicaDB
    end

    %% 5. 可观测层
    subgraph ObserveLayer["5. 可观测性 (Observability)"]
        Metrics["指标监控 (Prometheus)"]
        Logs["日志采集 (Loki)"]
        Tracing["链路追踪 (Tempo)"]
    end

    %% 主链路
    WAF -->|"HTTPS 转发"| Gateway
    Gateway -->|"路由 /orders"| OrderSvc
    Gateway -->|"路由 /users"| UserSvc
    Gateway -->|"路由 /payments"| PaySvc

    %% 数据访问
    OrderSvc --> PrimaryDB
    UserSvc --> PrimaryDB
    PaySvc --> MQ

    %% 旁路上报
    Gateway -.->|"访问日志"| Logs
    OrderSvc -.->|"指标埋点"| Metrics
    PaySvc -.->|"调用链 Span"| Tracing

    %% 样式定义
    style AccessLayer fill:#f8f9fa,stroke:#495057,stroke-width:1px
    style GatewayLayer fill:#e3f2fd,stroke:#1976d2,stroke-width:2px
    style ServiceLayer fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px
    style DataLayer fill:#fff3e0,stroke:#e65100,stroke-width:2px
    style ObserveLayer fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px

    style Gateway fill:#1976d2,color:#fff
    style OrderSvc fill:#2e7d32,color:#fff
    style PrimaryDB fill:#e65100,color:#fff
    style Cache fill:#ffeb3b,stroke:#fbc02d,color:#000
    style Metrics fill:#ab47bc,color:#fff`

test('Verify diagram renders fully within container bounds in Preview mode', async ({
  page,
}) => {
  await login(page)

  // Switch to Edit mode and paste code
  await page.getByRole('button', { name: '编辑' }).click()
  await page.locator('.code-textarea').fill(testMermaidCode)

  // Switch to Preview mode
  await page.getByRole('button', { name: '预览' }).click()
  const svg = page.locator('.mermaid-output svg')
  await expect(svg).toBeVisible({ timeout: 20_000 })
  await page.waitForTimeout(600)

  // Check metrics
  const previewMetrics = await page.evaluate(() => {
    const svgEl = document.querySelector('.mermaid-output svg') as SVGSVGElement
    const containerEl = document.querySelector('.canvas-viewport') as HTMLElement
    const containerRect = containerEl.getBoundingClientRect()
    const svgRect = svgEl.getBoundingClientRect()

    return {
      containerRect: {
        x: Math.round(containerRect.x),
        y: Math.round(containerRect.y),
        width: Math.round(containerRect.width),
        height: Math.round(containerRect.height),
        top: Math.round(containerRect.top),
        bottom: Math.round(containerRect.bottom),
      },
      svgRect: {
        x: Math.round(svgRect.x),
        y: Math.round(svgRect.y),
        width: Math.round(svgRect.width),
        height: Math.round(svgRect.height),
        top: Math.round(svgRect.top),
        bottom: Math.round(svgRect.bottom),
      },
      overflowTop: svgRect.top < containerRect.top - 1,
      overflowBottom: svgRect.bottom > containerRect.bottom + 1,
      overflowLeft: svgRect.left < containerRect.left - 1,
      overflowRight: svgRect.right > containerRect.right + 1,
    }
  })

  console.log('=== PREVIEW MODE METRICS ===', JSON.stringify(previewMetrics, null, 2))

  // Ensure NO overflow in any direction
  expect(previewMetrics.overflowTop).toBe(false)
  expect(previewMetrics.overflowBottom).toBe(false)
  expect(previewMetrics.overflowLeft).toBe(false)
  expect(previewMetrics.overflowRight).toBe(false)

  // Ensure top node and bottom node are visible in viewport
  await expect(page.locator('.mermaid-output')).toContainText('用户接入层')
  await expect(page.locator('.mermaid-output')).toContainText('可观测性')
})

test('Verify diagram renders fully within container bounds in Split mode', async ({
  page,
}) => {
  await login(page)

  // Switch to Split mode
  await page.getByRole('button', { name: '分屏' }).click()
  await page.locator('.code-textarea').fill(testMermaidCode)

  const svg = page.locator('.mermaid-output svg')
  await expect(svg).toBeVisible({ timeout: 20_000 })
  await page.waitForTimeout(600)

  const splitMetrics = await page.evaluate(() => {
    const svgEl = document.querySelector('.mermaid-output svg') as SVGSVGElement
    const containerEl = document.querySelector('.canvas-viewport') as HTMLElement
    const containerRect = containerEl.getBoundingClientRect()
    const svgRect = svgEl.getBoundingClientRect()

    return {
      containerRect: {
        x: Math.round(containerRect.x),
        y: Math.round(containerRect.y),
        width: Math.round(containerRect.width),
        height: Math.round(containerRect.height),
      },
      svgRect: {
        x: Math.round(svgRect.x),
        y: Math.round(svgRect.y),
        width: Math.round(svgRect.width),
        height: Math.round(svgRect.height),
      },
      overflowTop: svgRect.top < containerRect.top - 1,
      overflowBottom: svgRect.bottom > containerRect.bottom + 1,
      overflowLeft: svgRect.left < containerRect.left - 1,
      overflowRight: svgRect.right > containerRect.right + 1,
    }
  })

  console.log('=== SPLIT MODE METRICS ===', JSON.stringify(splitMetrics, null, 2))

  expect(splitMetrics.overflowTop).toBe(false)
  expect(splitMetrics.overflowBottom).toBe(false)
  expect(splitMetrics.overflowLeft).toBe(false)
  expect(splitMetrics.overflowRight).toBe(false)
})
