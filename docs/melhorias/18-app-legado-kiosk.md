# 18 — Tela kiosk do app legado (relógio + notificações + e-mails)

**Tipo:** feature (demanda e-mail + kiosk) · **Prioridade:** P2 · **Depende de:** 16 (endpoints de e-mail analisado), 17 (app enxuto)

## Contexto

Visão do dono (2026-06-12), inspirada em dashboards tmux de parede: o app legado no Galaxy Tab 2 (1024×600, paisagem, sempre ligado na parede) vira um painel com **três seções** — metade superior: relógio gigante + data; metade inferior dividida ao meio: esquerda = notificações do backend, direita = lista de e-mails (remetente + assunto, ou resumo de IA quando existir). Só leitura: o painel é para *ver de relance*, não para interagir.

A base já existe: `KioskManager` mantém tela acesa/fullscreen (API 16 ok), `HomeFragment` já tem relógio com Handler, `TranslationManager` e `HttpClient` sobrevivem à limpeza da task 17. O backend já expõe `GET /api/v1/notifications` (domínio existente) e `GET /api/v1/email/messages` (tasks 15/16).

Restrições do device (2012, Android 4.1): payloads pequenos, sem bibliotecas novas, sem WebView (regra dura de segurança — conteúdo de e-mail nunca é renderizado como HTML), Views simples.

## Objetivo

O tablet na parede mostra hora/data sempre corretas e, sem nenhum toque, as últimas notificações e e-mails se mantêm atualizados por polling, sobrevivendo a quedas do servidor com indicador de offline e retomada automática.

## O que fazer

1. Refazer `HomeFragment` + `fragment_home.xml` no layout de três seções.
2. Criar `NotificationApi` e `EmailApi` no app + mappers/modelos.
3. Polling com intervalos defasados, backoff em falha e indicador offline.
4. Chaves i18n novas no catálogo local (e uso do remoto quando disponível).

## Como fazer

- **Layout** (`res/layout/fragment_home.xml`, paisagem 1024×600): topo ~50% da altura — relógio `TextView` ~120sp (fonte light atual) centralizado + data por extenso abaixo (`EEEE, d 'de' MMMM` via locale); base dividida em dois painéis 50/50 (`LinearLayout` horizontal), cada um com título (`KIOSK_NOTIFICATIONS_TITLE` / `KIOSK_EMAILS_TITLE`) + `ListView` (mais leve que RecyclerView para 8 itens; o RecyclerView já está no app — usar o que ficar mais simples). Tema escuro atual (OLED-friendly) mantido.
- **Relógio**: reaproveitar o Handler existente do `HomeFragment`, tick de **1 s** (formato HH:mm:ss como na referência visual); pausar em `onPause`, retomar em `onResume` (padrão atual).
- **APIs novas** (`data/remote/api/`, imitar o padrão `VideoApi`/`VideoMapper`/`VideoItem` removido na task 17 — recuperar do git history se preciso):
  - `NotificationApi.java` → `GET /api/v1/notifications?page=1&page_size=8` (campos usados: type, title, message, created_at).
  - `EmailApi.java` → `GET /api/v1/email/messages?page=1&page_size=8` (campos: remetente, assunto, snippet/resumo, importância, veredito, received_at).
  - Mappers org.json + modelos de domínio enxutos; registro no `ServiceLocator`.
- **Polling**: notificações a cada **60 s**, e-mails a cada **120 s**, **defasados** (offset inicial de ~30 s entre eles) para não coincidirem nem com o tick do relógio acumulado. Handler + `postDelayed` (sem libs). Em falha: backoff exponencial (60s → 2min → 4min, teto 5min) + indicador discreto "offline" (`KIOSK_OFFLINE`) num canto; ao voltar, restaura intervalo normal. Resposta em voo descartada se o fragment morreu (padrão de callback do `HttpClient`).
- **Render dos itens** (`item_kiosk_notification.xml`, `item_kiosk_email.xml`): só `TextView`s — **proibido WebView/HTML** (todo conteúdo chega como texto puro do backend, e o cliente também não interpreta). E-mail: linha 1 = remetente + hora; linha 2 = assunto; linha 3 = resumo (se houver) em fonte menor; maliciosos/suspeitos aparecem com marcador de cor (vermelho/âmbar do tema) e sem resumo. Notificação: marcador de cor por tipo (info/success/warning/error) + título + mensagem.
- **Payload**: `page_size=8` mantém cada resposta pequena — meta < 4 KB por request (medir com os DTOs reais; se passar, reduzir page_size ou pedir DTO mais enxuto ao backend em task própria — **não** engordar endpoint existente).
- **i18n** (`assets/translations/pt-BR.json` + chaves equivalentes no catálogo remoto do backend para os outros clients se aplicável): `KIOSK_NOTIFICATIONS_TITLE`, `KIOSK_EMAILS_TITLE`, `KIOSK_EMPTY_NOTIFICATIONS`, `KIOSK_EMPTY_EMAILS`, `KIOSK_OFFLINE`, `KIOSK_IMPORTANCE_HIGH`. Tudo via `TranslationManager.t()`; mensagens vindas do servidor exibidas como chegam (regra do projeto).
- **Testes**: mappers com fixtures JSON; lógica de backoff isolada em classe testável (JUnit/Mockito); estados vazio/erro dos painéis.

## Critérios de aceite

- [ ] Relógio nunca para nem atrasa visivelmente (tick 1 s; pausa/retoma com o lifecycle).
- [ ] Painéis atualizam nos intervalos definidos, defasados, com payload < 4 KB por request (medido).
- [ ] Servidor fora do ar: indicador offline, backoff, retomada automática sem crash nem ANR.
- [ ] E-mails maliciosos/suspeitos visualmente distintos e sem resumo; nada renderiza HTML.
- [ ] Estados vazios com texto i18n (`KIOSK_EMPTY_*`).
- [ ] Toda string visível via `TranslationManager` (fallback local + remoto).
- [ ] Roda fluido no Galaxy Tab 2 real — validação manual do dono.
- [ ] Testes unitários verdes (`./gradlew test`).

## Fora de escopo

- Interação com itens (marcar lido, abrir e-mail, scroll infinito) — painel é só leitura.
- Modo retrato; outras resoluções além do tablet alvo.
- Push/SSE (polling basta para um painel de parede; evolução futura se o backend ganhar SSE genérico).
- Mudanças de backend (qualquer ajuste de DTO vira task própria).
