# 04 — Controle de acesso por whitelist de IPs

**Tipo:** segurança · **Prioridade:** P0

> **Decisão registrada (2026-06-11):** o sistema **não terá autenticação** (login/senha/token). É um produto de rede interna, single-user por natureza. O controle de acesso é por **whitelist de IPs cadastrados** — só dispositivos cuja origem está na lista falam com o servidor.

## Contexto

Hoje **não existe nenhum controle de acesso**: `internal/app/routes.go` monta todas as rotas de `/api/v1` sem middleware de proteção, e `internal/discovery` ainda anuncia o servidor na LAN via mDNS e UDP. Qualquer pessoa ou dispositivo na rede pode listar, baixar, mover e **apagar permanentemente do disco** (`DeleteFileFromDisk` usa `os.RemoveAll`) qualquer arquivo do NAS.

O modelo escolhido: o dono cadastra os IPs (ou faixas) dos seus dispositivos; todo o resto recebe 403. `localhost` é sempre liberado, para o dono nunca se trancar para fora do próprio servidor.

### Limitações aceitas (registrar, não resolver)

- IP identifica **dispositivo**, não pessoa — qualquer processo num dispositivo liberado tem acesso total.
- DHCP pode reatribuir um IP liberado a outro dispositivo — mitigável pelo usuário com reserva DHCP ou cadastro de IP fixo, fora do escopo do sistema.
- Spoofing de IP na própria LAN é teoricamente possível. Decisão consciente para rede doméstica confiável; revisitar se o produto um dia for exposto além da LAN.

## Objetivo

Nenhuma requisição é atendida (API, UI, futuro `/dav`) se o IP de origem não for `localhost` nem casar com uma entrada habilitada da whitelist. A lista é gerenciável pela UI a partir de um dispositivo já liberado, com efeito imediato.

## O que fazer

1. Tabela de IPs/faixas permitidos + migração.
2. Middleware global de whitelist no Gin, antes de qualquer rota.
3. Endpoints CRUD de gestão da lista + tela em Settings no frontend.
4. Bloquear falsificação de IP via cabeçalhos de proxy.

## Como fazer

- **Modelo**: migração criando `allowed_ip` (`id`, `cidr` text, `label`, `enabled`, `created_at`). Armazenar sempre como CIDR — IP único vira `/32` (IPv4) ou `/128` (IPv6); faixas como `192.168.1.0/24` são suportadas de graça.
- **Domínio**: pacote `internal/api/v1/accesscontrol/` no padrão do projeto (`handler.go`, `service.go`, `repository.go`, `interfaces.go`, `model.go`, `dto.go`), queries em `pkg/database/queries/accesscontrol/`, contexto em `context.go`, rotas em `routes.go`:
  - `GET /api/v1/access-control/ips` — lista;
  - `POST /api/v1/access-control/ips` — cadastra (validar CIDR com `net/netip`);
  - `PUT /api/v1/access-control/ips/:id` — edita/habilita/desabilita;
  - `DELETE /api/v1/access-control/ips/:id`.
- **Middleware**: registrado no engine **antes de tudo** (API, assets, SPA, Swagger). Lógica: resolver o IP remoto → loopback (`127.0.0.0/8`, `::1`) sempre passa → senão, casar contra as entradas habilitadas (`netip.Prefix.Contains`). Fora da lista → **403** com mensagem i18n (`ERROR_IP_NOT_ALLOWED`), incluindo o IP solicitante no corpo para facilitar o cadastro (não vaza nada além do que o cliente já sabe).
- **Anti-spoofing (crítico)**: chamar `router.SetTrustedProxies(nil)` — sem isso o `ClientIP()` do Gin honra `X-Forwarded-For` e qualquer cliente falsifica o IP de origem, furando a whitelist. Usar o IP do `RemoteAddr` da conexão. Se um dia houver reverse proxy na frente, configurar o proxy explicitamente como confiável (documentar no README).
- **Cache**: a lista muda raramente e é consultada em toda requisição — manter em memória (atômico), recarregando quando o CRUD altera algo (mesmo padrão `SetOnChange` usado pelo `aiproviders`). Nada de query por request.
- **IPv6**: normalizar com `net/netip` (cuidado com IPv4-mapped `::ffff:192.168.x.x`, que deve casar com a entrada IPv4 correspondente — `netip.Addr.Unmap`).
- **UX de primeiro acesso**: lista vazia = só localhost acessa. O dono abre a UI no próprio servidor e cadastra os dispositivos; a tela de Settings mostra o IP de quem está acessando para cadastro em um clique.
- **i18n**: mensagens novas nos dois catálogos (`pt-BR`, `en-US`).
- **Testes**: middleware (loopback passa; IP fora → 403; IP dentro de CIDR passa; entrada desabilitada não passa; `X-Forwarded-For` forjado **não** engana), CRUD com validação de CIDR inválido, recarga do cache ao alterar a lista.

## Critérios de aceite

- [x] Requisição de IP não cadastrado recebe 403 com mensagem i18n em **qualquer** rota (API, assets, SPA). *(middleware registrado no engine antes de todas as rotas; `TestMiddlewareBlocksUnknownIPWithI18nBody`)*
- [x] `localhost` acessa sempre, mesmo com a lista vazia — impossível se trancar para fora no próprio servidor. *(`TestMiddlewareLoopbackAlwaysPasses`, inclui `::1` e `127.0.0.53`)*
- [x] Cadastro/edição/remoção pela UI tem efeito imediato, sem restart. *(cache atômico recarregado a cada mutação; `TestIsAllowedReflectsCRUDImmediately`; seção em Settings com toggle/remover e cadastro do dispositivo atual em um clique)*
- [x] Faixas CIDR funcionam (ex.: liberar `192.168.1.0/24` libera a sub-rede). *(`TestMiddlewareAllowsRegisteredCIDRRange`)*
- [x] Teste prova que `X-Forwarded-For`/`X-Real-IP` forjados não contornam a whitelist. *(`TestMiddlewareForgedProxyHeadersDoNotBypass` + `SetTrustedProxies(nil)` no boot)*
- [x] Cliente IPv4 via socket IPv6 (`::ffff:...`) casa com a entrada IPv4 cadastrada. *(`TestMiddlewareIPv4MappedClientMatchesIPv4Entry`; normalização com `netip.Addr.Unmap` nos dois lados)*
- [x] `make ci` verde (backend + frontend). *(2026-06-11, integração contra Postgres 18 local; migração 0034 verificada no banco real)*

## Fora de escopo

- **Login, senha, tokens, sessões, usuários, 2FA** — decisão registrada: não haverá autenticação.
- **TLS/HTTPS** — rede interna; revisitar apenas se o produto for exposto além da LAN.
- Rate limiting / detecção de força bruta (sem senha, não há o que forçar).
- Bloqueio na camada de descoberta (mDNS/UDP continuam anunciando; a proteção é na porta 8000).
- Adaptação dos clientes (`frontend` consome a API normalmente; android/mobile/plugin só precisam estar em IP cadastrado — nenhuma mudança de código neles).
