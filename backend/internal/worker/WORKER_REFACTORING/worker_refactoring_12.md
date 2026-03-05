# Task 12 - Concorrencia, prioridade, retry e cancelamento minimo

## Objetivo
Adicionar controles operacionais para estabilidade e previsibilidade da execucao.

## Contexto atual
- Concurrency hardcoded e sem politicas por tipo de step.
- Sem prioridade formal, retry estruturado e cancelamento por job.

## Escopo
- Implementar limite de concorrencia por tipo de step (checksum/thumbnail/metadata etc).
- Aplicar prioridade de fila: `high > normal > low`.
- Implementar retry com backoff para erros transitorios e limite de tentativas.
- Implementar cancelamento minimo para `startup_scan` e `reindex_folder`.
- Garantir respeito a `context cancellation` em steps longos.
- Expor configuracoes em `internal/config`.

## Arquivos alvo
- `backend/internal/worker/*`
- `backend/internal/config/*`
- (se aplicavel) endpoint de cancelamento em `api/v1/jobs`

## Criterios de pronto
- Sem explosao de goroutines sob carga.
- Jobs de upload executam antes de startup quando concorrendo.
- Jobs cancelados mudam estado de forma consistente.

## Dependencias
- Tasks 4-11.
