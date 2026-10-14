# AI 模型与只读账务助手

财务助手通过当前登录用户的七个只读工具发现账户、分类和标签，查询账单详情、完整范围统计及当前余额。默认包含隐藏的有效数据，排除已删除数据；查询范围可以覆盖全部历史和已有未来账单。金额按币种分别计算，不自动换汇。执行预算限制一次请求的时间和轮次，不截短账本范围。识图、文本识别和编程助手保留各自的 provider 配置。

助手无需 embedding 接口、模型或密钥。新版忽略旧 `[llm_assistant_embedding]` 整段以及 `openai_embedding_model_id`，也不会读取它们的 `EBK_` / `EBKCFP_` 覆盖值；旧配置无效或密钥文件缺失不会阻止启动。旧 `AIAssistantEmbedding` 模型对应的表（默认物理名 `a_i_assistant_embedding`，实际名称由引擎 mapper/前缀/schema 决定）和记录保留用于回滚，助手不读取、更新或清理它。

以下是配置示例。请使用支持函数工具的 OpenAI 或兼容服务，按供应商要求设置路径前缀；默认配置不开启 AI，也不包含密钥。

```ini
[ai]
transaction_from_ai_image_recognition = true
enable_ai_assistant = true

[llm_image_recognition]
llm_provider = openai
openai_base_url = https://api.example.com/v1/
openai_model_id = your-vision-model
chat_completions_stream = false
request_timeout = 180000

[llm_image_recognition_fallback]
llm_provider = openai
openai_base_url = https://fallback.example.com/v1/
openai_model_id = your-vision-model
request_timeout = 180000

[llm_assistant]
llm_provider = openai
openai_base_url = https://api.example.com/v1/
openai_model_id = your-tool-capable-model
chat_completions_stream = false
request_timeout = 180000
run_timeout = 180000
query_timeout = 30000
max_rounds = 12

[llm_assistant_fallback]
llm_provider = openai
openai_base_url = https://fallback.example.com/v1/
openai_model_id = your-tool-capable-model
request_timeout = 180000
```

浏览器 JSON 与 SSE 使用同一执行器，上游协议由 provider 配置选择：`llm_provider = openai` 使用 Chat Completions，`chat_completions_stream = true` 强制使用流式 Chat Completions；`llm_provider = openai_responses_compatible` 使用 Responses。Responses 配置填写 `openai_compatible_base_url`、`openai_compatible_api_key` 和 `openai_compatible_model_id`，对应密钥文件变量为 `EBKCFP_LLM_ASSISTANT_OPENAI_COMPATIBLE_API_KEY`（备用段同理）。主备模型均须支持其配置协议的完整工具调用和可靠结束标志；其他不具备工具能力的 provider 会明确失败。

模型工具轮完整结束后才执行查询；带工具的轮次正文不作为最终答案。SSE 实时输出状态及允许展示的思考，最终正文可能在完整轮次结束后才出现。一次请求最多切换备用模型一次，已经发送正文或思考后不再切换。识图备用接口只接管识图请求。

`run_timeout`（整体）和 `query_timeout`（单次查询）单位为毫秒，`max_rounds` 为轮数，均须为 1–4294967295 的整数。仅在 `[llm_assistant]` 配置这些预算；主备共用。供应商 `request_timeout` 与整体截止时间取较早者。浏览器 JSON/SSE 超时均为服务端 `run_timeout` 加 10 秒，反向代理读取超时应至少覆盖该值，例如默认预算配置 `proxy_read_timeout 190s`；SSE 每 15 秒心跳，并须关闭代理缓冲。预算耗尽、取消、查询错误和供应商断流会显示明确终态，不能当作成功空账本。

摘要默认查询本月截至现在的收支、上月同进度比较、本期分类分布和当前余额；显式期间优先。当前余额与历史收支净额独立，不能重建历史任意时点余额。会话与续查只在当前页面短期保留；“下一页”需要显式操作，每次续查重新鉴权。工具没有写账单、任意 SQL、附件内容识别或跨用户查询能力。

## 密钥文件

现有 `EBKCFP_` 环境变量可从文件读取活跃模型密钥：

```dotenv
EBKCFP_LLM_IMAGE_RECOGNITION_OPENAI_API_KEY=/run/ezbookkeeping-secrets/vision_api_key
EBKCFP_LLM_IMAGE_RECOGNITION_FALLBACK_OPENAI_API_KEY=/run/ezbookkeeping-secrets/fallback_api_key
EBKCFP_LLM_ASSISTANT_OPENAI_API_KEY=/run/ezbookkeeping-secrets/chat_api_key
EBKCFP_LLM_ASSISTANT_FALLBACK_OPENAI_API_KEY=/run/ezbookkeeping-secrets/fallback_api_key
```

容器须挂载文件并给予运行用户读取权限，建议文件权限 `0600`、只读挂载。活跃 provider 的密钥在启动配置日志中脱敏。旧 embedding 密钥文件无需挂载；保留旧配置与密钥的安全备份，供旧版本回滚使用。不要提交实际密钥。

## 验证

自动测试使用本地 mock provider 与临时 SQLite 数据库，不调用真实供应商。旧表的 SQLite authorizer 会拒绝任何缓存读写，验证两条助手 API 仍成功；数据库更新测试验证旧表和已有记录保留：

```sh
BUILD_PIPELINE=1 go test ./pkg/settings ./pkg/api ./pkg/services ./pkg/llm/... ./cmd
go test -race ./pkg/settings ./pkg/api ./cmd
go vet ./...
```

真实验证必须使用隔离的专用测试部署、测试账本与用户，并显式启用：

```sh
EBK_AI_DEPLOYMENT_CHECK=1 \
EBK_AI_DEPLOYMENT_ISOLATED=1 \
EBK_AI_DEPLOYMENT_CONFIG=/absolute/path/to/test-ezbookkeeping.ini \
EBK_AI_DEPLOYMENT_WORK_DIR=/absolute/path/to/ezbookkeeping \
EBK_AI_DEPLOYMENT_UID=<test-user-uid> \
go test -count=1 -timeout=12m -run '^TestAIDeploymentIntegration$' ./pkg/api
```

需要活跃模型的密钥环境变量及文件，工作目录包含 `templates/`。测试识别合成收据并检查金额 `12.34`，助手要求实际执行 `list_accounts` 和 `get_account_balances` 后分别在 JSON/SSE 路径完成；不创建账单、不访问 embedding 缓存。支持配置的 Chat Completions、强制流式 Chat Completions 与 Responses；可加 `EBK_AI_DEPLOYMENT_FORCE_FALLBACK=1` 将隔离主模型端点替换为本地 503，从而测试实际备用模型（必须已配置）。仅验证助手可使用 `-run '^TestAIDeploymentIntegration$/assistant'`。这项 smoke 只验证两工具及传输，七题财务真值仍须按验收记录单独核对。仅识图可使用 `-run '^TestAIDeploymentIntegration$/receipt$'`。真实模型、三库及双端 GUI 的执行记录和待验证项见 [验收记录](ai-assistant-harness-verification.md)；mock 通过不代表真实模型验收完成。
