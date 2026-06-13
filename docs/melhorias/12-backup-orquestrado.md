# 12 — Backup orquestrado para o volume frio

**Tipo:** feature (visão de armazenamento) · **Prioridade:** P3 · **Depende de:** 04 (recomendado), 10 (layout de volumes)

## Contexto

Visão de armazenamento registrada (2026-06-11): SSD 512 GB = tier quente (sistema + Postgres + arquivos ativos); pool de HDs espelhado via **Windows Storage Spaces** = tier frio + área de backup; HD externo 2 TB = segunda cópia de backup, idealmente desconectável (proteção contra ransomware). **Redundância de disco é responsabilidade do SO (Storage Spaces), não do KuraNAS** — decisão de infraestrutura registrada na task 10. O KuraNAS cuida de *quais bytes estão onde* (task 13) e de *existir história deles* (esta task).

O schema já antecipa a feature: `home_file.last_backup` existe e nunca é preenchido. O checksum por arquivo (calculado pelo pipeline) dá a detecção de mudança de graça.

Princípio de design: **backup ≠ espelho**. Uma réplica do estado atual propaga ransomware e exclusões acidentais para dentro do backup. O backup precisa de retenção: o que sumiu ou mudou na origem permanece recuperável no backup por N dias.

## Objetivo

Um job periódico copia arquivos novos/alterados das raízes de armazenamento para a área de backup, carimba `last_backup`, mantém versões antigas por período de retenção configurável, e o usuário enxerga a saúde do backup (última execução, pendências, falhas) na UI.

## O que fazer

1. Job de backup no orquestrador (novo job type), agendado por configuração.
2. Cópia incremental (novo/alterado via size/mtime/checksum) com verificação pós-cópia.
3. Retenção: versões substituídas/excluídas vão para área de versões e expiram após N dias.
4. Carimbo de `last_backup` por arquivo + visão de status na UI.
5. A área de backup é invisível para o resto do sistema (scan, watcher, abas, analytics).

## Como fazer

- **Job/steps**: novo `JobType` `backup_run` em `internal/worker/job/job_domain.go` (com `IsValid`), steps no padrão `step_*.go` do `worker/engine/`: `backup_plan` (varre as raízes comparando com o destino e/ou `last_backup`) → `backup_copy` (copia em lotes) — ou um step único iterando, se o volume de código não justificar dois. Agendamento: ticker próprio no worker (configurável, default diário) enfileirando o job; reaproveitar o padrão do `startup_scan`.
- **Destino**: path da área de backup em configuração (tabela `configuration`), com layout `<destino>/current/<path-relativo-da-raiz>` + `<destino>/_versions/<timestamp>/...`. Validar que o destino não está dentro de nenhuma raiz indexada.
- **Cópia segura**: copiar para arquivo temporário no destino + `os.Rename` (atomicidade no mesmo volume); verificar checksum após a cópia antes de carimbar `last_backup`. Em falha, contabilizar e seguir (um arquivo ruim não derruba o job — mesma filosofia dos steps atuais).
- **Detecção de mudança**: comparar size + mtime; em divergência, confirmar por checksum (já existe em `home_file.checksum`). Arquivo com `last_backup` ≥ `updated_at` e presente no destino → pular.
- **Retenção**: antes de sobrescrever ou quando a origem foi excluída, mover a cópia existente para `_versions/<data>/`; step de expurgo remove versões além de `BACKUP_RETENTION_DAYS` (config, default 30).
- **Invisibilidade**: excluir o diretório de backup de `collectEntryPointSnapshot`/walkers/diff/`mark_deleted` (mesmo mecanismo de exclusão criado para `.kuranas-trash` na task 09 — generalizar numa lista de paths ignorados).
- **Status na UI**: seguir a regra de endpoints pequenos — ex.: `GET /api/v1/backup/status` (última execução, contadores) e `GET /api/v1/backup/pending` (quantos arquivos sem backup), cada um com seu `.sql` (`last_backup IS NULL OR last_backup < updated_at`). Notificação ao concluir/falhar via `emitNotification` (i18n).
- **HD externo (segunda cópia)**: fora do job — documentar no README do projeto que a segunda cópia é sincronização do diretório de backup para o HD externo (robocopy/agendador do Windows). O sistema não gerencia mídia desconectável nesta task.
- **Testes**: plan/copy com diretórios temporários (novo, alterado, excluído, retenção expirando), verificação de checksum pós-cópia, exclusão da área de backup no scan, endpoints de status.

## Critérios de aceite

- [x] Job `backup_run` roda no agendamento configurado e copia incrementalmente apenas o que mudou.
- [x] Toda cópia é verificada (checksum) antes de `last_backup` ser carimbado.
- [x] Excluir/alterar um arquivo na origem mantém a versão anterior recuperável em `_versions/` por N dias; expurgo remove além da retenção.
- [x] A área de backup não aparece na árvore, nas abas de mídia, nos analytics, e não dispara o watcher (o destino é validado fora de qualquer raiz indexada — scan/watcher/abas/analytics só enxergam as raízes).
- [x] UI mostra status do backup (última execução, pendentes, falhas) e notificação de conclusão/falha (i18n).
- [x] Interromper o servidor no meio de um backup não corrompe nada (recovery do orquestrador retoma; cópias parciais ficam só no temporário).
- [x] `make ci` verde (backend + frontend).

## Fora de escopo

- Restauração pela UI (recuperação manual pelo filesystem nesta fase; endpoint de restore é evolução).
- Cópia para fora de casa / nuvem (3-2-1 completo) — capítulo futuro.
- Deduplicação, compressão, criptografia do backup.
- Gerência da segunda cópia no HD externo desconectável (fica com o SO).
- Snapshots/versionamento contínuo de arquivo (a retenção aqui é por execução de backup).
