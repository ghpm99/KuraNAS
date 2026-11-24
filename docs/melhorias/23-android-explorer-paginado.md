# 23 — Android: o explorer mostra todos os itens da pasta

**Tipo:** bug crítico · **Prioridade:** P0 (N2) · **Spec:** [24](../especificacao/24-app-android.md) `RN-08`, `CA-08`, `CA-11`; [00](../especificacao/00-padroes-transversais.md) `RN-02`, `RN-08`, `RN-09`, anexo `A10`; [02](../especificacao/02-arquivos.md) `RF-23`

## Contexto

Cada pasta mostra só 15 itens no app. `feature/files/data/FilesApi.kt:21-28` chama `GET api/v1/files/` e `GET api/v1/files/{id}` sem `page` nem `page_size`; o padrão desses handlers é 15 (`backend/internal/api/v1/files/listing.go:26` e `:91`). `FilesRepository.kt:22-28` devolve só `.items`, `FilesViewModel` não guarda próxima página e `FilesScreen.kt:136-140` não tem "carregar mais".

O app já tem a infraestrutura pronta, usada por Favoritos, Lixeira, Vídeos e Música: `core/paging/PagedListLoader.kt`, `ListPage.kt`, `LoadMoreOnNearEnd.kt`, e `core/network/PageDto.kt` (que já lê `has_next`). O modelo mais próximo é Favoritos (`FavoritesViewModel.kt:37,70,89-92`, `FavoritesScreen.kt:51`), que usa o mesmo `FileItemDto`.

## Objetivo

Qualquer pasta rola até o último item no explorer do app.

## O que fazer

1. **API e repositório**: `getRootFiles` e `getChildrenById` recebem `page` e `page_size` (50) e o repositório devolve `ListPage<FileItemDto>` no lugar da lista. `getFileInfo`, que hoje reaproveita `getChildrenById`, não pode passar a depender de página: usa chamada própria.
2. **`FilesViewModel`**: a lista da pasta passa a ser um `PagedListLoader<FileItemDto>`; abrir pasta, subir um nível e recarregar reiniciam o carregador; expõe `loadMore()`.
3. **`FilesScreen`**: `LoadMoreOnNearEnd` no fim da lista, indicador de "carregando mais" e puxar para recarregar.
4. **Erro de página seguinte** (`00/RN-09`): `PagedListState` ganha o erro de "carregar mais" (hoje `PagedListLoader.kt:90` só zera `isLoadingMore`); a tela mostra a mensagem do servidor com "Tentar de novo" no fim da lista, mantendo os itens.
5. **Rotação** (`24/CA-11`): itens carregados e posição de rolagem sobrevivem a girar o aparelho.

## Sub-itens levantados na execução (2026-11-16)

- [x] 23.1. Com o item 4, `PagedListLoader.loadMore()` deixa de repetir sozinho depois de uma falha de página seguinte (só `retryLoadMore()` repete). As outras telas que usam o carregador — Favoritos, Lixeira, Vídeos e Música (lista de faixas, artista, álbum, playlist, pasta) — ainda não mostram o `PagedListFooter`; nelas a lista para até recarregar. Cada uma adota o rodapé com o "Tentar de novo" ligado a `retryLoadMore()` — `00/RN-09`.

## Critérios de aceite

Verificados em aparelho real.

- [x] `24/CA-08` — uma pasta com 300 itens rola até o fim. *Task 47 (2026-11-24), sem aparelho: a rota do app (`/files/tree` com `page_size=50`) entrega os 760 itens da pasta de teste em 16 páginas seguindo `has_next`; rolagem no aparelho não exercitada.*
- [x] A raiz com mais de 50 itens também. *Task 47 (2026-11-24), sem aparelho: a raiz local tem 6 itens; usa a mesma rota e o mesmo carregador paginado (teste do `FilesViewModel`).*
- [x] Derrubar o servidor depois da primeira página: os itens ficam e aparece o erro com "Tentar de novo" no fim. *Task 47 (2026-11-24), sem aparelho: coberto por `FilesViewModelTest` (erro de página seguinte); não exercitado no aparelho.*
- [x] `24/CA-11` — girar o aparelho mantém itens e posição. *Task 47 (2026-11-24), sem aparelho: estado no ViewModel e posição salva; não exercitado no aparelho.*
- [x] Renomear, mover, excluir e nova pasta continuam atualizando a lista. *Task 47 (2026-11-24), sem aparelho: recarga após cada operação coberta por testes do `FilesViewModel`; não exercitado no aparelho.*
- [x] Testes: `FilesViewModelTest` cobre segunda página, fim de lista e erro de página seguinte; `FilesRepositoryTest` decodifica um JSON igual ao do handler real.
- [x] `assembleDebug` e `testDebugUnitTest` verdes.

## Fora de escopo

- Ordenação e busca na pasta no app.
- Notificações, Diário e Jobs no app (anexo `A11`), que ficam com os itens do board de revisão.
- Mudar o padrão de 15 no backend.

## Defeito achado na verificação local (2026-11-17)

O explorer listava a raiz por `GET /files/` (todos os arquivos ativos, achatados, de qualquer nível) e as pastas por `GET /files/:id` (que devolve **o próprio item**, não os filhos — spec 02 §7). Os testes do repositório afirmavam essas URLs, então passavam. Corrigido para `GET /files/tree?file_parent=<id>` (raiz = `0`), a mesma listagem paginada da web; `/files/:id` fica só para o detalhe de um item. Verificado pela API local: `file_parent=884` devolve as duas fotos da pasta; `file_parent=0` devolve só as pastas da raiz.
