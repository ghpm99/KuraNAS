# 09 — Lixeira (exclusão recuperável)

**Tipo:** feature de maturidade · **Prioridade:** P2

## Contexto

Hoje a exclusão pela UI é **irreversível e imediata**: `DeleteFileFromDisk` (`internal/api/v1/files/operations.go`) faz `os.RemoveAll` no path. O `deleted_at` que existe no schema é só o rastro de "arquivo sumiu do disco" usado pelo scan (`mark_deleted`) — não há como restaurar conteúdo, porque os bytes se foram.

Todo NAS maduro (Synology `#recycle`, TrueNAS, Nextcloud trash) trata exclusão como operação recuperável, porque é o erro de usuário mais comum e mais destrutivo em um servidor de arquivos. Para um produto cujo trabalho é guardar arquivos, deletar para sempre com um clique — hoje acessível a qualquer dispositivo da rede (whitelist da task 04 ainda pendente) — é o maior risco de perda de dados do sistema.

## Objetivo

Excluir pela UI move o item para uma lixeira no próprio disco; o usuário consegue listar, restaurar e esvaziar; itens antigos são expurgados automaticamente por política de retenção.

## O que fazer

1. Trocar o `os.RemoveAll` do delete por mover para um diretório de lixeira.
2. Registrar os itens na lixeira (path original, data) para permitir restauração.
3. Endpoints: listar lixeira, restaurar item, excluir definitivamente, esvaziar.
4. Expurgo automático por idade (job periódico).
5. UI da lixeira no frontend.

## Como fazer

- **Armazenamento**: diretório `.kuranas-trash/` na raiz do `ENTRY_POINT` (excluído da indexação — o walker do scan e o watcher devem ignorá-lo explicitamente). Mover com `os.Rename` (mesmo volume, custo zero); nome no destino com sufixo único para evitar colisão.
- **Registro**: nova tabela `trash_item` via migração (`pkg/database/migrations/queries/`): `id`, `original_path`, `trash_path`, `deleted_at`, `size`. A linha de `home_file` correspondente é marcada `deleted_at` (a subárvore inteira — ver task 05).
- **Domínio**: seguir o padrão do projeto — pacote `internal/api/v1/trash/` (`handler.go`, `service.go`, `repository.go`, `interfaces.go`, `model.go`, `dto.go`), queries em `pkg/database/queries/trash/`, contexto em `context.go`, rotas em `routes.go`.
- **Endpoints** (pequenos, um recurso por endpoint, conforme a regra do projeto):
  - `GET /api/v1/trash` — lista paginada;
  - `POST /api/v1/trash/:id/restore` — `os.Rename` de volta (409 se o path original voltou a existir);
  - `DELETE /api/v1/trash/:id` — expurgo definitivo de um item;
  - `DELETE /api/v1/trash` — esvaziar.
- **Delete atual**: `DeleteFileFromDisk` passa a mover para a lixeira por padrão. Manter um parâmetro explícito (`?permanent=true`) para exclusão definitiva direta.
- **Retenção**: job periódico (orquestrador, novo job type `trash_purge` ou aproveitando o agendador existente) expurgando itens com mais de N dias (config na tabela de configuração; default sugerido: 30).
- **Restauração no índice**: ao restaurar, limpar `deleted_at` das linhas de `home_file` da subárvore e disparar `ScanDirTask` do destino.
- **i18n**: todas as mensagens/títulos novos nos dois catálogos.
- **Testes**: service com diretório temporário (mover, restaurar, colisão de nome, restaurar com path ocupado), expurgo por idade, handlers.

## Critérios de aceite

- [x] Excluir arquivo/pasta pela UI não remove bytes do disco; o item aparece em `GET /api/v1/trash`. *(DELETE /files/path default move para a lixeira; UI inalterada no fluxo de excluir + página /trash nova lista os itens)*
- [x] Restaurar devolve o item ao path original e ele reaparece na árvore de arquivos. *(os.Rename de volta + RestoreSubtree revive as linhas + ScanDirTask; a UI invalida as queries de files)*
- [x] Restaurar com o path original ocupado responde conflito (409) com mensagem i18n. *(ErrRestoreConflict → 409 TRASH_RESTORE_CONFLICT; a UI exibe a mensagem do backend verbatim)*
- [x] `.kuranas-trash/` não aparece na árvore, nas abas de mídia nem nos analytics. *(o conteúdo da lixeira nunca vira linha de `home_file`: o walker do diff pula o dir (`SkipDir`), o watcher fsnotify não o observa nem emite eventos dele, e o scanner de watch folders o ignora — árvore, mídia e analytics leem só `home_file`)*
- [x] Expurgo automático remove itens além da retenção configurada. *(trash.Purger: roda no boot e a cada 12h; retenção em app_settings via GET/PUT /trash/retention, default 30 dias)*
- [x] `?permanent=true` mantém o comportamento destrutivo atual para quem quiser. *(query param no DELETE /files/path; sem lixeira configurada o delete padrão recusa em vez de destruir)*
- [ ] `make ci` verde (backend + frontend).

## Fora de escopo

- Versionamento de arquivos/snapshots (mudou-salvou não gera versão).
- Quotas de espaço da lixeira (só retenção por idade nesta task).
- Lixeira por usuário (o sistema é single-user — ver task 04).
- Adaptação dos apps Android/plugin à nova semântica (a API antiga continua funcionando; `permanent` é opt-in).
