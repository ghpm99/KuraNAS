# Melhorias do sistema — board de execução

Backlog de melhorias levantado na análise de maturidade do sistema (2026-06-11), estendido com a visão de armazenamento do dono (tiering + backup, 2026-06-11). Cada arquivo é **uma task autocontida**: contexto, objetivo, o que/como fazer, critérios de aceite e fora de escopo.

**Este README é a fonte de verdade do andamento.** Um agente apontado para este arquivo deve conseguir continuar o trabalho de onde parou, mesmo após interrupção — o protocolo abaixo existe para isso.

## Protocolo de execução (para agentes)

1. **Leia o board.** Se existe task `em execução`, **retome-a**: abra o arquivo da task, veja quais critérios de aceite já estão marcados (`[x]`), confira no `git log` o que já foi commitado para ela, e continue do primeiro critério desmarcado. Não recomece do zero o que os checkboxes dizem que está feito — mas verifique que o último commit compila (`make ci`) antes de confiar.
2. **Se não há task em execução**, pegue a primeira `pendente` de cima para baixo cuja coluna **Depende de** esteja toda `concluída`.
3. **Ao iniciar**: mude o status dela para `em execução` neste README e commite essa mudança isolada (`docs(melhorias): inicia task NN`). É esse commit que marca o estado para uma retomada futura.
4. **Durante**: trabalhe em commits lógicos (workflow do projeto). Ao satisfazer um critério de aceite, **marque o checkbox `[x]` no arquivo da task no mesmo commit do código** que o satisfaz — os checkboxes são o progresso fino que permite retomar no meio.
5. **Ao concluir**: todos os checkboxes marcados + `make ci` verde → status `✅ concluída (data)` na tabela + commit (`docs(melhorias): conclui task NN`).
6. **Se travar** (decisão pendente do dono, dependência externa): status `🚫 bloqueada` com o motivo na coluna Notas, commite, e passe para a próxima `pendente` elegível.

Regras invariantes:

- **Uma task `em execução` por vez.** Não pular a ordem sem registrar o motivo em Notas.
- Toda task termina **verde no `make ci`**; o **contrato HTTP não muda** salvo a task dizer o contrário (frontend, 2 apps Android e plugin consomem a API); **i18n obrigatório** em toda string visível; commits lógicos diretos em `develop`, sem `Co-Authored-By`.
- Mudança de status é sempre um commit — o board nunca fica só na memória de quem executa.

## Board

| # | Task | Tipo | Prioridade | Depende de | Status | Notas |
|---|---|---|---|---|---|---|
| 01 | [Indexação de diretórios](01-indexacao-de-diretorios.md) | bug crítico | P0 | — | ✅ concluída (2026-06-11) | causa raiz do bug reportado |
| 02 | [Filtro deleted_at](02-filtro-deleted-at.md) | bug | P0 | — | ✅ concluída (2026-06-11) | |
| 03 | [Debounce do watcher perde eventos](03-debounce-watcher-perde-eventos.md) | bug | P1 | — | ✅ concluída (2026-06-11) | |
| 04 | [Whitelist de IPs](04-whitelist-de-ips.md) | segurança | P0 | — | ✅ concluída (2026-06-11) | decisão: sem autenticação |
| 05 | [Operações atualizam banco sincronamente](05-operacoes-atualizam-banco-sincronamente.md) | consistência | P1 | 01, 02 | pendente | |
| 06 | [Watcher por eventos (fsnotify)](06-watcher-por-eventos-fsnotify.md) | performance | P2 | 01, 03 | pendente | |
| 07 | [Remover pipeline legado](07-remover-pipeline-legado.md) | dívida técnica | P2 | 01 | pendente | legado guarda o único exemplo de indexação de dirs |
| 08 | [Decompor get_files query](08-decompor-get-files-query.md) | dívida técnica | P2 | 02 | pendente | |
| 09 | [Lixeira](09-lixeira.md) | feature | P2 | 02, 05 | pendente | |
| 10 | [Múltiplas raízes](10-multiplas-raizes.md) | feature | P3 | 01, 05 | pendente | fundação da visão de armazenamento |
| 11 | [Acesso WebDAV](11-acesso-webdav.md) | feature | P3 | 04 | pendente | melhor após 10 |
| 12 | [Backup orquestrado](12-backup-orquestrado.md) | feature | P3 | 10 | pendente | retenção ≠ espelho |
| 13 | [Tiering quente/frio](13-tiering-quente-frio.md) | feature | P3 | 01, 05, 10 | pendente | path lógico × físico |

Status possíveis: `pendente` · `em execução` · `✅ concluída (AAAA-MM-DD)` · `🚫 bloqueada`.

## Decisões registradas (valem para todas as tasks)

- **Sem autenticação** (2026-06-11): nada de login/senha/token. Controle de acesso é whitelist de IPs (task 04). TLS fora de escopo enquanto o produto for de rede interna.
- **Redundância de disco é do SO, não do app** (2026-06-11): HDs avulsos em pool espelhado via Windows Storage Spaces; o KuraNAS enxerga um volume comum. Detalhe na task 10.
- **Visão de armazenamento** (2026-06-11): SSD = tier quente; pool de HDs = tier frio + backup com retenção; HD externo 2 TB = segunda cópia desconectável (gerida pelo SO). Tasks 10 → 12/13.
- **Backup ≠ espelho**: backup tem retenção de versões; espelho propaga ransomware/exclusão acidental (task 12).
- **Tiering é transparente**: arquivo migrado não muda de lugar na árvore lógica — separação path lógico × localização física (task 13).
