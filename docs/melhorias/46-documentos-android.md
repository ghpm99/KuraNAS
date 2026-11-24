# 46 — Documentos: tela no Android

**Tipo:** feature · **Prioridade:** P1 · **Spec:** [07](../especificacao/07-documentos.md) `RF-12`, `CA-12`; [24](../especificacao/24-app-android.md) `RN-08`, `RF-03`

## Contexto

O app Android não tem Documentos (spec 24 §8). O dono quer achar e abrir um documento pelo celular.

## Objetivo

Listar, buscar por nome e conteúdo, abrir e baixar documentos pelo app.

## O que fazer

1. **Dados** `feature/documents/data`: `DocumentsApi` (`GET /documents`, `/documents/count`, `/documents/search`), DTOs com os nomes de campo do contrato de 07 §8, repositório.
2. **Lista paginada** com `page`/`page_size` explícitos (50), "carregar mais" ao chegar ao fim, puxar para recarregar, estados distintos de carregando/vazio/erro e erro de página seguinte sem esvaziar a lista (`24/RN-08`).
3. **Filtro por formato** (chips) e **busca** por nome e conteúdo, com o trecho no resultado.
4. **Abrir**: texto no visualizador do app; PDF e demais formatos por "Abrir com…" (baixa para o cache e entrega a outro aplicativo — `03/RF-18`; `07/RF-12` permite PDF no app ou por "Abrir com…", e o visualizador do app só renderiza texto).
5. **Baixar** para a pasta Downloads do aparelho (usa o download em fluxo da task 36).
6. **Entrada no menu "Mais"** (`24/RF-03`).

## Critérios de aceite

- [x] `07/CA-12` — buscar um documento pelo conteúdo, abrir e baixar funcionam. *Backend local: `/documents/search?q=reajuste` acha o PDF pelo conteúdo e `/files/download/:id` o entrega (200); o app usa essas rotas (repositório testado com a resposta real) e abre texto no visualizador e PDF por "Abrir com…" (`07/RF-12` permite as duas formas), baixando pela fila de transferências (testes do ViewModel). Aparelho não exercitado.*
- [x] A lista rola além de três páginas; girar o aparelho mantém os itens e a posição (`24/CA-11`). *Backend local com 600+ documentos (página 3 de 50 responde 200); paginação e estado no ViewModel, posição em `rememberLazyListState` (testes); rotação no aparelho não exercitada.*
- [x] Com o servidor recusando, a mensagem do servidor é mostrada (`24/CA-14`). *Teste do repositório e do ViewModel com `403 {"error": ...}` mostrando a mensagem como veio; a recusa por IP também cai na tela de acesso não liberado da task 39.*
- [x] Textos em `strings.xml` (pt-BR e en-US); nenhum literal.
- [x] `assembleDebug` e `testDebugUnitTest` verdes; teste do repositório decodificando um JSON igual ao do handler real. *O teste lê os golden do backend (`backend/internal/api/v1/documenttext/testdata/`).*

## Fora de escopo

- Filtros além de formato; painel de pastas; renomear/mover pelo app.

## Depende de

Tasks 44, 36 e 39 (menu "Mais").
