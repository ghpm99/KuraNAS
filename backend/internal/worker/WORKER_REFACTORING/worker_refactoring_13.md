# Task 13 - Testes, aposentadoria do legado e checklist de done

## Objetivo
Fechar a refatoracao com cobertura de testes, remocao do monolito e validacao final dos criterios.

## Contexto atual
- Existem testes de workers legados, mas nao para job orchestration completo.
- Criterios exigem remover/aposentar `StartFileProcessingPipeline` e `ScanDirWorker`.

## Escopo
- Adicionar testes unitarios para:
  - orquestrador (jobs/steps/dependencias)
  - idempotencia e `skipped`
  - diff `changed vs unchanged`
- Adicionar testes de integracao para:
  - upload (retorno imediato + job processando)
  - startup scan incremental
  - watcher com debounce e criacao de job
- Remover/aposentar caminhos legados:
  - `StartFileProcessingPipeline` monolitico
  - `ScanDirWorker` descontinuado
- Atualizar analytics/logs e documentacao do modulo worker.

## Arquivos alvo
- `backend/internal/worker/*_test.go`
- `backend/tests/*` (se necessario)
- `backend/internal/worker/fileProcessingPipeline.go`
- `backend/internal/worker/dir.go`
- docs de refatoracao/resultados

## Criterios de pronto
- 3 fluxos principais funcionando end-to-end via Job System:
  - startup scan incremental
  - upload assincrono com `job_id`
  - watcher com debounce criando jobs
- API/consumer consegue acompanhar status, steps e erros.
- Comandos de validacao executados:
  - `cd backend && go test ./... -cover`
  - `make -C backend test`

## Dependencias
- Tasks 1-12.
