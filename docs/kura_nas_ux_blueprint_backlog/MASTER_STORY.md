# MASTER STORY - KuraNAS UX Blueprint

## Objetivo

Acompanhar a evolucao do KuraNAS do estado atual para o blueprint de UX, mantendo:

- backlog visivel
- status por task
- historico curto do que foi concluido
- proxima entrega recomendada

## Snapshot atual

- Status geral: `IN_PROGRESS`
- Fase atual recomendada: `Sistema`
- Proxima task recomendada: `TASK-016`

## Quadro de tasks

| ID | Titulo | Status | Fase |
| --- | --- | --- | --- |
| TASK-001 | Migrar IA global e rotas base | DONE | Fundacao |
| TASK-002 | Criar tokens visuais e AppShell unificado | DONE | Fundacao |
| TASK-003 | Entregar Home hub v1 | DONE | Fundacao |
| TASK-004 | Refatorar a estrutura da pagina Arquivos | DONE | Explorer |
| TASK-005 | Tornar a abertura de midia transversal e consistente | DONE | Explorer |
| TASK-006 | Criar Favoritos v1 como dominio proprio | DONE | Explorer |
| TASK-007 | Reorganizar IA e rotas do dominio Musica | DONE | Musica |
| TASK-008 | Entregar Home de musica e contexto de reproducao | DONE | Musica |
| TASK-009 | Normalizar metadados e playlists automaticas de musica | DONE | Musica |
| TASK-010 | Reorganizar IA e rotas do dominio Videos | DONE | Videos |
| TASK-011 | Refinar classificacao e paginas de detalhe de videos | DONE | Videos |
| TASK-012 | Evoluir o player de video para contexto completo | DONE | Videos |
| TASK-013 | Criar pipeline de classificacao persistida para imagens | DONE | Imagens |
| TASK-014 | Reorganizar IA da biblioteca de imagens | DONE | Imagens |
| TASK-015 | Evoluir viewer e acoes de imagem | DONE | Imagens |
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
| 2026-03-14 | TASK-003 | Home v1 passou a combinar continuidade de musica e video, recentes e estado do NAS. | Frontend ganhou busca local no hub, cards responsivos com dados de `analytics`, `video` e `music`, novos textos i18n compartilhados e cobertura de testes para o fluxo inicial. |
| 2026-03-14 | TASK-004 | Arquivos ganhou um explorer mais explicito com breadcrumb, toggle de grade/lista e preview lateral sem poluir o grid. | Frontend recebeu `FilesExplorerScreen`, `FileContent` com modos `grid/list`, acesso movel a arvore e atualizacao de i18n/testes do dominio. |
| 2026-03-14 | TASK-005 | Abertura de midia passou a seguir uma regra unica entre Home, Arquivos e Favoritos. | Frontend ganhou `useMediaOpener`, redirecionamento de audio/video/imagem para experiencias dedicadas e suporte a abrir imagem alvo no viewer via rota. |
| 2026-03-14 | TASK-006 | Favoritos virou pagina propria com filtros por tipo e contexto dedicado. | Frontend ganhou `FavoritesScreen`, filtros locais `Tudo/Pastas/Arquivos/Midias`, reaproveito vertical do `starred`, novos textos i18n e cobertura para o fluxo da nova area. |
| 2026-03-14 | TASK-007 | Musica passou a expor subrotas restauraveis com shell secundario e landing de contexto. | Frontend ganhou IA por URL em `/music`, `/music/playlists`, `/music/artists`, `/music/albums`, `/music/genres` e `/music/folders`, header contextual, i18n compartilhado e testes do dominio. |
| 2026-03-14 | TASK-008 | `/music` passou a abrir uma home de consumo com contexto claro de reproducao. | Frontend ganhou Home de musica com continuar ouvindo, playlists em destaque, artistas/albuns recentes, contexto de origem no player/fila e cobertura para os fluxos novos. |
| 2026-03-14 | TASK-009 | Musica passou a consumir catalogo normalizado do backend com playlists automaticas de primeiro nivel. | Backend ganhou contratos proprios em `/music/library` e playlists automaticas para continuar ouvindo/recentes/favoritas; frontend migrou Home, artistas, albuns, generos e playlists para os contratos novos com i18n e testes atualizados. |
| 2026-03-14 | TASK-010 | Videos passou a navegar por Home, Continuar, Series, Filmes, Pessoais, Clipes e Pastas com shell contextual. | Frontend ganhou subrotas/restauracao por URL em `/videos/*`, header/sidebar dedicados, home por categorias reaproveitando playlists e catalogo, ajustes de i18n compartilhado e cobertura de testes para o novo fluxo. |
| 2026-03-14 | TASK-011 | Videos ganhou detalhe contextual por URL com foco em series, filmes, pessoais e clipes, mais progresso por item. | Backend passou a devolver progresso por item e heuristicas mais robustas para serie/pessoal/clipe; frontend migrou o detalhe para rotas dedicadas, agrupou episodios por temporada e manteve a gestao de pastas sem quebrar o fluxo existente. |
| 2026-03-14 | TASK-012 | O player de video passou a manter contexto visivel, fila contextual e retorno consistente para a biblioteca. | Frontend ganhou tela contextual com origem, proximos itens/relacionados, sincronizacao da URL com a sessao de playback, conclusao correta do ultimo item, i18n compartilhado e cobertura de testes do fluxo. |
| 2026-03-14 | TASK-013 | Imagens passou a receber classificacao persistida no pipeline, com categoria semantica e score de confianca. | Backend passou a classificar `capture/photo/other` no fluxo de metadados e expor isso no contrato de imagens; frontend removeu heuristica local para `Capturas` e `Câmera/Fotos`, mantendo o filtro consumindo a classificacao persistida. |
| 2026-03-14 | TASK-014 | Imagens passou a navegar por Biblioteca, Recentes, Capturas, Fotos, Pastas e Albuns automáticos com shell contextual. | Frontend ganhou subrotas restauraveis em `/images/*`, header/sidebar do dominio, visoes de pastas e albuns automaticos com i18n compartilhado e cobertura para a nova IA da biblioteca. |
| 2026-03-14 | TASK-015 | O viewer de imagens virou uma experiencia principal de consumo com acoes e contexto reais. | Frontend ganhou novo viewer com slideshow, filmstrip opcional, painel de metadados por secoes, favoritar com mutation otimista e handoff para `/files` abrindo a pasta de origem por `path`, com i18n e testes atualizados. |

## Como atualizar ao concluir uma task

1. Trocar o status da task no quadro acima para `DONE`.
2. Adicionar uma linha em `Historico`.
3. Manter o resumo em no maximo duas frases.
4. Registrar apenas o que mudou de fato em produto, backend e frontend.

## Regra operacional

Nenhuma task deve ser considerada concluida sem:

- ajuste de i18n quando houver texto novo
- testes/validacao de 100% do escopo alterado
- não estar de acordo com o clean code e as boas praticas no desenvolvimento de software
- atualizacao desta story principal
