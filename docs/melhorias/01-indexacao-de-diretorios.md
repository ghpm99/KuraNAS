# 01 — Indexar diretórios no pipeline de scan

**Tipo:** bug crítico · **Prioridade:** P0

## Contexto

A aba de arquivos navega por **linhas de pasta** no banco: `GetFilesTreeHandler` (`backend/internal/api/v1/files/listing.go`) busca a linha da pasta pelo id e lista os filhos por `parent_path = <path da pasta>`. Para uma pasta existir na árvore, ela precisa ter uma linha própria em `home_file`.

Só que **nenhum caminho moderno de indexação cria linhas de diretório** — todos pulam pastas explicitamente:

- `executeDiffAgainstDBStep` (`internal/worker/engine/step_executors.go`): `if d.IsDir() { return nil }`. Esse passo serve `startup_scan`, `fs_event` e os rescans disparados pelas operações de arquivo.
- `dispatchWatcherChanges` (`internal/worker/engine/watcher.go`): `if snap.IsDir { continue }`.

O único código que inseria diretórios é o pipeline legado (`scan.ScanFilesWorker`, que faz `filepath.Walk` incluindo pastas), mas ele só roda quando `JobOrchestrator == nil` — ou seja, nunca em produção. As pastas que aparecem hoje na árvore são resquício da época em que o legado rodava.

**Sintoma observado:** pasta nova criada dentro da pasta de músicas, cheia de arquivos — as músicas aparecem na aba de músicas (as queries de música derivam as "pastas" do `parent_path` dos próprios arquivos de áudio, ver `pkg/database/queries/music/get_music_folders.sql`), mas a pasta não aparece na aba de arquivos e seu conteúdo fica inalcançável pela navegação.

O mesmo defeito atinge a própria UI: `CreateFolder` (`internal/api/v1/files/operations.go`) cria a pasta no disco e dispara `ScanDirTask`, que cai no mesmo diff que pula diretórios — pasta criada pelo KuraNAS também nunca ganha linha no banco. Mover pasta é pior: `mark_deleted` marca a linha antiga como deletada e o diff nunca cria a nova.

## Objetivo

Toda pasta existente no disco sob o `ENTRY_POINT` tem uma linha correspondente (não deletada) em `home_file`, criada/atualizada pelos mesmos fluxos que já indexam arquivos — sem reiniciar o servidor nem rescan manual.

## O que fazer

1. Fazer o passo `diff_against_db` indexar diretórios além de arquivos.
2. Fazer o watcher de entry point tratar diretórios novos/renomeados em vez de descartá-los.
3. Garantir que instalações existentes convergem sozinhas (backfill via `startup_scan`).

## Como fazer

- **`executeDiffAgainstDBStep`**: remover o `if d.IsDir() { return nil }`. Para diretórios, fazer o lookup por path (`GetFileStatByPath`); se não existir linha, **upsert direto via `FilesService`** (criar o `FileDto` com `ParseFileInfoToFileDto`, que já distingue tipo diretório) em vez de enfileirar o plano completo de processamento — diretório não tem metadata/thumbnail/playlist. Se preferir manter tudo via job, criar um plano só com o step `persist`; o upsert direto é a opção mais simples e suficiente.
- **`dispatchWatcherChanges`**: trocar o `if snap.IsDir { continue }` por persistência da linha do diretório (mesmo upsert). Diretório removido já é coberto pelo job `mark_deleted` existente (ele lista linhas do banco por prefixo e faz `os.Stat`, que funciona para pastas).
- **Comparação de mudança**: para diretórios, mtime/size não são sinal confiável de mudança de conteúdo; basta garantir existência da linha e nome/path corretos. Não recalcular checksum de diretório nesse fluxo.
- **Backfill**: nenhuma migração necessária — o `startup_scan` do próximo boot percorre a árvore e cria as linhas de pasta que faltam. Confirmar isso num teste de integração.
- **Testes**: estender `diff_step_pg_integration_test.go` e `watcher_test.go` cobrindo: pasta nova → linha criada; pasta movida → linha nova criada e antiga marcada deletada; árvore aninhada (pasta dentro de pasta nova).

## Critérios de aceite

- [ ] Criar uma pasta no disco (fora da UI) dentro do `ENTRY_POINT` → ela aparece na aba de arquivos no próximo ciclo do watcher, com o conteúdo navegável.
- [ ] Criar uma pasta pela UI (`CreateFolder`) → ela aparece na árvore sem reiniciar o servidor.
- [ ] Mover uma pasta pela UI → ela aparece no destino e some da origem na árvore.
- [ ] Reproduzir o cenário original (pasta nova com músicas) → pasta e arquivos visíveis na aba de arquivos **e** na aba de músicas.
- [ ] Após um boot em base existente, pastas que faltavam ganham linha (backfill via `startup_scan`).
- [ ] Testes de integração cobrindo diff e watcher com diretórios; `make ci-backend` verde (cobertura ≥ 80%).

## Fora de escopo

- Remodelar a hierarquia para `parent_id` (hoje é por string `parent_path`) — fica como evolução futura.
- Trocar o watcher de polling por eventos nativos (task 06).
- Atualização síncrona do banco nas operações de arquivo (task 05) — esta task só garante que o caminho assíncrono converge.
- Remoção do pipeline legado (task 07).
