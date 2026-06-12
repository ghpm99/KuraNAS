# 14 — Contas de e-mail + OAuth2 (read-only)

**Tipo:** feature (demanda e-mail + kiosk) · **Prioridade:** P2 · **Depende de:** 04 (endpoints atrás da whitelist)

## Contexto

Demanda do dono (2026-06-12): integrar as contas pessoais de e-mail (Gmail e Hotmail/Outlook) ao KuraNAS para filtrar a montanha de spam/phishing e não perder e-mail importante — o resultado aparece no painel kiosk do app legado (task 18). Esta task é a fundação: vincular contas com **OAuth2 somente leitura** e guardar os tokens com segurança. Não existe nada de e-mail no sistema hoje.

O modelo de ameaças da demanda foi analisado e registrado em "Decisões registradas" (regras duras de e-mail) — o veredito é viável **porque** o acesso é read-only e os tokens são cifrados: no pior caso um vazamento expõe leitura da caixa, nunca envio/exclusão, e jamais o servidor de arquivos.

Fluxos OAuth escolhidos (contas pessoais, servidor na LAN sem domínio público):

- **Microsoft (Hotmail/Outlook pessoal)**: **Device Code Flow** — app registrado no Entra como *public client* ("Allow public client flows" habilitado), audience "Personal Microsoft accounts", endpoint `…/consumers/oauth2/v2.0/devicecode`. O backend exibe `user_code` + `verification_uri` na tela de Settings; o dono conclui o login em qualquer dispositivo. Só precisa de `EMAIL_MS_CLIENT_ID` (sem secret). Refresh token pessoal: ~90 dias deslizantes — renova a cada sync, na prática permanente.
- **Google (Gmail)**: o Device Flow do Google **não aceita escopos do Gmail** → Authorization Code + **PKCE**, client tipo *Desktop app*, redirect loopback `http://localhost:8000/api/v1/email/oauth/google/callback`. Pegadinha LAN: `localhost` resolve na máquina do navegador, não no NAS — o vínculo (operação única por conta) é feito num navegador na própria máquina do NAS, ou via túnel `ssh -L 8000:<ip-do-nas>:8000` na máquina do usuário.
- **Caveat Google**: consent screen em modo *Testing* expira refresh tokens em 7 dias → publicar **In production** mesmo sem verificação (escopo restrito mostra a tela "app não verificado"; uso pessoal: Avançado → continuar; o token deixa de expirar).

## Objetivo

O dono vincula suas contas Gmail e Hotmail pela tela de Settings do frontend; o backend guarda os tokens cifrados (AES-256-GCM) e os renova sozinho; contas podem ser listadas, pausadas (sync on/off) e removidas. Nenhum e-mail é buscado ainda (task 15).

## O que fazer

1. Migração `0038` com a tabela `email_account`.
2. Novo `pkg/crypto/aesgcm.go` (Seal/Open com chave da env `EMAIL_TOKEN_KEY`).
3. Domínio `internal/api/v1/email/` com CRUD de contas + os dois fluxos OAuth.
4. Seção "E-mail" na tela de Settings do frontend (vincular Google/Microsoft, listar, pausar, remover).
5. Sem `EMAIL_TOKEN_KEY` configurada, a feature inteira se recusa a ligar (erro i18n explícito).

## Como fazer

- **Migração** `pkg/database/migrations/queries/0038_create_email_account_table.sql` (confirmar o próximo número livre — tasks 12/13 podem criar migrações antes): `email_account(id SERIAL PK, provider VARCHAR CHECK ('google','microsoft'), address TEXT, display_name TEXT, token_ciphertext BYTEA, status VARCHAR CHECK ('linked','error','reauth_required'), sync_enabled BOOL DEFAULT TRUE, last_sync_at TIMESTAMPTZ, last_error TEXT, created_at, updated_at, UNIQUE(provider, address))`. O blob de token (JSON com access/refresh/expiry) é cifrado inteiro; nonce AES-GCM prefixado no ciphertext.
- **Cripto**: `pkg/crypto/aesgcm.go` — chave de 32 bytes base64 em `EMAIL_TOKEN_KEY`. Tokens **nunca** aparecem em DTO, log ou resposta de erro.
- **Domínio** `internal/api/v1/email/` no padrão por prefixo de arquivo (imitar `notifications/` e `aiproviders/`): `model.go`, `dto.go`, `interfaces.go`, `repository.go`, `service.go`, `service_oauth_google.go`, `service_oauth_microsoft.go`, `handler.go`. SQL um-por-pergunta em `pkg/database/queries/email/` com `//go:embed` (imitar `queries/notifications/notifications.go`).
- **Endpoints** (granulares, regra do backend/CLAUDE.md):
  - `GET /api/v1/email/accounts` — lista (sem tokens, claro).
  - `DELETE /api/v1/email/accounts/:id` — remove conta e apaga tokens.
  - `PUT /api/v1/email/accounts/:id/sync-enabled` — liga/desliga sync.
  - `POST /api/v1/email/accounts/google/auth-url` — gera URL de autorização (PKCE: verifier guardado server-side com TTL).
  - `GET /api/v1/email/oauth/google/callback` — troca code por tokens, persiste, responde página mínima de sucesso.
  - `POST /api/v1/email/accounts/microsoft/device-code` — inicia device flow, retorna `user_code` + `verification_uri`; backend faz polling do token em goroutine.
  - `GET /api/v1/email/accounts/microsoft/device-code/status` — frontend acompanha (pending/linked/expired).
- **Allowlist de hosts** (regra dura): o client HTTP do OAuth só fala com `accounts.google.com`, `oauth2.googleapis.com`, `login.microsoftonline.com` (Graph entra na task 15). Qualquer outro host é recusado.
- **Refresh**: helper no service que devolve access token válido (renova com o refresh token se expirado, repersiste cifrado); falha de refresh → status `reauth_required` + `last_error`.
- **Wiring**: `newEmailContext` em `internal/app/context.go` + `RegisterEmailRoutes` em `internal/app/routes.go`, nil-guarded quando `EMAIL_TOKEN_KEY` ausente (endpoints respondem erro i18n `EMAIL_FEATURE_DISABLED_NO_KEY`).
- **Envs**: `EMAIL_TOKEN_KEY` (obrigatória), `EMAIL_GOOGLE_CLIENT_ID`, `EMAIL_GOOGLE_CLIENT_SECRET`, `EMAIL_MS_CLIENT_ID`. Documentar no `.env.example` se existir.
- **Frontend**: `src/service/email.ts` + `src/components/settings/EmailSettingsSection.tsx` + `useEmailSettings.ts` (imitar `AIProvidersSettingsSection.tsx`/`useAIProvidersSettings.ts`), plugada na tela de Settings. Para Microsoft, exibir `user_code` + link e fazer poll do status; para Google, abrir a auth-url e instruir sobre a limitação do loopback (texto i18n).
- **i18n**: `EMAIL_ACCOUNT_LINKED`, `EMAIL_ACCOUNT_LINK_FAILED`, `EMAIL_ACCOUNT_NOT_FOUND`, `EMAIL_ACCOUNT_REMOVED`, `EMAIL_FEATURE_DISABLED_NO_KEY`, `EMAIL_OAUTH_DEVICE_CODE_PROMPT`, `EMAIL_OAUTH_REAUTH_REQUIRED`, `SETTINGS_EMAIL_TITLE`, `SETTINGS_EMAIL_ADD_GOOGLE`, `SETTINGS_EMAIL_ADD_MICROSOFT`, `SETTINGS_EMAIL_SYNC_ENABLED` em `translations/{pt-BR,en-US}.json`.
- **Testes**: repository com sqlmock; OAuth com `httptest.Server` simulando os endpoints de token (imitar `pkg/ai/providers/*/provider_test.go`); cripto com vetores de ida-e-volta; handler com service mockado.

## Critérios de aceite

- [ ] Vincular conta Google e conta Microsoft pela UI funciona de ponta a ponta (validação manual do dono).
- [x] Escopos pedidos são exatamente `gmail.readonly` e `Mail.Read` + `offline_access` — nada de envio. *(Ajuste registrado na implementação: somam-se os escopos de **identidade** `openid email` (e `profile` na Microsoft), necessários para o `id_token` informar **qual endereço** foi vinculado — não concedem nenhuma capacidade de e-mail além da leitura já pedida. Testes garantem a ausência de qualquer escopo de envio/escrita.)*
- [x] Tokens ilegíveis no banco (AES-GCM) e ausentes de qualquer DTO/log/resposta.
- [x] Sem `EMAIL_TOKEN_KEY`, endpoints respondem erro i18n e nada é gravado.
- [x] Access token expirado é renovado sozinho; refresh inválido marca `reauth_required`.
- [x] Remover conta apaga os tokens.
- [x] `make ci` verde (backend + frontend).

## Fora de escopo

- Buscar/armazenar e-mails (task 15) e análise por IA (task 16).
- Qualquer escopo OAuth além de leitura — **jamais** entra, em nenhuma task futura (regra dura).
- Outros provedores de e-mail (IMAP genérico, Yahoo etc.).
