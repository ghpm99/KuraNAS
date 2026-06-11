# 05 — Operações de arquivo atualizam o banco na própria transação

**Tipo:** consistência · **Prioridade:** P1

## Contexto

As operações de arquivo (`internal/api/v1/files/operations.go` — `CreateFolder`, `MoveFile`, `RenameFile`, `DeleteFileFromDisk`, `CopyFile`, upload) mexem no **disco** e depois disparam `ScanDirTask` para que um rescan assíncrono conserte o **banco**. Problemas:

- **Janela de inconsistência**: entre a resposta HTTP e a conclusão do job, a UI lê dado velho — o arquivo movido ainda aparece no lugar antigo, o renomeado com o nome antigo.
- **A operação já sabe exatamente o que mudou** (path de origem, path de destino), mas joga essa informação fora e delega a um diff que precisa redescobrir tudo varrendo o disco.
- **Diretórios nunca convergem** no rescan (ver task 01): mover/renomear pasta deixa as linhas dos descendentes com paths velhos até o próximo scan completo — e a linha da própria pasta nunca é recriada.

A hierarquia é modelada por strings (`path`, `parent_path` absolutos, sem `parent_id`), então mover/renomear uma pasta implica reescrever o prefixo de path de todos os descendentes.

## Objetivo

Imediatamente após a resposta de uma operação de arquivo, uma leitura na API reflete o novo estado — o rescan assíncrono vira rede de segurança, não mecanismo primário de sincronização.

## O que fazer

Para cada operação, atualizar as linhas afetadas em `home_file` na mesma transação (via `DbContext.ExecTx`), mantendo o `ScanDirTask` existente como reconciliação.

## Como fazer

- **`CreateFolder`**: após o `os.Mkdir`, inserir a linha do diretório (`CreateFile`).
- **`RenameFile`** e **`MoveFile`**: após o `os.Rename`, em uma transação:
  1. atualizar `name`/`path`/`parent_path` da linha movida;
  2. se for diretório, atualizar os descendentes com um único `UPDATE` de troca de prefixo, em novo `.sql` em `pkg/database/queries/files/` (ex.: `update_descendant_paths.sql`):
     ```sql
     UPDATE home_file
     SET path = $2 || substr(path, length($1) + 1),
         parent_path = $2 || substr(parent_path, length($1) + 1)
     WHERE starts_with(path, $1 || '/') OR parent_path = $1;
     ```
     Usar `starts_with` e concatenação posicional — **não** usar `LIKE` com path (o `\` do Windows é caractere de escape do LIKE; já causou bug, ver comentário em `get_files.sql`). Atenção ao separador: os paths são gravados com separador do SO (`\` em produção Windows), o sufixo do `starts_with` deve usar o separador correto.
- **`DeleteFileFromDisk`**: após o `os.RemoveAll`, marcar `deleted_at` da linha e dos descendentes (mesma técnica de prefixo).
- **`CopyFile`** e **upload**: podem continuar delegando ao pipeline (o conteúdo novo precisa de metadata/checksum/thumbnail de qualquer forma), mas inserir as linhas básicas (arquivo/pasta) sincronamente para a UI já enxergar.
- **Ordem disco → banco**: manter o disco como fonte de verdade; se o update do banco falhar após o `os.Rename`, logar e deixar o `ScanDirTask` reconciliar (não tentar desfazer a operação no disco).
- **Manter o `ScanDirTask`** ao final de cada operação, como hoje — ele segue necessário para metadata/checksum e para corrigir qualquer divergência.
- **Testes**: ampliar `operations_test.go` cobrindo o estado do banco imediatamente após cada operação (sem rodar workers), incluindo mover/renomear pasta com descendentes.

## Critérios de aceite

- [ ] Após `POST` de move/rename/delete/create-folder, um `GET` imediato na árvore reflete o novo estado, **com workers desligados** (`ENABLE_WORKERS=false`).
- [x] Mover/renomear pasta atualiza `path`/`parent_path` de todos os descendentes em uma transação. *(`syncMovedRows`: `UpdateFile` + `UpdateDescendantPaths` no mesmo `withTransaction`; `TestMoveDirectorySyncsRowAndDescendantsInOneTransaction`)*
- [ ] Delete marca `deleted_at` da subárvore inteira.
- [ ] Falha no update do banco não corrompe a operação de disco (operação responde sucesso, log registra, rescan reconcilia).
- [x] Nenhuma query nova usa `LIKE` sobre paths. *(`update_descendant_paths.sql` e `mark_deleted_subtree.sql` usam `starts_with` com prefixo + separador montados no Go)*
- [ ] `make ci-backend` verde (cobertura ≥ 80%).

## Fora de escopo

- Migrar a hierarquia para `parent_id` com FK — mudança de modelagem maior, registrar como ideia futura.
- Mudanças no contrato HTTP das operações (paths, payloads e respostas ficam como estão).
- Indexação de diretórios pelo pipeline assíncrono (task 01) e remoção do legado (task 07).
- Mover os complementos por tipo (`audio_metadata` etc.) — eles referenciam `file_id`, não path, e não precisam de mudança.
