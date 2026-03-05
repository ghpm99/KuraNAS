## Diagnóstico do sistema atual

### 1) Worker “monolítico” virou _orquestrador + executor_

O `StartFileProcessingPipeline` faz:

- descobrir arquivos
- decidir o que fazer
- executar etapas
- coordenar dependências
- reportar progresso

Isso mistura **4 responsabilidades** e impede:

- reuso das etapas em outros fluxos (upload, watcher)
- paralelismo real (porque o “cérebro” tá preso no mesmo lugar)
- observabilidade fina (só dá pra reportar progresso do monolito)

### 2) Watcher dispara “tudo” porque não existe um “plano” incremental

O `startEntryPointWatcher` hoje é praticamente “evento → roda pipeline”.
O certo é “evento → cria _job_ do tipo certo com _escopo_ certo”, e só então os workers executam.

### 3) Checksum e processamento “duplicados” em vários lugares

Você já percebeu isso. O problema não é só duplicação — é **inconsistência**:

- em um fluxo calcula checksum “de um jeito”
- em outro fluxo “de outro jeito”
- e fica impossível saber se o estado no banco representa o arquivo real

### 4) Progresso depende do desenho da pipeline, não do estado real

Se o progresso é “um canal alimentado pela pipeline”, você fica refém do _código que está rodando agora_.
O ideal é progresso ser **estado da execução** (persistido) e não “prints em tempo real”.

---

## Princípio que vai destravar tudo: separar “Orquestração” de “Execução”

Você quer workers pequenos. Ótimo. Mas **alguém precisa montar o plano**.

### Componentes (papéis claros)

1. **Job Orchestrator (Planner)**

- transforma eventos/requests em um **Job** com **Steps** (grafo/DAG)
- define dependências: `persist` depende de `metadata + checksum`, etc.
- define escopo (arquivo X, pasta Y, scan full)
- define prioridade e modo (foreground/background)

2. **Worker Pool (Executores)**

- executam _uma coisa bem definida_ (metadata, checksum, thumbnail…)
- idempotentes (rodar duas vezes não quebra)
- reportam progresso no estado do Step

3. **Job Store / State Store**

- persiste Job/Steps/Status/erros
- permite UI consultar depois (e não “perder” status se reiniciar)

4. **Progress API (poll) + opcional WebSocket/SSE**

- UI pode acompanhar com polling
- e se quiser experiência “live”, adiciona WS/SSE como melhoria, não como requisito

Essa separação resolve seu dilema: “pipeline precisa alimentar canal de progresso”.
Ela não precisa. O progresso vira **estado do job**, e qualquer fluxo (startup, watcher, upload) usa o mesmo mecanismo.

---

## Modelo de dados mínimo (pra suportar progresso, retry, UI e reinício)

### Job

- `id`
- `type`: `startup_scan`, `upload_process`, `fs_event`, `reindex_folder`
- `scope`: `{file_id?, path?, root?}`
- `priority`: `low|normal|high`
- `status`: `queued|running|partial_fail|failed|completed|canceled`
- `created_at`, `started_at`, `ended_at`

### Step

- `id`, `job_id`
- `type`: `scan`, `stat`, `diff`, `metadata`, `checksum`, `persist`, `thumbnail`, `video_playlist_index`, etc.
- `depends_on`: lista
- `status`: `queued|running|failed|completed|skipped`
- `progress`: `0..100` (ou contadores)
- `attempts`, `last_error`, `started_at`, `ended_at`
- `metrics`: duração, bytes processados, etc.

**Isso é o suficiente** pra UI acompanhar, pra você debugar, e pra você retomar jobs após restart.

---

## Como quebrar o StartFileProcessingPipeline sem perder funcionalidade

Hoje ele faz tudo. A refatoração correta é:

### 1) Startup scan vira Job do tipo `startup_scan`

Steps típicos:

1. `scan_filesystem(root)` → produz lista de paths + stat
2. `diff_against_db(snapshot)` → classifica `new/modified/deleted/unchanged`
3. Para cada arquivo em `new/modified`: enfileira sub-job(s) `process_file(file_id/path)`
   (ou um step “fan-out” interno)
4. Para deletados: `mark_deleted_in_db` (ou step específico)
5. (Opcional) consolidadores: `rebuild_indexes`, `rebuild_playlists_if_needed`

O pulo do gato é o step 2: **diff barato**.

---

## Detecção de alteração: o que você deve fazer na prática (barato → caro)

Você comentou “se não alterou, não joga na pipeline”. Isso é crucial.

### Estratégia recomendada (rápida e confiável)

1. **Stat key** (barato)

- tamanho
- `mtime`
- inode (se aplicável)
- caminho

2. Se stat mudou → **fingerprint leve**

- hash parcial (ex.: primeiros/últimos N KB) OU
- checksum completo se você exigir precisão total

3. Só então dispara processamento pesado (thumbnail, metadata profunda, etc.)

No banco, mantenha:

- `size`, `mtime`, `stat_hash` (um hash do conjunto), `checksum` (opcional), `processed_at`

Isso reduz seu “startup scan” de minutos pra “varrer e comparar”, e só processa o que precisa.

---

## Upload: como retornar “sucesso” rápido sem perder consistência

Seu objetivo: request só salva arquivo e dispara processamentos em paralelo.

### Padrão recomendado

- Upload salva o arquivo
- Cria um `Job upload_process` com steps:

  - `stat+diff` (pra saber se substituiu algo)
  - `metadata`
  - `checksum`
  - `persist`
  - `thumbnail` (condicional)
  - `playlist_index` (condicional)

Retorno da API:

- `201 Created`
- `job_id` + `file_id`
- UI pode acompanhar pelo `/jobs/:id` e mostrar “processando…”

**Importante:** não rode “paralelo solto” com goroutine por tarefa sem controle.
Você quer:

- fila com prioridade
- limite de concorrência por tipo (thumbnail costuma ser pesado, checksum também)
- backpressure

---

## Watcher: o que faz sentido migrar pra worker e o que não faz

### O watcher deve ser “burro”

Ele só:

- traduz evento do FS em `Job fs_event`
- agrega/debounce (muito importante)
- enfileira

**O que NÃO deve ficar no watcher:**

- decidir pipeline inteira
- calcular checksum
- persistir
- gerar thumbnail
- qualquer lógica “de produto”

### O que faz sentido adicionar ao sistema de workers (e tirar de camadas soltas)

- **Debounce/Aggregation worker**: “recebi 200 eventos em 2s, agrupa por pasta/arquivo e cria 1 job”
- **Diff worker**: valida “mudou mesmo?” (stat+fingerprint)
- **Reconciliação**: jobs periódicos (ex.: “reconcile thumbnails missing”, “reconcile checksums missing”)

---

## Reavaliando seus workers atuais (o que fazer com cada um)

### ScanDirWorker (descontinuado)

- Remover (ou transformar em step `scan_filesystem` reutilizado por startup e “scan folder” sob demanda).

### UpdateCheckSumWorker

Você pode manter, mas com uma mudança de conceito:

- ele não é “um worker solto”
- ele vira step `checksum(file_id)` dentro do Job system

E vira **a única forma** de atualizar checksum (centralização real).

### CreateThumbnailWorker

Perfeito como executor especializado, mas:

- precisa ser **idempotente** (“thumbnail já existe e está atualizada?” → skip)
- precisa ter “limite de concorrência” (FFmpeg/imagick podem matar CPU)

### GenerateVideoPlaylistsWorker

Aqui eu sugiro separar em dois:

- `video_catalog_index` (indexa metadados e estrutura de série/pasta)
- `playlist_builder` (gera playlists “inteligentes” a partir do index)

E rodar isso:

- ao final de processamento de novos vídeos (batch/debounce)
- e também como job periódico/repair (“rebuild playlists”)

### StartFileProcessingPipeline

Deixa de existir como “worker executor”.
Vira um **Job type** que o Orchestrator monta.

---

## Monitoramento de progresso: como resolver sem canal acoplado

Você já viu que “canal de progresso” acopla tudo.

### Sugestão prática

- Cada step escreve progresso no Job Store:

  - `processed_files`, `total_files`
  - ou `bytes_processed`
  - `stage`
  - erros acumulados

A UI consulta:

- `GET /jobs/:id`
- `GET /jobs?status=running`
- `GET /jobs/:id/steps`

Se quiser “ao vivo”, adiciona:

- SSE/WS que só “streama mudanças de estado”
  Mas o **estado é a fonte da verdade**, não o stream.

---

## Regras operacionais que você precisa (pra não virar caos)

### 1) Idempotência

Rodar `thumbnail(file_id)` duas vezes:

- não pode duplicar registros
- não pode corromper
- deve virar “skip” se já está ok

### 2) Retries com categoria de erro

- erro transitório (arquivo em uso, timeout ffmpeg) → retry com backoff
- erro permanente (formato inválido) → falha marcada e segue

### 3) Priorização

- Upload do usuário: **high**
- Watcher: **normal**
- Startup scan: **low** (mas pode subir prioridade em “primeiro boot”)

### 4) Limites de concorrência por tipo

Ex:

- checksum: N
- thumbnail: M (menor)
- metadata: K

Isso evita CPU 100% eterno.

### 5) Debounce no watcher

Eventos do FS vêm em rajadas.
Agrupe por:

- path
- janela de tempo (ex.: 500ms–2s)
- tipo de evento (rename é especial)

---

## Planejamento de implementação (incremental, sem quebrar o core)

### Fase 1 — Fundar o “Job System” (sem mudar comportamento ainda)

- Criar `JobStore` (DB) + modelos `Job` e `Step`
- Criar `Queue` + `WorkerPool` genérico (por tipo)
- Criar endpoints de observabilidade `/jobs`
- Adaptar 1 worker simples (ex.: thumbnail) pra rodar como Step

**Entrega de valor imediata:** você passa a ter rastreio e debug de execução.

### Fase 2 — Centralizar checksum e thumbnail como Steps

- Substituir chamadas diretas na pipeline por `enqueue step`
- Implementar “skip se up-to-date”
- Implementar limites de concorrência

**Entrega:** elimina duplicação e reduz custo.

### Fase 3 — Refatorar Upload para “save + enqueue job”

- Upload retorna `job_id`
- UI passa a acompanhar progresso via polling

**Entrega:** melhora UX e tempo de resposta.

### Fase 4 — Transformar StartFileProcessingPipeline em Job type

- Implementar steps scan + diff
- Fan-out para process_file jobs
- Colocar startup scan com prioridade low e métricas

**Entrega:** seu core vira modular e escalável.

### Fase 5 — Watcher vira produtor de jobs (com debounce)

- Watcher não executa pipeline
- Só cria jobs `fs_event` e pronto

**Entrega:** elimina cascata de processamento.

### Fase 6 — Video playlists: index + builder com batch

- Index como step por vídeo (ou por pasta)
- Builder como job batch (debounced)

---

## Pontos que hoje estão fora do worker e deveriam entrar (na sua lógica atual)

Pelo que você descreveu, itens típicos que valem migrar:

1. **Decisão de “precisa processar?”**
   Hoje parece espalhado. Deve virar step `diff`/`needs_processing`.

2. **Consolidação pós-processamento** (ex.: gerar playlists ao final)
   Deve virar **job batch** disparado por gatilho (“novos vídeos chegaram”) com debounce.

3. **Repair / reconciliation**
   Coisas que inevitavelmente ficam inconsistentes (thumb faltando, checksum nulo) devem ser jobs periódicos, não “jeitinho” em endpoints.

4. **Limpeza e marcação de deletados**
   Watcher detecta delete → job `mark_deleted` com consistência no DB.

---

## Principais riscos e como evitar

### Risco 1: “Vou criar um sistema de jobs gigante”

Evita com:

- modelo simples de Job/Step
- poucos tipos de step no começo
- sem features distribuídas ainda (sem Kafka/Redis no início)

### Risco 2: “Jobs duplicados pra mesma coisa”

Evita com:

- chave de dedupe (ex.: `step_type + file_id + file_version`)
- e/ou lock otimista (versionamento por `mtime/size`)

### Risco 3: “Progresso vira mentiroso”

Evita com:

- progresso baseado em contadores reais (total vs processed)
- e step status confiável

---

## Resultado final (o que você ganha)

- Startup scan deixa de “processar tudo”: vira scan+diff e só processa mudanças
- Upload vira rápido (request não bloqueia)
- Watcher para de explodir pipeline
- Workers ficam pequenos, testáveis e reaproveitáveis
- Progresso vira consistente, consultável e persistido (não depende de canal)
- Você cria base para features futuras (dashboard de jobs, cancelamento, prioridades)
