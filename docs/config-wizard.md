# Assistente de configuração de infra (.env)

Editor guiado das variáveis de ambiente do servidor, acessível em
**About → Ferramentas técnicas → Abrir assistente** (`/internal/config-wizard`).
Complementa a tela de **Settings** (que cobre o que é DB-backed) editando o que
hoje só existe no arquivo `.env`.

## Decisões registradas

- **Gating loopback.** Os endpoints `/api/v1/configuration/env*` só respondem ao
  loopback (127.0.0.1/::1), checado pelo `RemoteAddr` real (não por
  `X-Forwarded-For`). Consequência aceita: o assistente **não** funciona ao
  acessar o NAS de outro aparelho da LAN — apenas na própria máquina do servidor.
  Isso adiciona uma barreira além da whitelist de IP para a superfície que grava
  segredos e dados de bootstrap.
- **Segredos write-only.** O `GET` nunca devolve o valor de um segredo
  (`DB_PASSWORD`, `EMAIL_TOKEN_KEY`, client secrets, API keys) — só um flag
  `configured`. No `PUT`, um segredo em branco **mantém** o valor atual; só
  sobrescreve quando o usuário digita algo novo.
- **Restart manual.** O `.env` é lido apenas no boot (`config.InitializeConfig`).
  Uma gravação responde `restart_required: true` e a UI mostra um banner; o
  servidor **não** se reinicia sozinho.
- **Anti-lockout.** Antes de gravar, valores são validados (inteiros positivos,
  bool `true`/`false`, origens com esquema http(s), `EMAIL_TOKEN_KEY` base64 de
  32 bytes). Há endpoints dedicados `POST .../test-db` e `POST .../test-path`
  para checar conexão/caminho antes de salvar, o `.env` é copiado para um
  `.env.<timestamp>.bak` a cada escrita, e as chaves perigosas (`DB_*`,
  `ALLOWED_ORIGINS`, `EMAIL_TOKEN_KEY`) exigem confirmação explícita.

## Variáveis editáveis

Definidas no catálogo `envCatalog` em
`backend/internal/api/v1/configuration/service_env.go`, agrupadas pelas etapas do
stepper: Geral, Banco, Acesso/CORS, Email/OAuth, IA e Workers. Adicionar uma
variável é só incluir uma entrada no catálogo (chave, grupo, tipo, default,
flag de perigosa) e uma label `ENV_FIELD_<KEY>` nos dois catálogos de tradução.
