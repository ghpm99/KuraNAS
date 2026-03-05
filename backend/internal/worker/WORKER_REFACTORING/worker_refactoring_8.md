# Task 8 - Upload assincrono com retorno de job_id

## Objetivo
Refatorar upload para responder rapido e processar pesado em background via job.

## Contexto atual
- `UploadFilesHandler` salva arquivos e chama `ScanDirTask`.
- Nao retorna `job_id` nem progresso por job.

## Escopo
- Alterar fluxo de upload para:
  - salvar arquivo localmente
  - criar job `upload_process` com steps condicionais
  - retornar resposta imediata com `job_id` e referencia dos arquivos
- Definir prioridade `high` para jobs de upload.
- Manter compatibilidade de contrato o maximo possivel (avaliar `200` vs `201/202`).
- Incluir i18n para novas mensagens de resposta/erro.

## Arquivos alvo
- `backend/internal/api/v1/files/operations_handler.go`
- `backend/internal/api/v1/files/service.go`
- `backend/translations/*.json`

## Criterios de pronto
- Upload nao bloqueia processamento pesado.
- Resposta inclui `job_id`.
- Processamento posterior ocorre via steps (`metadata/checksum/persist/thumbnail/playlist_index`).

## Dependencias
- Tasks 4-7.
