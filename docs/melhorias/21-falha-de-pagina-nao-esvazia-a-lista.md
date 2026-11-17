# 21 — Web: falha ao carregar mais não troca a lista pela tela de erro

**Tipo:** bug crítico · **Prioridade:** P0 · **Spec:** [00](../especificacao/00-padroes-transversais.md) `RN-09`, `RN-10`, `RN-13`, `RN-14`, `RN-31`, `CA-04`, `CA-05`, `CA-16`, anexo `A4`

## Contexto

É o defeito "na segunda requisição a tela fica sem itens". Com o react-query v5, quando `fetchNextPage` ou uma revalidação em segundo plano falha, `status` vira `'error'` **com os dados ainda em cache**. A revalidação automática ao voltar o foco está ligada (`refetchOnWindowFocus: true`, `components/providers/appProviders.tsx:22`). Oito telas testam o erro antes de desenhar os itens e devolvem só a tela de erro:

| Tela | Onde |
|---|---|
| Arquivos, Favoritos, Recentes, busca na pasta | `features/files/fileContent/fileContent.tsx:120-133` |
| Busca de documentos | `features/files/search/DocumentSearchResults.tsx:45-53` |
| Vídeos — filmes | `features/videos/components/videoContent/movieBrowser/VideoMovieBrowser.tsx:24-32` |
| Vídeos — pastas e vídeos da pasta | `…/folderBrowser/VideoFolderBrowser.tsx:33-41` e `:57-65` |
| Notificações | `components/notifications/NotificationsScreen.tsx:98-106` |
| Diário | `components/activityDiary/ActivityList/ActivityList.tsx:39-47` (`hasEntriesError` em `providers/activityDiaryProvider/index.tsx:225`) |

Quem já faz certo: Música (`features/music/views/components/MusicCollectionListFeedback.tsx` — erro em tela cheia só com lista vazia; com itens, erro abaixo da lista; `useMusicInfinitePages.ts:19-24` repete `fetchNextPage` quando foi a página seguinte que falhou) e Imagens (`components/imageContent/imageContent.tsx:203-216`, `useImageContent.ts:253-281`). Não existe componente genérico: o de Música mora dentro da feature.

## Objetivo

Em toda lista, uma falha depois da primeira carga mantém os itens na tela e mostra o erro no fim da lista, com "Tentar de novo" repetindo só o que falhou.

## O que fazer

1. **Componente compartilhado** `components/listFeedback/` (a partir de `MusicCollectionListFeedback`, que passa a usá-lo): recebe `isEmpty`, `isError`, `errorMessage`, `onRetry`, `hasNextPage`, `isFetchingNextPage`, `onLoadMore` e os `children`. Regras: erro em tela cheia **só** com `isError && isEmpty`; com itens, desenha os itens e o `ErrorState` abaixo; o sentinela de carregar mais recebe `hasNextPage && !isError`.
2. **Hook compartilhado** de repetição: `fetchNextPage` quando `isFetchNextPageError`, senão `refetch` (hoje em `useMusicInfinitePages` e `useImageContent`).
3. **Uma tela por commit**, adotando 1 e 2 e ganhando teste "primeira página carrega, segunda falha → itens da primeira continuam + erro no fim + tentar de novo continua": 3.1 Arquivos/Favoritos/Recentes/busca na pasta (`fileContent.tsx`) · 3.2 busca de documentos · 3.3 Vídeos — filmes · 3.4 Vídeos — pastas e vídeos da pasta · 3.5 Notificações · 3.6 Diário · 3.7 seletor de pasta (`folderPicker.tsx`: erro na primeira página mostra estado de erro em vez de lista vazia — `RN-10`, anexo `A3`).
4. **Revalidação em segundo plano que falha** (`RN-13`, `RN-14`): nas mesmas telas, mantém os itens e avisa por aviso curto, sem `ErrorState` em tela cheia.
5. Os testes que hoje fixam o comportamento errado são reescritos: `fileContent.states.test.tsx`, `fileContent.test.tsx`, `fileContent.search.test.tsx`, `NotificationsScreen.integration.test.tsx`, `activityDiaryComponents.test.tsx`, `VideoFolderBrowser.behavior.test.tsx`, `VideoMovieBrowser.behavior.test.tsx`, `DocumentSearchResults*.test.tsx`.

## Sub-itens levantados na execução (2026-11-16)

- [x] 21.1. `features/music/views/components/ArtistPage.tsx` tem um `LoadMoreSentinel` fora do `ListFeedback` com `!albumsQuery.isError`: uma revalidação em segundo plano que falha para o carregar mais dos álbuns. Passa a desligar só com `isFetchNextPageError` — `RN-09`, `RN-13`.
- [x] 21.2. O seletor de pasta mostra só o título do erro; a mensagem traduzida que o servidor devolve (`{"error": ...}`) não aparece. Guardar a mensagem da falha e passá-la como `errorMessage` — `RN-10`.

## Critérios de aceite

Verificados no sistema rodando, em cada uma das telas da tabela, com pelo menos três páginas.

- [ ] `00/CA-04` — derrubar o servidor entre a primeira e a segunda página: os itens ficam, aparece erro com "Tentar de novo" no fim; religar e tentar continua do ponto certo. *(marcar por tela: Arquivos · Favoritos · Recentes · busca na pasta · busca de documentos · filmes · pastas de vídeo · Notificações · Diário)*
- [ ] Com a lista carregada, derrubar o servidor, trocar de aba por mais de um minuto e voltar: os itens continuam na tela.
- [ ] `00/CA-05` — com o servidor desligado desde o início, a tela mostra o estado de erro (não o vazio).
- [ ] No seletor de pasta, com o servidor desligado, aparece erro com "Tentar de novo".
- [ ] Música e Imagens continuam passando em `00/CA-04` (o componente compartilhado não as regrediu).
- [x] `make ci` verde.

## Fora de escopo

- Capturas, Lixeira, Assistente, Analytics e Jobs (anexo `A5`–`A8`): não têm scroll infinito ainda; entram pelos itens do board de revisão.
- Android (o `PagedListLoader` já mantém a lista; a ausência de aviso de erro é do item de Android).

## Depende de

Task 20 (sem ela, a segunda página nem é pedida em Arquivos e não há como verificar).
