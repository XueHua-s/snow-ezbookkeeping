# Isolated Docker application smoke and E2E

Executed 2026-10-15 (Asia/Shanghai) after the final backend/frontend gates. The full application ran inside Docker, including the Linux backend, built frontend, templates, actual HTTP authentication, SQLite and seven-tool runner. Its model endpoint was a local deterministic protocol stub, **not a real model**. This supplements [harness verification](ai-assistant-harness-verification.md); it does not close the real-primary/fallback seven-question release gate.

## Build and isolation

- Docker context: Colima, Linux/arm64. Exact project backend base `golang:1.27.1-alpine3.24`; `go version` inside the builder reported `go1.27.1 linux/arm64`. CGO enabled with Alpine GCC/musl. Build exited 0.
- Runtime base: `alpine:3.24.2`; project `docker/docker-entrypoint.sh`, UID/GID 1000:1000, actual `templates/`, final `dist` as `/ezbookkeeping/public`. Startup/schema synchronization succeeded. Image build exited 0.
- Validation image `ebk-harness-app-smoke:20261015`, image ID `sha256:3f7e65e9dd60d117cb6c71947f4a983dc8f9299bb11446d0ac1945306b1910c3`.
- This was a validation image using the already-tested final frontend artifact, not a rerun of the stock Dockerfile's frontend build stage. Python was included solely for the isolated stub and data inspection; no release Dockerfile or application source was changed.
- Backend source/templates/module files and `dist` were copied into a task-created directory and byte-compared to the final checkout. Sorted path+NUL+content SHA-256: backend `85b6c6c02c4bb9b314e771bc0f4c0e3670c53e771c45e32907676eaf21f8889b`; frontend `4a1fc13d3fb13cd190b72c66b167524b5b9e16daf70e575979b70a758e43b897`.
- The cached untagged Go image was inspected first and reported Go 1.26.1 with `GOTOOLCHAIN=local`; it was not used to compile this Go 1.27.1 project. The exact required base was available. An initial bind from macOS `/var/folders` appeared empty in Colima; the isolated source copy was moved under the shared `~/Library/Caches` directory and the build rerun successfully.
- Builder mounted only the task source copy read/write and existing Go dependency cache read-only. Runtime containers have **no host bind mounts**. Synthetic SQLite/config were copied into the image; no real-user database, credentials, production configuration or storage was mounted.

Fixture: one dedicated synthetic user, CNY account 100 with balance 170000 minor units, category 200, transactions 300/301/302 with amounts 12345/12346/12347 and first amount hidden; one legacy cache sentinel. SQLite online backup copied only this earlier isolated fixture. Network `ebk-harness-app-network` contains app `ebk-harness-docker-app` and stub `ebk-harness-docker-mock`. Published ports were loopback-only app `127.0.0.1:32873→8080`, stub counters `127.0.0.1:32874→32871`.

## Executed HTTP results

A Python standard-library client performed real `/api/authorize.json` login and kept the token only in memory. The following assertions all passed, process exit 0:

| Check | Actual outcome |
| --- | --- |
| Authentication boundary | Real dedicated-user login succeeded; unauthenticated assistant request rejected. |
| Frontend serving | Both desktop and mobile application documents served successfully from the container's final `public` artifact. |
| JSON seven-tool turn | Actual successful evidence contained exactly `list_accounts`, `list_categories`, `list_tags`, `query_transactions`, `get_transaction`, `aggregate_transactions`, `get_account_balances`; no replacement query handlers. |
| SSE seven-tool turn | Complete reply and exactly one done terminal; actual service/runner telemetry recorded seven executions and two model rounds. |
| Explicit continuation | First transaction reference 300; authenticated continuation body plus explicit next-page question returned reference 301. Raw cursor was not printed or put into a URL. |
| Upstream failure | Stub 503 produced JSON failure and SSE `event:error` with the existing error shell and `context.assistant.state=failed`; no done terminal. |
| Real cancellation | After the stub observed an active slow upstream request, the HTTP client closed the SSE socket. The stub's socket poll observed the upstream TCP connection close and incremented its cancellation counter. |
| Resource reuse | A subsequent authenticated assistant query succeeded after cancellation. |
| Read-only ledger | After requests: exactly 3 transactions, signed amount sum 37038 minor units, balance 170000, one retained legacy sentinel. |
| Runtime identity | `docker exec ... id` returned UID/GID 1000:1000. |
| Observation redaction | Actual seven-tool JSON trace had null first-status time; SSE trace had written-status time; canceled trace present. Synthetic API key/signing secret, bill-comment canary, complete model reply and Bearer prefix were absent from captured app logs. |

The fixture values are CNY expense 370.38 and current balance 1700.00; this fixture is separate from the six-record 890.00 net-income oracle. The stub chooses tool calls and emits predetermined prose. A successful Docker tool trace does not demonstrate real-model reasoning or numerical explanation correctness.

The first draft of the HTTP smoke incorrectly expected `type:error` inside SSE data; the actual protocol intentionally uses `event:error` and the preserved error JSON shell. The assertion was corrected to check that shell and failed assistant state. No production change was needed.

## Browser results

Actual browser interaction against the Docker application on port 32873 passed on **both desktop and mobile routes**:

- Completed query showed effective range, Asia/Shanghai timezone, current-page coverage and query evidence.
- Hidden 123.45 was absent from the rendered amount and displayed as `***`; explicit next page displayed 123.46.
- A slow request was canceled through the UI, producing canceled state and removing continuation.
- Stub 503 produced localized request-failed state, without a completed answer.
- Clear removed prior conversation; summary then completed and displayed the localized user intent.
- Reference navigation opened the desktop authenticated transaction-detail dialog and the actual mobile `/mobile#!/transaction/detail?id=300` route with masked amount; mobile assistant session cleared on leave. Both reference flows recorded zero unhandled rejections.
- Unsafe `javascript:` links counted zero. Temporary unhandled-rejection observers recorded zero assistant-operation failures.

The coordinator captured and visually inspected desktop/mobile screenshots alongside the interaction checks; no visual defect was found. A prior expected 503 toast was transiently visible in a screenshot. These are actual Docker application routes in a desktop embedded browser, not physical-phone certification. The earlier host-process GUI checks are separate evidence. Initial automation selectors for the stop control/current Framework7 page were corrected; no application source change was required. All requested local Docker HTTP/browser E2E checks passed.

## Reproduce the application image

Use only an isolated synthetic fixture. Run from the repository root after the documented `npm run build` gate. The source/config/data copy must be under a directory shared with the Docker daemon (Colima shared `~/Library/Caches` in this run):

```sh
TASK_DOCKER_DIR="$(mktemp -d "$HOME/Library/Caches/ebk-harness-docker-XXXXXX")"
cp go.mod go.sum ezbookkeeping.go LICENSE "$TASK_DOCKER_DIR/"
cp -R pkg cmd templates conf dist docker "$TASK_DOCKER_DIR/"
mkdir -p "$TASK_DOCKER_DIR/data" "$TASK_DOCKER_DIR/storage" "$TASK_DOCKER_DIR/log"

docker run --rm --name ebk-harness-linux-builder \
  -e BUILD_PIPELINE=1 -e GOMODCACHE=/gomod \
  -v "$HOME/go/pkg/mod:/gomod:ro" \
  -v "$TASK_DOCKER_DIR:/work" -w /work \
  golang:1.27.1-alpine3.24 sh -c \
  'test -f go.mod && apk add --no-cache gcc musl-dev git && go version && CGO_ENABLED=1 go build -o /work/ezbookkeeping ./ezbookkeeping.go'
```

The read-only module-cache version assumes dependencies were already downloaded by local tests. On a fresh environment, use a newly created task-only writable cache instead; do not downgrade the declared Go version. Do not copy a production database. For a new fixture, start with an empty `data/test.db`, allow application schema initialization, register a dedicated test user and create synthetic accounts/transactions through the normal UI. For repeatable numeric/ID assertions, use the existing temporary test fixture or an explicitly synthetic seed with the values listed above. A copied database should use SQLite's backup API rather than copying a live WAL database file alone.

Create a test-only `test.ini` from `conf/ezbookkeeping.ini` and set these values (secret/model names below are placeholders, not real credentials):

```ini
[server]
http_addr = 0.0.0.0
http_port = 8080
root_url = http://127.0.0.1:32873/
static_root_path = /ezbookkeeping/public
[database]
type = sqlite3
db_path = /ezbookkeeping/data/test.db
auto_update_database = true
[storage]
local_filesystem_path = /ezbookkeeping/storage
[security]
secret_key = <fresh-isolated-test-signing-secret>
[ai]
enable_ai_assistant = true
[llm_assistant]
llm_provider = openai
openai_base_url = http://ebk-harness-docker-mock:32871/v1/
openai_api_key = <synthetic-test-only-value>
openai_model_id = local-protocol-stub
proxy = none
```

The stub must bind `0.0.0.0:32871` inside its container. For a complete tool turn, respond to the first Chat Completions request with a `tool_calls` message and `finish_reason:"tool_calls"`; after matching tool results arrive, respond with final text and `finish_reason:"stop"`. This run selected all seven names for `all seven`, one-row `query_transactions` for default questions, and used the verified continuation candidate for explicit paging. `fail` returned HTTP 503. `slow` delayed while polling the provider socket for EOF; a GET stats endpoint returned only request/active/canceled/failure counters, never prompts, tokens, cursors or results.

A runnable test-only `mock.py` for that behavior is included below; save it in the task build directory. It intentionally supports only the Chat protocol used in this Docker smoke. Other protocols were exercised by the separate protocol suite.

```python
"""Isolated GUI smoke upstream; deterministic protocol stub, not a real model."""
import json,time,select,socket,threading
stats={"requests":0,"slow_active":0,"canceled":0,"failures":0};lock=threading.Lock()
from http.server import ThreadingHTTPServer,BaseHTTPRequestHandler
class Handler(BaseHTTPRequestHandler):
 def log_message(self,*args): pass
 def do_GET(self):
  body=json.dumps(stats).encode();self.send_response(200);self.end_headers();self.wfile.write(body)
 def do_POST(self):
  with lock: stats["requests"]+=1
  body=json.loads(self.rfile.read(int(self.headers['Content-Length'])))
  messages=body.get('messages',[])
  last_user=max((i for i,m in enumerate(messages) if m.get('role')=='user'),default=-1)
  question=messages[last_user].get('content','') if last_user>=0 else ''
  if 'slow' in question.lower():
   with lock: stats['slow_active']+=1
   try:
    for _ in range(100):
     readable,_,_=select.select([self.connection],[],[],0.05)
     if readable and self.connection.recv(1,socket.MSG_PEEK)==b'':
      with lock: stats['canceled']+=1
      return
   finally:
    with lock: stats['slow_active']-=1
  if 'fail' in question.lower():
   with lock: stats['failures']+=1
   self.send_response(503);self.end_headers();return
  results=[m for m in messages[last_user+1:] if m.get('role')=='tool']
  if not results:
   args={'page_size':1}
   for m in messages:
    content=m.get('content','')
    if m.get('role')=='system' and 'Verified continuation candidate' in content:
     try:
      candidate=json.loads(content.split('never reveal raw cursor):\n',1)[1])['candidate'];args={'cursor':candidate['cursor']}
     except Exception: pass
   message={'role':'assistant','content':None,'tool_calls':[{'id':'gui_query','type':'function','function':{'name':'query_transactions','arguments':json.dumps(args)}}]};reason='tool_calls'
  else:
   message={'role':'assistant','content':'这是隔离测试账本的查询结果。此处使用本地协议模拟器验证界面；请查看下方查询范围、完整性及账单引用。\n\n[安全链接](https://example.test) [恶意链接](javascript:alert(1)) <script>window.__harnessInjected=true</script>'};reason='stop'
  if not results and 'all seven' in question.lower():
   calls=[('list_accounts',{}),('list_categories',{}),('list_tags',{}),('query_transactions',{'page_size':1}),('get_transaction',{'transaction_id':'300'}),('aggregate_transactions',{}),('get_account_balances',{})]
   message={'role':'assistant','content':None,'tool_calls':[{'id':'docker_'+name,'type':'function','function':{'name':name,'arguments':json.dumps(args)}} for name,args in calls]};reason='tool_calls'
  data=json.dumps({'id':'gui-mock','object':'chat.completion','choices':[{'index':0,'message':message,'finish_reason':reason}]}).encode()
  try:
   self.send_response(200);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(data)));self.end_headers();self.wfile.write(data)
  except (BrokenPipeError,ConnectionResetError):pass
ThreadingHTTPServer(('0.0.0.0',32871),Handler).serve_forever()
```

The validation Dockerfile used:

```dockerfile
FROM alpine:3.24.2
RUN apk add --no-cache ca-certificates tzdata python3 \
 && addgroup -S -g 1000 ezbookkeeping \
 && adduser -S -G ezbookkeeping -u 1000 ezbookkeeping
WORKDIR /ezbookkeeping
COPY --chown=1000:1000 ezbookkeeping ./ezbookkeeping
COPY --chown=1000:1000 dist ./public
COPY --chown=1000:1000 templates ./templates
COPY --chown=1000:1000 conf ./conf
COPY --chown=1000:1000 data ./data
COPY --chown=1000:1000 storage ./storage
COPY --chown=1000:1000 log ./log
COPY --chown=1000:1000 test.ini ./test.ini
COPY --chown=1000:1000 mock.py ./mock.py
COPY docker/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
USER 1000:1000
ENV EBK_CONF_PATH=/ezbookkeeping/test.ini
EXPOSE 8080
ENTRYPOINT ["/docker-entrypoint.sh"]
```

Build and start only task-named resources:

```sh
docker build -f "$TASK_DOCKER_DIR/Dockerfile.validation" \
  -t ebk-harness-app-smoke:20261015 "$TASK_DOCKER_DIR"
docker network create ebk-harness-app-network
docker run -d --name ebk-harness-docker-mock --network ebk-harness-app-network \
  -p 127.0.0.1:32874:32871 ebk-harness-app-smoke:20261015 python3 /ezbookkeeping/mock.py
docker run -d --name ebk-harness-docker-app --network ebk-harness-app-network \
  -p 127.0.0.1:32873:8080 ebk-harness-app-smoke:20261015
```

Then log in at `/desktop` and `/mobile` using the dedicated fixture user. HTTP routes are `/api/authorize.json`, authenticated `/api/v1/llm/assistant/chat.json` and `/api/v1/llm/assistant/chat/stream.json`; supply timezone headers consistently, store the returned token in memory, and send continuation only in authenticated JSON bodies. Assert both terminal semantics and actual query evidence, not only HTTP 200.

## Cleanup and limitations

After browser/HTTP evidence and screenshot inspection, the following cleanup was actually executed successfully. Both task source/config/data copies were also removed; unrelated resources (including `codex-s03-build`) and shared caches were preserved. Repeated runs should remove only their own task-created resources:

```sh
docker rm -f ebk-harness-docker-app ebk-harness-docker-mock
docker network rm ebk-harness-app-network
docker image rm ebk-harness-app-smoke:20261015
```

Remove the task source/config/synthetic-data directory after confirming it is the newly created `ebk-harness-docker-*` directory. Do not prune shared images/caches or remove unrelated containers. Builder/probe containers used `--rm`. Official Go/Alpine images pulled for this run may remain cached.

This establishes application-container compatibility and deterministic local E2E behavior. It does not certify the stock multi-stage Dockerfile's frontend rebuild, amd64 deployment, physical mobile hardware, real model availability, real model financial answers, or production rollout.
