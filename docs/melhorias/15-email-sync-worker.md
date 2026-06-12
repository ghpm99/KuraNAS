# 15 — Sincronização de e-mail (worker + pré-filtro)

**Tipo:** feature (demanda e-mail + kiosk) · **Prioridade:** P2 · **Depende de:** 14 (contas vinculadas)

## Contexto

Com as contas vinculadas (task 14), o backend precisa buscar as mensagens novas periodicamente, extrair o que importa (remetente, assunto, headers de autenticação, metadados de anexos), sanitizar o corpo e barrar o lixo óbvio **antes** de qualquer IA. O pré-filtro determinístico é parte do modelo de ameaças: spam/phishing barrado aqui **não chega ao LLM** (regra dura 7), o que reduz tanto custo quanto superfície de prompt injection.

As regras duras de segurança (em "Decisões registradas") que esta task implementa diretamente: nunca baixar anexo (metadados apenas); HTML → texto puro com remoção de Unicode invisível e corpo ≤ 16 KB; client HTTP com allowlist fixa de hosts; URLs de e-mail são dados, jamais visitadas.

## Objetivo

Job periódico `email_sync` busca mensagens novas de cada conta habilitada via Gmail API / Microsoft Graph, persiste em `email_message` já sanitizadas e pré-classificadas (`pending` ou `prefiltered_spam`), com expurgo por retenção. Um endpoint enxuto lista as mensagens para os clients.

## O que fazer

1. Migração `0039` com a tabela `email_message`.
2. Clients das APIs externas em `pkg/mailfetch/` (Gmail + Graph) com allowlist de hosts.
3. Sanitizador (HTML→texto) e pré-filtro determinístico no domínio `email`.
4. Job `email_sync` (steps `email_fetch` → `email_prefilter`) agendado por ticker.
5. Endpoints: listagem paginada de mensagens (DTO enxuto) e gatilho manual de sync.
6. Expurgo por retenção (`EMAIL_RETENTION_DAYS`).

## Como fazer

- **Migração** `0039_create_email_message_table.sql`: `email_message(id SERIAL PK, account_id INT FK→email_account, provider_message_id TEXT, sender_name TEXT, sender_address TEXT, subject TEXT, snippet TEXT, sanitized_body TEXT NULL, received_at TIMESTAMPTZ, auth_results JSONB, attachment_meta JSONB, link_domains JSONB, status VARCHAR CHECK ('pending','prefiltered_spam','analyzed','failed'), created_at, UNIQUE(account_id, provider_message_id))` + índice `(account_id, received_at DESC)`. `auth_results` = `{spf, dkim, dmarc}` extraído do header `Authentication-Results`; `attachment_meta` = lista de `{filename, mime, size}`; `link_domains` = domínios das URLs encontradas no corpo (dados para evidência — as URLs **nunca** são visitadas).
- **Clients** `pkg/mailfetch/` (espelhar a estrutura de `pkg/ai/providers/`): `fetcher.go` define a interface comum (`ListNewMessages(ctx, account, since) ([]RawMessage, error)`); `gmail/client.go` usa `users.messages.list` + `users.messages.get?format=metadata` e corpo via `format=full` **sem** `attachments.get`; `graph/client.go` usa `/me/messages?$select=...,internetMessageHeaders,bodyPreview,body` **sem** `$expand` de conteúdo de anexo. Allowlist hardcoded: `gmail.googleapis.com`, `graph.microsoft.com` (+ hosts de token da task 14); qualquer outro host → erro.
- **Sanitização** (`internal/api/v1/email/sanitizer.go`): HTML → texto com `golang.org/x/net/html` percorrendo só text nodes (scripts/CSS descartados); remoção de caracteres Unicode invisíveis (zero-width, controles bidi); truncar em 16 KB **antes** do parse pesado; gerar `snippet` ≤ 280 chars.
- **Pré-filtro** (`internal/api/v1/email/prefilter.go`), heurísticas determinísticas → `prefiltered_spam`: DMARC `fail`; domínio do remetente divergente dos domínios dos links (típico de phishing) combinado com outros sinais; anexo com extensão perigosa (`.exe`, `.scr`, `.js`, `.vbs`, `.bat`, dupla extensão); padrões clássicos de spam no assunto. Cada regra é uma função pura testável; o resultado guarda quais regras dispararam (vira evidência na task 16).
- **Job**: novo `JobType` `email_sync` + `StepType`s `email_fetch`, `email_prefilter` em `internal/worker/job/job_domain.go` (incluir nos `IsValid`); executores `internal/worker/engine/step_email_fetch.go` e `step_email_prefilter.go` (imitar `step_takeout.go`). Agendamento: ticker enfileirando `email_sync` a cada `EMAIL_SYNC_INTERVAL_MINUTES` (default 10), no padrão do `startup_scan`. Sync incremental: cursor por conta (`last_sync_at` + `UNIQUE(account_id, provider_message_id)` torna o reprocesso idempotente). Cap de mensagens por execução (ex.: 100/conta) contra bombas.
- **Token expirado** durante o fetch: marcar conta `reauth_required` + notificação i18n via `notifications.GroupOrCreate` (`EMAIL_OAUTH_REAUTH_REQUIRED`); o job segue com as outras contas (`partial_fail` se alguma falhou).
- **Endpoints**:
  - `GET /api/v1/email/messages?page=&page_size=` — DTO enxuto (id, remetente, assunto, snippet, received_at, status, e quando existir análise: veredito/importância/resumo — task 16). **Sem corpo**. Pensado para o kiosk num tablet de 2012: payload pequeno.
  - `POST /api/v1/email/accounts/:id/sync` — gatilho manual (enfileira o job).
- **Retenção**: step ou rotina no fim do job expurga mensagens com `received_at` além de `EMAIL_RETENTION_DAYS` (default 30).
- **i18n**: `EMAIL_SYNC_COMPLETED`, `EMAIL_SYNC_FAILED` em `translations/{pt-BR,en-US}.json`.
- **Testes**: clients com `httptest.Server` (fixtures de resposta Gmail/Graph) — nasce junto por causa da cobertura ≥ 80%; sanitizador com fixtures de HTML malicioso (script, zero-width, corpo gigante); pré-filtro regra a regra; teste garantindo que **nenhuma** chamada de download de anexo é feita (servidor de teste falha se a rota for tocada); allowlist recusando host estranho; repository com sqlmock.

## Critérios de aceite

- [ ] Sync periódico busca só o novo e não duplica (idempotente via UNIQUE).
- [ ] Corpo armazenado é texto puro ≤ 16 KB, sem tags e sem Unicode invisível.
- [ ] Anexos: só metadados; teste prova que rota de download nunca é chamada.
- [ ] URLs do corpo viram `link_domains` (dados) e nenhuma é visitada.
- [ ] Client recusa qualquer host fora da allowlist (teste).
- [ ] Spam óbvio cai em `prefiltered_spam` com as regras disparadas registradas.
- [ ] Token inválido marca a conta `reauth_required` + notificação, sem derrubar o job.
- [ ] Expurgo por retenção funciona.
- [ ] `make ci` verde.

## Fora de escopo

- Análise por IA (task 16).
- Reputação externa de URLs/anexos (decisão registrada: fora do v1).
- Pastas além da caixa de entrada; busca/filtros avançados.
- Qualquer ação na caixa (marcar lido, mover, excluir) — acesso é read-only.
