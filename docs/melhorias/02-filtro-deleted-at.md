# 02 — Consertar o filtro `deleted_at` da listagem de arquivos

**Tipo:** bug · **Prioridade:** P0

## Contexto

Em `pkg/database/queries/files/get_files.sql`, o filtro de exclusão lógica é:

```sql
AND (
    $15
    OR hf.deleted_at = $16
)
```

Não existe ramo `deleted_at IS NULL`. As consequências:

- Quando o chamador passa `DeletedAt{HasValue: false}` querendo dizer "só ativos" (caso do `GetFilesTreeHandler` em `internal/api/v1/files/listing.go`), `$15 = true` simplesmente **desliga o filtro** — arquivos soft-deletados aparecem na árvore e nas listagens.
- Quando passa `HasValue: true`, a query compara `deleted_at = <timestamp exato>`, que na prática nunca casa nada. É por isso que `FindFilesDeleted` (`internal/worker/scan/files.go`) é inócua: ela consulta com `HasValue: true` e valor zero → `deleted_at = '0001-01-01'` → conjunto vazio.

O `utils.Optional[time.Time]` não consegue expressar as três intenções necessárias: "ignorar", "só ativos" (`IS NULL`) e "só deletados" (`IS NOT NULL`). As queries de música não sofrem disso porque têm `deleted_at IS NULL` fixo — um dos motivos de a aba de música parecer mais correta que a de arquivos.

## Objetivo

A listagem de arquivos respeita o estado de exclusão lógica: por padrão a árvore e as listagens só retornam arquivos ativos, e fluxos internos conseguem pedir explicitamente "só deletados" ou "todos".

## O que fazer

1. Trocar o filtro de `deleted_at` no `FileFilter` por um tri-state explícito.
2. Atualizar `get_files.sql` para os ramos `IS NULL` / `IS NOT NULL` / ignorar.
3. Revisar todos os chamadores e mapear a intenção real de cada um.

## Como fazer

- **Modelo**: substituir `DeletedAt utils.Optional[time.Time]` no `FileFilter` (`internal/api/v1/files/model.go`) por um enum, p.ex. `DeletedFilter` com `DeletedFilterAny` / `DeletedFilterOnlyActive` / `DeletedFilterOnlyDeleted`. Passar para a query como um parâmetro de texto e resolver com `CASE`, no mesmo estilo do filtro `category` que já existe na query:

```sql
AND CASE $15
    WHEN 'active'  THEN hf.deleted_at IS NULL
    WHEN 'deleted' THEN hf.deleted_at IS NOT NULL
    ELSE TRUE
END
```

- **Chamadores a revisar** (busca por `FileFilter{`):
  - `GetFilesTreeHandler` e demais handlers de listagem → `OnlyActive`.
  - `executeMarkDeletedStep` (`internal/worker/engine/step_executors.go`) → `Any` (ele precisa ver deletados para poder restaurá-los — hoje funciona por acidente; deixar a intenção explícita).
  - `FindFilesDeleted` (legado) → será removida na task 07; até lá, corrigir ou deixar documentado que é inócua.
- **Atenção ao posicionamento dos parâmetros**: a query usa placeholders posicionais (`$1..$19`); a mudança altera a aridade — ajustar `Repository.GetFiles` (`listing.go`) e os testes sqlmock em conjunto.
- **Testes**: caso de repositório com linha deletada + filtro `OnlyActive` → não retorna; `OnlyDeleted` → retorna; integração no fluxo da árvore.

## Critérios de aceite

- [x] Arquivo marcado com `deleted_at` não aparece na aba de arquivos (árvore, listagem por path, children). *(handlers de listagem passam `DeletedFilterOnlyActive`; `TestPostgres_DeletedFilterTriState` prova o ramo `IS NULL` contra Postgres real)*
- [x] O fluxo de restauração do `mark_deleted` (arquivo reaparece no disco → `deleted_at` limpo) continua funcionando, com teste cobrindo. *(`executeMarkDeletedStep` agora declara `DeletedFilterAny`; `TestMarkDeletedStep_RestoresReappearedFile_Postgres`)*
- [x] Não existe mais comparação `deleted_at = <timestamp>` em query nenhuma de `queries/files/`. *(só restam `IS NULL`/`IS NOT NULL` e colunas de escrita)*
- [x] Todos os construtores de `FileFilter` declaram explicitamente a intenção do filtro de deletados. *(listing/handlers → OnlyActive; GetFileById/GetFileByNameAndPath/mark_deleted → Any com comentário; legado scan → Any preservando comportamento, FindFilesDeleted → OnlyActive e deixa de ser inócua)*
- [x] `make ci-backend` verde (cobertura ≥ 80%). *(make ci completo verde em 2026-06-11, integração contra Postgres 18 local)*

## Fora de escopo

- Lixeira/restauração para o usuário final (task 09) — aqui é só o filtro de dados.
- Decompor a mega-query `get_files.sql` em queries menores (task 08) — esta task mexe só no ramo do `deleted_at`, na aridade dos parâmetros e nos chamadores.
- Remoção do código legado (`FindFilesDeleted`) — task 07.
