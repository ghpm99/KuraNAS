# 13 — Tiering quente/frio com path lógico × localização física

**Tipo:** feature (visão de armazenamento) · **Prioridade:** P3 · **Depende de:** 01, 05, 10

## Contexto

Visão de armazenamento registrada (2026-06-11): o SSD de 512 GB guarda os arquivos ativos (tier quente) e o volume frio (pool de HDs espelhado via Storage Spaces — ver task 10) guarda os pouco usados, aumentando a capacidade útil sem o usuário gerenciar nada manualmente.

O insumo para a política **já existe no banco**: `home_file.last_interaction` e a tabela `recent_file` registram uso. "Arquivos não tocados há N dias" é uma query.

Decisão de design (a parte importante): a migração é **transparente** — o arquivo continua aparecendo no mesmo lugar da árvore lógica (`Documentos/Faculdade/...`), mas seus bytes moram no volume frio. Isso é viável justamente pela arquitetura do sistema: **todo acesso a conteúdo passa pela API** (blob, stream, download, thumbnail), então basta separar o *path lógico* (identidade do arquivo, o que a navegação mostra) da *localização física* (onde os bytes estão). A alternativa — migração visível, arquivo "mudando de pasta" sozinho — mistura organização lógica com política de armazenamento e foi descartada.

## Objetivo

Arquivos sem interação há N dias (configurável) migram automaticamente para o tier frio sem mudar de lugar na árvore, na busca ou nas abas de mídia; abrir/baixar/tocar um arquivo frio funciona normalmente (apenas mais lento); o índice nunca diverge por causa da migração.

## O que fazer

1. Separar path lógico de localização física no modelo (`physical_path` nullable em `home_file`).
2. Fazer todos os acessos a conteúdo resolverem a localização física.
3. **Blindar o scan/watcher/mark_deleted**: a existência de um arquivo tiered é verificada no path físico, e a área fria não é indexada como raiz navegável.
4. Job de migração (rebaixamento automático por idade; promoção de volta ao quente quando o arquivo voltar a ser usado).
5. Configuração + visibilidade na UI (tier do arquivo, espaço por tier).

## Como fazer

- **Modelo**: migração adicionando `home_file.physical_path TEXT NULL` — `NULL` significa "os bytes estão no próprio `path`" (caso de 100% dos arquivos hoje; nenhum backfill necessário). Área fria: diretório configurado no volume frio (ex.: `F:\kuranas-cold\`), espelhando a estrutura relativa da raiz de origem.
- **Resolução de conteúdo**: criar um único helper no domínio `files` (ex.: `ResolveContentPath(file) string`) e usá-lo em **todos** os pontos que abrem o arquivo no disco: blob/download, stream de vídeo/música, geração de thumbnail, checksum, metadata. Buscar por `os.Open`/`os.Stat` nesses fluxos para não esquecer nenhum.
- **Integração com o índice (o ponto crítico)**: para o scan, um arquivo migrado *sumiu* do path lógico no SSD — sem blindagem, o `mark_deleted` o marcaria deletado e o watcher dispararia eventos falsos. Regras:
  - `executeMarkDeletedStep` e qualquer verificação de existência usam `physical_path` quando preenchido;
  - o diretório frio entra na lista de paths ignorados pelo walker/watcher (mesma lista criada nas tasks 09/12) — ele nunca vira raiz navegável nem gera linhas próprias;
  - o diff (`diff_against_db`) não re-enfileira arquivo tiered como "novo" nem como "sumido".
  - Testes de integração cobrindo exatamente esses três cenários são obrigatórios.
- **Job de migração**: novo `JobType` `tier_migration` no orquestrador, agendado (default diário, madrugada). Rebaixamento: query por `last_interaction < now() - N dias` (e tamanho mínimo configurável — não vale a pena migrar arquivos minúsculos); para cada arquivo: copiar para o frio com verificação de checksum → atualizar `physical_path` em transação → remover do quente. Nessa ordem — falha no meio deixa no máximo uma cópia extra, nunca zero cópias.
- **Promoção**: quando um arquivo frio registra interação (o sistema já grava `last_interaction`/`recent_file`), o job da noite o traz de volta ao quente (operação inversa, mesma verificação). Promoção síncrona no momento do acesso fica fora — o acesso serve direto do frio, só mais lento.
- **Pressão de espaço**: gatilho adicional opcional — se o espaço livre do SSD cair abaixo de um limiar configurável, o job aperta o critério (migra os menos usados até voltar ao limiar).
- **Operações de arquivo**: move/rename/delete (task 05) operam sobre o path lógico; quando `physical_path` está preenchido, rename lógico não toca nos bytes (só atualiza `path`), delete/lixeira removem do path físico. Mapear esses casos em `operations.go`.
- **UI**: badge/indicador de tier no detalhe do arquivo e, nos analytics, espaço usado por tier (endpoints pequenos, um por pergunta, conforme a regra do projeto). Config (idade, tamanho mínimo, limiar de espaço, diretório frio) na tela de Settings. i18n em tudo.
- **Backup (task 12)**: o backup lê via `ResolveContentPath` — arquivo frio é backupado normalmente; nenhuma dependência de ordem além disso.

## Critérios de aceite

- [x] Arquivo sem interação além do limiar migra para o frio no job noturno e **não muda de lugar** na árvore, busca e abas de mídia. *(job `tier_migration`: `ListDemotionCandidates` por `last_interaction`/tamanho + `tieringengine.Run`; o path lógico nunca é tocado, só `physical_path`.)*
- [x] Abrir/baixar/tocar/gerar thumbnail de arquivo frio funciona (resolução via `physical_path`).
- [x] Scan completo + watcher rodando com arquivos tiered: nenhum é marcado deletado, duplicado ou re-enfileirado (teste de integração). *(`TestMarkDeletedStep_KeepsTieredFileActive_Postgres`, `TestMarkDeletedStep_TieredFileSurvivesWatcherRemoveEvent_Postgres`, `TestDiffStep_IgnoresTieredFile_Postgres`)*
- [x] Arquivo frio que volta a ser usado é promovido ao quente no ciclo seguinte. *(`ListPromotionCandidates` com o mesmo cutoff simétrico; promoções rodam antes das demoções no mesmo passe.)*
- [x] Migração interrompida no meio (kill do processo) nunca perde arquivo: ou está no quente, ou no frio com `physical_path` consistente (recovery do orquestrador + ordem copiar→atualizar→remover). *(ordem testada em `TestRun_DemotionDbFailureKeepsHotCopy`; jobs `running` voltam a `queued` no `recoverInterruptedWork`.)*
- [x] Rename/move/delete lógicos funcionam para arquivos tiered (incluindo lixeira). *(rename/move viram operação só-lógica quando `physical_path` está preenchido; delete/lixeira agem no path físico via `MoveToTrashFrom`. Testes: `TestRenameFileTieredLeavesColdBytesAndUpdatesPath`, `TestMoveFileTieredUpdatesPathWithoutHotCopy`, `TestDeleteFileFromDiskTiered{TrashesColdCopy,PermanentRemovesColdCopy}`.)*
- [x] UI mostra tier do arquivo e espaço por tier; parâmetros configuráveis em Settings. *(`TieringSettingsSection` na tela de Settings: config + chips de uso quente/frio de `GET /tiering/usage`; badge "volume frio" no detalhe do arquivo via `tier` no `FileDto`.)*
- [x] `make ci` verde (backend + frontend). *(backend 80,1% ≥ 80%; frontend 1272 testes, thresholds 89/90 atingidos; build + typecheck:test ok.)*

## Fora de escopo

- Promoção síncrona no acesso (servir do frio é aceitável; promoção é assíncrona).
- Mais de dois tiers, cache de leitura, write-back.
- Pin manual de arquivo/pasta no quente ("nunca migrar") — evolução natural, registrar como ideia.
- Tiering de diretórios inteiros como unidade (a unidade é o arquivo).
- Redundância do volume frio — é do Storage Spaces (decisão na task 10).
