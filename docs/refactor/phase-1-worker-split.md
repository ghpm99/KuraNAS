# Fase 1 — Quebrar `internal/worker` em sub-pacotes

**Risco:** médio · **Pré-requisito:** Fase 0 · **Importadores externos:** só `internal/app/app.go` e `internal/app/context_run_test.go`

## Objetivo

Separar o pacote único `worker` (38 arquivos, tudo no mesmo nível) em sub-pacotes por responsabilidade, para que abrir a pasta diga o que é cada coisa.

## ⚠️ Achados do recon (2026-06-09) — o plano original foi revisado

Ao inspecionar o acoplamento real, **dois pontos do plano inicial não se sustentam**:

1. **São 3 pacotes, não 4.** `engine/` e `steps/` separados **geram ciclo de import**:
   - `WorkerContext` (struct central em `worker.go`) referencia `*JobScheduler` e `*JobOrchestrator` (tipos do engine).
   - **Todos** os step executors recebem `*WorkerContext` (→ dependem do engine).
   - Mas `buildStepExecutors` (engine) referencia esses steps (→ engine depende deles).
   - Resultado: `engine ↔ steps` é mutuamente dependente. Separá-los exigiria um refactor de interfaces (steps deixarem de receber `WorkerContext` e passarem a receber dependências estreitas) — fora do escopo de uma fase de **organização**. Portanto os steps ficam **dentro de `engine/`** como arquivos `step_*.go` (agrupamento visual por nome, não por pacote).

2. **A consolidação do watcher sai do escopo.** `internal/worker/watcher.go` (`startEntryPointWatcher`, observa o `ENTRY_POINT`) é **acoplado ao `WorkerContext`**; `internal/watcher/` é o **FolderWatcher** (observa watch-folders do usuário, move para bibliotecas). **Não são duplicados.** Mesclá-los arrastaria `WorkerContext` para dentro de `internal/watcher/`. Então o entry-point watcher **fica em `engine/`** e `internal/watcher/` permanece intocado.

### Verificações que embasam isso

- `worker/scan/` (pipeline) **não usa `WorkerContext` nem tipos de job** → pacote limpo. Depende de files/video/ai/logger/utils.
- `engine` chama funções de `scan` (`StartFileProcessingPipeline`, `getMetadata`, `getCheckSum`, `CreateThumbnailWorker`, `GenerateVideoPlaylistsWorker`, `UpdateFileRecord`, `createFileRecord`, `ScanDirWorker`) → direção única `engine → scan`.
- Extrair `job/` implica prefixar **~278 referências** a tipos/constantes de job (`JobType*`, `StepType*`, `JobPriority*`, `JobStatus*`, `StepStatus*`, `JobScope`, `PlannedJob/Step`) com `job.` em **13 arquivos**. Mecânico e guiado pelo compilador, mas é volume. (As structs genéricas `Job`/`Step` têm 0 usos externos → sem risco de replace ambíguo.)
- Extrair `scan/` implica **exportar 4 símbolos** que o engine usa hoje sem qualificar: `getMetadata`→`GetMetadata`, `getCheckSum`→`GetCheckSum`, `createFileRecord`→`CreateFileRecord`, `pythonScriptRunner`→`PythonScriptRunner` (+ `SetPythonScriptRunnerForTesting` já exportado). Os tipos `FileWalk`/`ResultWorkerData` permanecem internos ao `scan`.

## O que precisa ser feito (plano revisado: 3 pacotes)

### Sub-pacotes-alvo

| Sub-pacote | Recebe (fonte) | Depende de |
|---|---|---|
| `worker/job/` | `job_domain.go` (+`job_domain_test.go`) — enums/tipos de Job/Step. **Pacote folha, não importa ninguém.** | — |
| `worker/scan/` | pipeline de arquivo: `directory_walker.go`, `dir.go`, `file_checksum.go`, `file_metadata.go`, `file_database_persistence.go`, `file_dto_converter.go`, `file_processing_pipeline.go`, `file_result_monitor.go`, `files.go`, `video_playlist.go`, `thumbnail.go` (+ testes internos) | files, video, ai, logger, utils |
| `worker/engine/` | `worker.go`, `job_orchestrator.go`, `job_scheduler.go`, `step_executors.go`, e os steps acoplados ao `WorkerContext`: `checksum.go`, `ollama_step.go`, `takeout_step.go`, `takeout_step_payload.go`, `ai_playlist_cluster_step.go`, `watcher.go` (entry-point) (+ todos os testes do engine) | `job`, `scan` |

### Execução incremental (cada passo = 1 commit, CI verde)

A ordem importa para manter compilável a cada commit:

1. **`worker/scan/`** (maior ganho visual, contido): mover os 11 arquivos da pipeline; exportar os 4 símbolos; ajustar as ~10 chamadas no engine para `scan.X`. Deixa `job_domain` + engine no `worker` raiz por ora.
2. **`worker/job/`**: mover `job_domain.go`; prefixar as ~278 referências com `job.` nos arquivos do engine (compilador como rede de segurança).
3. **`worker/engine/`**: mover o restante para `engine/`; ajustar `internal/app/app.go` (`worker.StartWorkers`→`engine.StartWorkers`, `worker.WorkerContext`→`engine.WorkerContext`) e `context_run_test.go`.

> Pode-se parar após o passo 1 e já ter a maior parte do ganho. Os passos 2–3 são incrementos opcionais de polimento.

### Cuidados

- `worker/scan/` importa `internal/api/v1/files` (usa `files.FileModel`/conversões) — direção válida.
- Atenção ao nome: o novo pacote `worker/job` (singular) coexiste com `internal/api/v1/jobs` (plural, já importado como `jobs`). São distintos; conferir os imports em cada arquivo do engine.

## Resultado esperado

- `internal/worker/` com sub-pacotes coesos (`job/`, `scan/`, `engine/`); steps como `step_*.go` dentro de `engine/`.
- `internal/watcher/` intocado (decisão registrada: não consolidar).
- Comportamento idêntico (mesmos workers, mesmos jobs/steps).

## Critério de aceite

- `make ci-backend` verde a cada passo (testes de worker incluídos).
- `go vet ./...` sem ciclos.
- Subir o servidor e confirmar que workers iniciam (`ENABLE_WORKERS`) e o entry-point watcher roda.
