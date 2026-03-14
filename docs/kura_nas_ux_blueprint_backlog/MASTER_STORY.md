# MASTER STORY - KuraNAS UX Blueprint

## Objetivo

Acompanhar a evolucao do KuraNAS do estado atual para o blueprint de UX, mantendo:

- backlog visivel
- status por task
- historico curto do que foi concluido
- proxima entrega recomendada

## Snapshot atual

- Status geral: `IN_PROGRESS`
- Fase atual recomendada: `Fundacao`
- Proxima task recomendada: `TASK-003`

## Quadro de tasks

| ID | Titulo | Status | Fase |
| --- | --- | --- | --- |
| TASK-001 | Migrar IA global e rotas base | DONE | Fundacao |
| TASK-002 | Criar tokens visuais e AppShell unificado | DONE | Fundacao |
| TASK-003 | Entregar Home hub v1 | TODO | Fundacao |
| TASK-004 | Refatorar a estrutura da pagina Arquivos | TODO | Explorer |
| TASK-005 | Tornar a abertura de midia transversal e consistente | TODO | Explorer |
| TASK-006 | Criar Favoritos v1 como dominio proprio | TODO | Explorer |
| TASK-007 | Reorganizar IA e rotas do dominio Musica | TODO | Musica |
| TASK-008 | Entregar Home de musica e contexto de reproducao | TODO | Musica |
| TASK-009 | Normalizar metadados e playlists automaticas de musica | TODO | Musica |
| TASK-010 | Reorganizar IA e rotas do dominio Videos | TODO | Videos |
| TASK-011 | Refinar classificacao e paginas de detalhe de videos | TODO | Videos |
| TASK-012 | Evoluir o player de video para contexto completo | TODO | Videos |
| TASK-013 | Criar pipeline de classificacao persistida para imagens | TODO | Imagens |
| TASK-014 | Reorganizar IA da biblioteca de imagens | TODO | Imagens |
| TASK-015 | Evoluir viewer e acoes de imagem | TODO | Imagens |
| TASK-016 | Separar Analytics em visao geral e biblioteca/indexacao | TODO | Sistema |
| TASK-017 | Criar Settings e consolidar configuracoes | TODO | Sistema |
| TASK-018 | Limpar About e retirar Activity Diary da camada principal | TODO | Sistema |
| TASK-019 | Implementar busca global e remover legado restante | TODO | Sistema |

## Historico

| Data | Item | Resumo | Alteracoes |
| --- | --- | --- | --- |
| 2026-03-13 | STORY-SETUP | Backlog inicial criado a partir do blueprint e do codigo atual. | Adicionados diagnostico, backlog faseado e master story em `docs/kura_nas_ux_blueprint_backlog/`. |
| 2026-03-13 | TASK-001 | Rotas base e navegacao principal migradas para Home, Files, Favorites e Settings. | Frontend recebeu novas rotas `/home`, `/files`, `/favorites` e `/settings`, com redirects de `/` e `/starred`, Home/Settings iniciais, atualizacao da navegacao e i18n compartilhado. |
| 2026-03-13 | TASK-002 | Tema global, tokens visuais e AppShell foram unificados para as paginas principais. | Frontend recebeu tokens centralizados, novo `AppShell`, topbar/sidebar modulares, desacoplamento do `ActivityDiaryProvider` e atualizacao de i18n/testes do shell. |

## Como atualizar ao concluir uma task

1. Trocar o status da task no quadro acima para `DONE`.
2. Adicionar uma linha em `Historico`.
3. Manter o resumo em no maximo duas frases.
4. Registrar apenas o que mudou de fato em produto, backend e frontend.

## Regra operacional

Nenhuma task deve ser considerada concluida sem:

- ajuste de i18n quando houver texto novo
- testes/validacao minima do escopo alterado
- atualizacao desta story principal
