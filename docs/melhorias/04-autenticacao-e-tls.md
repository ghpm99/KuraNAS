# 04 — Autenticação na API (e TLS opcional)

**Tipo:** segurança · **Prioridade:** P0

## Contexto

Hoje **não existe nenhuma autenticação**: `internal/app/routes.go` monta todas as rotas de `/api/v1` sem middleware de auth. Ao mesmo tempo, `internal/discovery` anuncia o servidor na LAN via mDNS e UDP. O resultado é que qualquer pessoa ou dispositivo na rede pode listar, baixar, mover e **apagar permanentemente do disco** (`DeleteFileFromDisk` usa `os.RemoveAll`) qualquer arquivo do NAS.

Para um produto cuja função é guardar os arquivos do usuário, essa é a lacuna de maturidade mais grave frente a qualquer NAS estabelecido (Synology, TrueNAS, OMV — todos têm usuários/sessões como núcleo).

Consumidores que precisam acompanhar a mudança: frontend (`frontend/src/service/*.ts`), os dois apps Android (`android/`, `mobile/`) e o plugin (`plugin/src/background/uploader.js`).

## Objetivo

Nenhuma rota que lê ou modifica dados responde sem credencial válida. Modelo mínimo: **um usuário (dono do NAS) com senha**, sessão por token. TLS disponível por configuração para quem quiser expor fora do localhost.

## O que fazer

1. Backend: senha de admin + endpoint de login que emite token + middleware que protege `/api/v1`.
2. Frontend: tela de login, armazenamento do token, envio em todas as chamadas, tratamento de 401.
3. Clientes restantes (android, mobile, plugin): suporte ao token em fases subsequentes (a API deve aceitar um período de transição configurável — ver "Como fazer").
4. TLS opcional via certificado configurado.

## Como fazer

- **Credencial**: hash da senha (bcrypt/argon2) armazenado na tabela de configuração (ou env `ADMIN_PASSWORD_HASH` no primeiro momento). Fluxo de primeiro acesso define a senha.
- **Sessão**: `POST /api/v1/auth/login` recebe a senha e devolve um token opaco persistido (tabela `session`) ou JWT assinado com segredo gerado no primeiro boot. Token opaco é mais simples de revogar — preferível aqui.
- **Middleware**: registrado no grupo `/api/v1` em `routes.go`, validando `Authorization: Bearer <token>`. Exceções: `/auth/login`, `/health`, `/translations` (o frontend precisa dos textos antes do login) e os assets/SPA. As rotas de descoberta (mDNS/UDP) continuam anunciando, mas a API por trás passa a exigir token.
- **Transição dos clientes**: flag de configuração `AUTH_REQUIRED` (default `true` em instalação nova, `false` em upgrade até o usuário ativar) para não quebrar android/mobile/plugin de uma vez. Documentar no README que o default vira `true` em versão futura.
- **Mensagens**: erros 401/403 via i18n (`ERROR_UNAUTHORIZED`, etc.) nos dois catálogos (`pt-BR`, `en-US`), conforme a regra do projeto.
- **TLS**: `TLS_CERT_FILE`/`TLS_KEY_FILE` no env; quando presentes, `Run()` sobe com `http.ListenAndServeTLS`. Sem certificado, comportamento atual (HTTP) — TLS é opcional nesta task.
- **Rate limit do login**: contenção simples de força bruta (ex.: backoff exponencial por IP em memória). Nada sofisticado.
- **Testes**: middleware (sem token → 401; token inválido → 401; token válido → passa), login (senha errada → 401 + backoff), exceções liberadas.

## Critérios de aceite

- [ ] Com `AUTH_REQUIRED=true`, qualquer rota de dados sem token responde 401 com mensagem i18n.
- [ ] Login com senha correta devolve token; token funciona em chamadas subsequentes; logout/revogação invalida.
- [ ] Frontend: fluxo completo de login, persistência de sessão e redirecionamento em 401.
- [ ] `/health`, `/auth/login`, `/translations` e o SPA continuam acessíveis sem token.
- [ ] Com `TLS_CERT_FILE`/`TLS_KEY_FILE` configurados, o servidor sobe em HTTPS.
- [ ] Instalação existente faz upgrade sem quebrar clientes (flag de transição documentada).
- [ ] `make ci` verde (backend + frontend).

## Fora de escopo

- Multiusuário, perfis, ACL por pasta, quotas — o modelo é single-user nesta task.
- OAuth/OIDC, 2FA, passkeys.
- Adaptação dos apps `android/`, `mobile/` e do `plugin/` (tasks próprias quando a flag de transição for virar default).
- Renovação automática de certificado (Let's Encrypt etc.) — o usuário fornece o certificado.
