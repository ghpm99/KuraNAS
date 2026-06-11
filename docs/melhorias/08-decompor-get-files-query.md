# 08 — Decompor a mega-query `get_files.sql`

**Tipo:** dívida técnica · **Prioridade:** P2

## Contexto

`pkg/database/queries/files/get_files.sql` é uma query única com **9 filtros opcionais** no padrão `($n OR coluna = $m)` (id, name, path, path-prefix, parent_path, format, type, deleted_at, category) atendendo a todos os chamadores de `Repository.GetFiles` via o struct `FileFilter`.

Isso contraria a regra do próprio projeto (`backend/CLAUDE.md` → "One small, optimized query per repository call... No god-query") e tem custos concretos:

- **Bugs se escondem**: o filtro quebrado de `deleted_at` (task 02) viveu despercebido dentro dela.
- **O planner não otimiza**: com todos os predicados presentes em todas as execuções, o Postgres não consegue escolher plano/índice bom por caso de uso; um plano genérico serve listagem por parent, lookup por id e varredura por prefixo igualmente mal.
- **Intenção ilegível**: pelo call site (`FileFilter{...}`) não dá para saber qual pergunta está sendo feita; cada chamador usa 1–2 filtros de 9.

A análise dos chamadores mostra poucas perguntas reais: filhos por `parent_path` (árvore), lookup por `id`, lookup por `path`, varredura por prefixo (`mark_deleted`), busca por nome.

## Objetivo

Cada pergunta ao banco tem seu próprio `.sql` pequeno e seu próprio método de repositório, alinhado à regra "one small, optimized query per repository call" — sem mudança de contrato HTTP.

## O que fazer

1. Mapear todos os chamadores de `GetFiles`/`FileFilter` e a combinação de filtros que cada um realmente usa.
2. Criar um `.sql` + método de repositório por pergunta.
3. Migrar os chamadores e aposentar `get_files.sql` e o `FileFilter` genérico (ou reduzi-lo ao que sobrar).

## Como fazer

- Queries previstas (nomes ilustrativos), todas em `pkg/database/queries/files/` com `//go:embed` em `files.go`, todas filtrando deletados explicitamente (depende da task 02 para a semântica tri-state onde necessário):
  - `get_children_by_parent_path.sql` — árvore/listagem (paginada, só ativos);
  - `get_file_by_id.sql` — lookup pontual;
  - `get_file_by_path.sql` — lookup pontual (já existe `get_file_stat_by_path.sql` como referência de estilo);
  - `get_files_by_path_prefix.sql` — varredura paginada do `mark_deleted` (todos os estados);
  - `search_files_by_name.sql` — busca por nome (avaliar se o domínio `search` já cobre; se sim, não duplicar).
- Métodos correspondentes em `Repository` (ex.: `GetChildrenByParentPath`, `GetFileById`, ...), cada um com assinatura explícita em vez do struct de filtros. O service expõe métodos equivalentes; os handlers migram um a um.
- Ordenação e paginação: manter `ORDER BY type, name, id DESC` e o padrão `LIMIT pageSize+1` + `UpdatePagination` para não mudar o shape das respostas.
- Migração incremental: uma query nova por commit, com o chamador migrado e testado; `get_files.sql` só é apagado quando ficar sem chamadores.
- Aproveitar para conferir índices: `parent_path`, `path` (lookup e prefixo) — criar migração de índice se faltar (ex.: índice b-tree em `parent_path`, `text_pattern_ops`/`starts_with`-friendly em `path`).
- **Testes**: os testes sqlmock existentes (`repository_light_test.go`) se decompõem junto; integração pg (`repository_pg_integration_test.go`) cobre as queries novas.

## Critérios de aceite

- [ ] Cada `.sql` novo responde exatamente uma pergunta e não tem filtros opcionais no padrão `($n OR ...)`.
- [ ] Nenhuma mudança de contrato HTTP (paths, params e shapes idênticos — validado pelos testes de handler existentes).
- [ ] `get_files.sql` removido ou reduzido a um único caso real remanescente, documentado.
- [ ] Índices verificados para `parent_path` e lookup/prefixo de `path` (migração criada se necessário).
- [ ] `make ci-backend` verde (cobertura ≥ 80%).

## Fora de escopo

- A correção do `deleted_at` em si (task 02 — pré-requisito conceitual, pode ser feita antes ou junto da primeira query migrada).
- Mudar ordenação, paginação ou shape de resposta de qualquer endpoint.
- Refatorar queries de outros domínios (music/video/image/analytics).
