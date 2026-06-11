# Melhorias do sistema — tasks

Backlog de melhorias levantado na análise de maturidade do sistema (2026-06-11). Cada arquivo é **uma task autocontida**: contexto, objetivo, o que/como fazer, critérios de aceite e o que está fora de escopo.

Origem: investigação do bug "pasta nova de músicas aparece na aba de músicas mas não na aba de arquivos", que revelou problemas estruturais na sincronização disco ↔ banco, além de lacunas de maturidade frente a NAS estabelecidos (autenticação, watcher, lixeira, protocolos).

## Regras válidas para todas as tasks

- Cada task termina **verde no `make ci`** antes de ser considerada concluída.
- **O contrato HTTP não muda** salvo quando a task disser explicitamente o contrário (frontend, 2 apps Android e plugin consomem a API).
- Commits lógicos diretos em `develop`, conforme o workflow do projeto.
- Toda string visível ao usuário passa pelo i18n (regra do `CLAUDE.md` raiz).

## Ordem sugerida e status

| # | Arquivo | Tipo | Prioridade | Status |
|---|---|---|---|---|
| 01 | [01-indexacao-de-diretorios.md](01-indexacao-de-diretorios.md) | bug crítico | P0 | pendente |
| 02 | [02-filtro-deleted-at.md](02-filtro-deleted-at.md) | bug | P0 | pendente |
| 03 | [03-debounce-watcher-perde-eventos.md](03-debounce-watcher-perde-eventos.md) | bug | P1 | pendente |
| 04 | [04-whitelist-de-ips.md](04-whitelist-de-ips.md) | segurança | P0 | pendente |
| 05 | [05-operacoes-atualizam-banco-sincronamente.md](05-operacoes-atualizam-banco-sincronamente.md) | consistência | P1 | pendente |
| 06 | [06-watcher-por-eventos-fsnotify.md](06-watcher-por-eventos-fsnotify.md) | performance | P2 | pendente |
| 07 | [07-remover-pipeline-legado.md](07-remover-pipeline-legado.md) | dívida técnica | P2 | pendente |
| 08 | [08-decompor-get-files-query.md](08-decompor-get-files-query.md) | dívida técnica | P2 | pendente |
| 09 | [09-lixeira.md](09-lixeira.md) | feature | P2 | pendente |
| 10 | [10-multiplas-raizes.md](10-multiplas-raizes.md) | feature | P3 | pendente |
| 11 | [11-acesso-webdav.md](11-acesso-webdav.md) | feature | P3 | pendente |

> Dependências fortes: a 01 e a 02 destravam a confiabilidade básica da aba de arquivos. A 04 vem antes de qualquer exposição fora da máquina local. A 05 reduz o trabalho da 06. A 07 fica mais segura depois da 01 (o único código que indexava diretórios mora no legado). A 11 depende da 04.
