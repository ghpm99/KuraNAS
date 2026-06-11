# 06 — Watcher por eventos nativos (fsnotify) em vez de polling com snapshot integral

**Tipo:** performance/escalabilidade · **Prioridade:** P2

## Contexto

O watcher de entry point (`internal/worker/engine/watcher.go`) faz, **a cada 5 segundos**, um `filepath.WalkDir` da árvore inteira e mantém **dois mapas com todos os paths** (snapshot anterior e atual) em memória, comparando-os para detectar mudanças.

Custos para um NAS real:

- I/O de varredura completa a cada 5s — em centenas de milhares de arquivos isso é caro e, em disco mecânico, **impede o disco de hibernar** (relevante para NAS doméstico ligado 24/7).
- Memória proporcional ao número total de arquivos, dobrada (dois snapshots).
- O comentário do próprio `executeDiffAgainstDBStep` reconhece que carregar a árvore em memória "não escala além de algumas dezenas de milhares de arquivos" — e o watcher faz exatamente isso, 12 vezes por minuto.
- Latência de detecção de até 5s + debounce, contra detecção imediata com eventos nativos.

O alvo de produção é Windows, que tem `ReadDirectoryChangesW` (recursivo nativo); Linux tem inotify. A lib `fsnotify` abstrai ambos.

## Objetivo

Mudanças no filesystem são detectadas por eventos do SO, com custo de CPU/IO próximo de zero em repouso; a varredura completa passa a existir só como reconciliação (boot + periódica de baixa frequência).

## O que fazer

1. Substituir o loop de snapshot por um watcher `fsnotify` sobre o `ENTRY_POINT`.
2. Manter o pipeline de despacho atual (jobs `fs_event`/`mark_deleted` por path; fallback para scan completo em rajadas).
3. Rebaixar a varredura completa para reconciliação: no boot (`startup_scan`, já existe) e opcionalmente periódica (ex.: diária).

## Como fazer

- **Dependência**: `github.com/fsnotify/fsnotify`. Atenção: o backend cross-compila para Windows com CGO; fsnotify é Go puro, sem impacto no build.
- **Recursividade**: inotify (Linux) **não é recursivo** — é preciso adicionar watch por diretório e adicionar/remover watches conforme pastas são criadas/removidas (manter um registro dos watches ativos). No Windows, `fsnotify` usa `ReadDirectoryChangesW` por diretório igualmente; a recursividade manual cobre os dois. Encapsular isso num componente `recursiveWatcher` testável.
- **Tradução de eventos**: `Create`/`Write` → reaproveitar `buildFileProcessingPlan` (arquivo) ou persistência de diretório (task 01); `Remove`/`Rename` (lado origem) → job `mark_deleted` no path. Coalescer eventos repetidos do mesmo path numa janela curta (o debounce corrigido da task 03 serve de base).
- **Rajadas e overflow**: filas de eventos do SO estouram (inotify queue overflow, buffer do RDCW). Ao detectar overflow/erro do watcher, cair para `enqueueFilesystemEventJob` (scan completo) — mesmo fallback que já existe para >50 mudanças.
- **Reconciliação**: manter `startup_scan` no boot; agendar um `fs_event` de árvore completa em frequência baixa e configurável (default sugerido: 24h) para capturar qualquer evento perdido.
- **Remoção**: apagar `collectEntryPointSnapshot`/`snapshotDiffPaths` e o ticker de 5s quando o novo watcher estiver coberto por testes.
- **Testes**: o `recursiveWatcher` com diretório temporário (criar/alterar/remover arquivo e pasta aninhada); simulação de overflow disparando o fallback. Validar manualmente no Windows (build de produção) antes de concluir.

## Critérios de aceite

- [x] Criar/alterar/remover arquivo ou pasta sob o `ENTRY_POINT` gera os jobs corretos em menos de 2s, sem varredura completa. *(eventos nativos chegam imediatamente — `TestRecursiveWatcherDetectsCreateWriteAndRemove` usa timeout de 2s; despacho no flush de 500ms; `TestDispatchWatcherChangesResolvesAgainstDisk`)*
- [x] Em repouso (nenhuma mudança), o processo não faz walk da árvore (verificável por log/profile) — disco pode hibernar. *(`collectEntryPointSnapshot`/ticker de 5s removidos; os únicos `WalkDir` restantes são o registro inicial de watches no boot e o de pastas recém-criadas; o flush de 500ms só consulta um mapa em memória)*
- [x] Pastas criadas após o boot passam a ser monitoradas (watch recursivo dinâmico). *(`watchTree` no Create de diretório, emitindo conteúdo que correu na frente do watch; `TestRecursiveWatcherWatchesDirectoriesCreatedAfterStart`)*
- [x] Overflow/erro do watcher dispara reconciliação completa automaticamente. *(`onError` → `enqueueFilesystemEventJob`; `TestWatcherErrorFallbackEnqueuesFullReconciliation`)*
- [x] Reconciliação periódica configurável existe e roda. *(`reconciliationLoop` + `WATCHER_RECONCILE_HOURS`, default 24h; `TestReconcileIntervalIsConfigurableWithSaneDefault`)*
- [ ] Funciona no build Windows de produção (validação manual documentada na task).
- [ ] `make ci-backend` verde (cobertura ≥ 80%).

## Fora de escopo

- O watcher de **watch folders** (`internal/watcher/`, polling de 60s para auto-organização) — segue como está; unificação é decisão futura.
- Indexação de diretórios (task 01) e correção do debounce (task 03) — pré-requisitos, não parte desta task.
- Múltiplas raízes (task 10).
