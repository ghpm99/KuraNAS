# 20 — Web: a segunda página de Arquivos, Favoritos, Recentes, busca na pasta e seletor de pasta

**Tipo:** bug crítico · **Prioridade:** P0 (N2, N7) · **Spec:** [00](../especificacao/00-padroes-transversais.md) `RN-02`, `RN-08`, `RN-49`, `RN-53`, `CA-01`, `CA-02`, `CA-16`, anexo `A1`–`A3`; [02](../especificacao/02-arquivos.md) `CA-01`

## Contexto

O servidor responde listagens com `pagination.has_next` (`backend/pkg/utils/dto.go:9-14`). Cinco telas da web leem `pagination.hasNext`, que nunca existe, então **a segunda página nunca é pedida**:

| Tela | Onde lê `hasNext` | Itens que mostra no máximo |
|---|---|---|
| Arquivos, Favoritos, Recentes | `frontend/src/features/files/providers/fileProvider/index.tsx:156` | 200 |
| Busca na pasta | `frontend/src/features/files/search/useFileSearchResults.ts:41` | 100 |
| Seletor de pasta (mover/copiar) | `frontend/src/components/folderPicker/folderPicker.tsx:74` | 200 |

O tipo errado está declarado em `features/files/providers/fileProvider/fileContext.ts:73-83` (`hasNext`, `hasPrevious`, `pageSize`). O tipo certo já existe e é usado por música, imagens, vídeos, notificações, capturas, diário e lixeira: `frontend/src/types/pagination.d.ts` (`Pagination<T>`, `PageInfo` com `page`, `page_size`, `has_next`, `has_prev`).

Os testes passam porque os mocks foram escritos à mão com `hasNext`:
`useFileSearchResults.test.tsx`, `useFileSearchResults.integration.test.tsx`, `fileProvider/index.test.tsx`, `fileProvider.test.tsx` (helper em `:106-110`), `fileProvider.deepLink.test.tsx`, `fileProvider.integration.test.tsx` (mistura os dois formatos), `FilesExplorerScreen.search.test.tsx`, `FilesExplorerScreen.detailsStar.test.tsx`, `folderPicker.test.tsx`, `FavoritesNavigation.test.tsx`.

## Objetivo

Toda listagem de arquivos da web carrega todas as páginas, e nenhum teste do frontend consegue mais passar com um envelope que o servidor não envia.

## O que fazer

Uma alteração por item, em commit próprio.

Os itens 2, 3 e 4 saíram num commit só (2026-11-16): os três leem o mesmo tipo `PaginationResponse` de `fileContext.ts`, e trocá-lo num só quebra a compilação dos outros dois.

1. **Fixture do envelope real**: um teste no backend grava (golden file versionado) o JSON que `utils.PaginationResponse` produz para uma página com próxima e para a última página; um construtor único de página para testes do frontend (`buildServerPage(items, {page, pageSize, hasNext})`) produz exatamente esse formato e é conferido contra o golden file. É a base de `CA-16`.
2. **`fileProvider`**: `PaginationResponse` de `fileContext.ts` passa a ser o `Pagination<T>` compartilhado; `getNextPageParam` lê `has_next`. Todos os mocks dos testes de `fileProvider` e de `FilesExplorerScreen` passam a usar o construtor do item 1.
3. **Busca na pasta** (`useFileSearchResults.ts`): mesma correção e mesmos mocks.
4. **Seletor de pasta** (`folderPicker.tsx`): lê `has_next`; mocks corrigidos. (O erro de primeira página e o filtro no cliente, também citados em A3, são da task 21 e de sub-item próprio — não entram aqui.)
5. **Tamanho de página** (`RN-08`): as três listagens passam a pedir 50 (lista) ou 60 (grade), no lugar de 200/100/200.
6. **`service/documents.ts`**: `DocumentSearchPagination` é substituído pelo tipo compartilhado (hoje duplica os mesmos campos).

## Critérios de aceite

Verificados no sistema rodando, em uma pasta com mais de 500 itens.

- [ ] `00/CA-01`, `CA-02` — em Arquivos, rolar até o fim carrega a segunda e a terceira página sem clique; os itens anteriores continuam na tela.
- [ ] `02/CA-01` — a pasta com mais de 500 itens mostra todos; o total do cabeçalho confere.
- [ ] O mesmo em Favoritos e em Recentes com mais de três páginas.
- [ ] Buscar na pasta um termo com mais de 150 resultados mostra todos ao rolar.
- [ ] No seletor de pasta de "Mover", uma pasta com mais de 150 subpastas mostra todas.
- [x] `00/CA-16` — existe teste por tela que usa o construtor conferido contra a resposta real do handler.
- [x] `grep -rn "hasNext\b" frontend/src` não encontra nenhuma leitura de campo de resposta da API (sobra só estado de tela, como em `useImageViewer`).
- [x] `make ci` verde.

## Fora de escopo

- Falha de página seguinte que troca a lista pelo erro (task 21).
- Paginação do explorer do Android (task 23).
- Mudar o envelope do servidor.
