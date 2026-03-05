# Task 11 - Watcher como produtor de job com debounce

## Objetivo
Mudar watcher para apenas gerar jobs `fs_event` agregados, sem executar pipeline completa.

## Contexto atual
- `watcher.go` detecta mudanca e enfileira `ScanFiles` global.

## Escopo
- Refatorar watcher para:
  - agregar eventos em janela (ex.: 500ms-2s)
  - deduplicar por path/escopo
  - criar jobs `fs_event` com prioridade `normal`
- Tratar eventos de delete/rename com escopo correto.
- Garantir protecao contra tempestade de jobs redundantes.

## Arquivos alvo
- `backend/internal/worker/watcher.go`
- componentes de orchestrator/scheduler

## Criterios de pronto
- Watcher nao dispara pipeline completa.
- Rajadas de eventos geram quantidade controlada de jobs.
- Jobs gerados sao rastreaveis via API de jobs.

## Dependencias
- Tasks 4-5 e 9-10.
