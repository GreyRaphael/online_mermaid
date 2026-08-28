import { expect, test } from '@playwright/test'

const testMermaidCode = `flowchart TD
    %% 1. 客户端层
    subgraph ClientLayer["1. 客户端接入层 (Client & Ingress)"]
        Client["开发者终端 / IDE / OpenCode / Cursor / 量化脚本"]
        WinLocal["本地代理 (Win10 / WSL 桥接)<br/>172.19.112.1:4001"]
        aTrust["aTrust 零信任虚拟专网隧道"]
        Client --> WinLocal
        WinLocal --> aTrust
    end

    %% 2. LiteLLM 统一管控层
    subgraph GatewayLayer["2. 统一管控网关层 (LiteLLM Gateway)"]
        LiteLLM["LiteLLM Proxy 网关 (v1.92.0)<br/>mcphub-test.csc.com.cn:4000<br/>• 负责 APIKey 鉴权 (如 衍生品交易部 5000元 配额)<br/>• 模型名别名重写 & 敏感数据审计拦截"]
        DB[("PostgreSQL 数据库<br/>(密钥/团队/计费/策略)")]
        LiteLLM <--> DB
    end

    %% 3. 致胜中转与本地大算力池
    subgraph ZhiShengCluster["3. 致胜 AI 综合网关 & 本地推理算力池 (10.237.103.82)"]
        direction TB
        Nginx18889["致胜中转网关 (10.237.103.82:18889)<br/>Nginx 1.20.1 + Java 网关 (/ai/llm/chat)<br/>🟢 零信任内网完全免密直连 / 支持流式与思考链"]

        subgraph LocalGPU["致胜本地 GPU 实例池 (owned_by: local)"]
            M_235B["🔥 Qwen3-235B-A22B-Instruct-2507-zxjt (235B MoE 旗舰)"]
            M_72B["🌟 Qwen72 (72B 密集大模型)"]
            M_70B["🌟 deepseek-70b (70B 推理模型)"]
            M_V31["deepseek-v3.1-zxjt (V3.1 内部版)"]
            M_Qwen32["qwen3-32b-zxjt / qwq-32b-zxjt"]
            M_Flash_ZXJT["deepseek-v4-flash-zxjt"]
        end
        Nginx18889 --> LocalGPU
    end

    %% 4. AIOps 容器与底层算力集群
    subgraph AIOpsCluster["4. AIOps 容器平台 & 物理推理节点 (10.48.104.17)"]
        direction TB
        AIOps_K8s["AIOps K8s 生产网关 (端口 32550)<br/>• 强制 APIKey 鉴权<br/>• 挂载 DeepSeek-V3 671B & DS-V4-Flash"]

        subgraph DualReplicas["Flash 双活物理容器副本 (512K 上下文)"]
            Node_8443["8443 端口实例 (vLLM)<br/>max_model_len: 524288"]
            Node_8444["8444 端口实例 (vLLM)<br/>max_model_len: 524288"]
        end
    end

    %% 5. 云端算力
    subgraph CloudLayer["5. 云端采购与外部中继"]
        AliBailian["阿里云百炼平台 (Bailian)<br/>• GLM-5.2 (1M 上下文)"]
        SiliconFlow["硅基流动 API (api.siliconflow.cn)<br/>• ❌ 已在网关被 Blocked 停用"]
    end

    %% 链路关系连线
    %% 途径 A: 走 LiteLLM 统一调度
    aTrust -->|"【途径 A】标准调度 (带 Key 鉴权)"| LiteLLM
    LiteLLM -->|"① glm-5.2 / deepseek-v4-pro / qwen3.8-27b"| Nginx18889
    LiteLLM -->|"② deepseek-v4-flash-0731 (重写为 zxjt)"| Nginx18889
    LiteLLM -->|"③ deepseek-v3-0324 (带集团 Token)"| AIOps_K8s

    %% 途径 B: 直连致胜网关
    aTrust -.->|"【途径 B】免密直连 (白嫖算力/无配额限制)"| Nginx18889

    %% 途径 C: 直连底层的 8443/8444
    aTrust -.->|"【途径 C】直连双活底层节点 (512K)"| DualReplicas

    %% 18889 向上游与底层连接
    Nginx18889 -->|"百炼中继通道"| AliBailian
    Nginx18889 -.->|"底层 Upstream 负载分流"| DualReplicas

    %% 样式定义
    style ClientLayer fill:#f8f9fa,stroke:#495057,stroke-width:1px
    style GatewayLayer fill:#e3f2fd,stroke:#1976d2,stroke-width:2px
    style ZhiShengCluster fill:#e8f5e9,stroke:#2e7d32,stroke-width:2px
    style AIOpsCluster fill:#fff3e0,stroke:#e65100,stroke-width:2px
    style CloudLayer fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px

    style LiteLLM fill:#1976d2,color:#fff
    style Nginx18889 fill:#2e7d32,color:#fff
    style AIOps_K8s fill:#e65100,color:#fff
    style M_235B fill:#ffeb3b,stroke:#fbc02d,color:#000
    style AliBailian fill:#ab47bc,color:#fff
    style SiliconFlow fill:#e0e0e0,color:#9e9e9e,stroke-dasharray: 5 5`

test('Verify diagram renders fully within container bounds in Preview mode', async ({
  page,
}) => {
  await page.goto('/')
  await page.getByLabel('用户名').fill('admin')
  await page.getByLabel('密码').fill('admin123')
  await page.getByRole('button', { name: '登录' }).click()
  await expect(page.getByRole('button', { name: '切换侧栏' })).toBeVisible()

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
  await expect(page.locator('.mermaid-output')).toContainText('客户端接入层')
  await expect(page.locator('.mermaid-output')).toContainText('云端采购与外部中继')
})

test('Verify diagram renders fully within container bounds in Split mode', async ({
  page,
}) => {
  await page.goto('/')
  await page.getByLabel('用户名').fill('admin')
  await page.getByLabel('密码').fill('admin123')
  await page.getByRole('button', { name: '登录' }).click()
  await expect(page.getByRole('button', { name: '切换侧栏' })).toBeVisible()

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
