# AI 模型与独立嵌入接口

AI 助手的生成模型和向量检索模型可以使用不同的 OpenAI 兼容接口、密钥和模型名称。识图的主接口与备用接口单独配置。

以下示例对应 `gpt-6-astra` 生成模型和 `qwen3.7-text-embedding-flash` 嵌入模型。请合并到实际部署配置中；仓库中的默认配置不会启用 AI 功能，也不包含部署密钥。

```ini
[llm]
transaction_from_ai_image_recognition = true
enable_ai_assistant = true

[llm_image_recognition]
llm_provider = openai
openai_base_url = https://api.burn.hair/v1/
openai_model_id = gpt-6-astra
chat_completions_stream = true
request_timeout = 180000

[llm_image_recognition_fallback]
llm_provider = openai
openai_base_url = https://proxy.xhblog.top/v1/
openai_model_id = gpt-6-astra
chat_completions_stream = true
request_timeout = 180000

[llm_assistant]
llm_provider = openai
openai_base_url = https://proxy.xhblog.top/v1/
openai_model_id = gpt-6-astra
chat_completions_stream = true
request_timeout = 180000

[llm_assistant_fallback]
llm_provider = openai
openai_base_url = https://proxy.xhblog.top/v1/
openai_model_id = gpt-6-astra
chat_completions_stream = true
request_timeout = 180000

[llm_text_recognition]
llm_provider = openai
openai_base_url = https://proxy.xhblog.top/v1/
openai_model_id = gpt-6-astra
chat_completions_stream = true
request_timeout = 180000

[llm_assistant_embedding]
llm_provider = openai
openai_base_url = https://<workspace-id>.cn-beijing.maas.aliyuncs.com/compatible-mode/v1/
openai_embedding_model_id = qwen3.7-text-embedding-flash
request_timeout = 60000
```

`llm_assistant_embedding` 留空或不配置时，继续复用 `llm_assistant` 的接口和密钥，兼容原有安装。独立配置目前要求 `llm_provider = openai` 和非空的 `openai_embedding_model_id`。接口地址必须包含供应商要求的路径前缀。

嵌入请求每批最多发送 20 条文本。响应会按索引恢复输入顺序；空项、重复或越界索引、空向量、数量错误以及批次之间的维度不一致都会被拒绝，避免写入错误的向量缓存。更换模型名称后会生成对应模型的新缓存。识图备用接口只处理识图失败，不参与助手的嵌入请求。

## 密钥文件

可用现有的 `EBKCFP_` 环境变量从文件读取密钥，例如：

```dotenv
EBKCFP_LLM_IMAGE_RECOGNITION_OPENAI_API_KEY=/run/ezbookkeeping-secrets/burn_api_key
EBKCFP_LLM_IMAGE_RECOGNITION_FALLBACK_OPENAI_API_KEY=/run/ezbookkeeping-secrets/chat_api_key
EBKCFP_LLM_ASSISTANT_OPENAI_API_KEY=/run/ezbookkeeping-secrets/chat_api_key
EBKCFP_LLM_ASSISTANT_FALLBACK_OPENAI_API_KEY=/run/ezbookkeeping-secrets/chat_api_key
EBKCFP_LLM_TEXT_RECOGNITION_OPENAI_API_KEY=/run/ezbookkeeping-secrets/chat_api_key
EBKCFP_LLM_ASSISTANT_EMBEDDING_OPENAI_API_KEY=/run/ezbookkeeping-secrets/embedding_api_key
```

容器必须挂载这些文件，且运行用户需要读取权限。建议密钥文件由容器用户拥有、权限为 `0600`，以只读方式挂载。独立嵌入密钥与其他模型密钥一样，在启动配置日志中被隐藏。不要把实际密钥提交到仓库。

## 验证

AI 自动测试不调用真实模型供应商。使用与 CI 相同的设置可跳过无关的实时汇率接口测试：

```sh
go vet ./...
BUILD_PIPELINE=1 go test ./...
go test -race ./pkg/settings ./pkg/api ./cmd
```

真实部署验证需要显式启用，并提供配置文件、包含 `templates/` 的工作目录和测试账户 UID：

```sh
EBK_AI_DEPLOYMENT_CHECK=1 \
EBK_AI_DEPLOYMENT_CONFIG=/absolute/path/to/ezbookkeeping.ini \
EBK_AI_DEPLOYMENT_WORK_DIR=/absolute/path/to/ezbookkeeping \
EBK_AI_DEPLOYMENT_UID=<test-user-uid> \
go test -count=1 -timeout=12m -run '^TestAIDeploymentIntegration$' ./pkg/api
```

调用环境还需要相应的密钥环境变量及文件。该测试使用仓库内的合成收据，检查金额 `12.34`，并验证助手完成嵌入检索和流式回复。它不创建账单，但会读取测试账户的账单、调用所配置的供应商，并更新该账户的嵌入缓存。仅验证识图可将 `-run` 改为 `'^TestAIDeploymentIntegration$/receipt$'`。
