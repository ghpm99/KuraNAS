# Current State Gap Analysis

## Resumo executivo

O produto ja possui partes importantes do blueprint, mas elas estao distribuidas de forma inconsistente:

- A navegacao global ainda trata o explorer como pagina inicial e mantem `activity-diary` como area de primeiro nivel.
- O shell global existe, mas ainda e parcial: sidebar + header + area de conteudo, sem Home, sem subnavegacao contextual consistente e sem busca global real.
- O dominio de videos e o mais maduro no backend, com playlists inteligentes, classificacao e continuidade de reproducao.
- O dominio de imagens esta mais adiantado no frontend do que no backend, com heuristicas locais e viewer, mas sem classificacao persistida.
- O dominio de musica possui biblioteca, views por artista/albuns/generos/pastas e playlists manuais, mas ainda sem home propria nem normalizacao forte.
- Analytics e About ja existem, mas Settings nao existe e o produto ainda nao foi consolidado nas tres macro-areas do blueprint: explorar, consumir e administrar.

## Mapa de gaps

| Area | Objetivo do blueprint | Estado atual | Gap principal | Observacao pragmatica |
| --- | --- | --- | --- | --- |
| Navegacao global | `Home`, `Arquivos`, `Favoritos`, `Imagens`, `Musicas`, `Videos`, `Analytics`, `Configuracoes`, `Sobre` | `frontend/src/app/App.tsx` usa `/` para `FilePage`, expone `/starred` e `/activity-diary`; nao ha `/home`, `/favorites` ou `/settings` | Arquitetura de informacao ainda centrada em explorer e legado | A primeira entrega deve ser rotas novas + redirects temporarios |
| Sidebar e topbar | Um AppShell unico com subnavegacao contextual | `Sidebar.tsx` e `Header.tsx` duplicam navegacao, incluem `activity-diary` e nao refletem o sitemap final | Shell existe, mas nao representa o novo produto | Vale reaproveitar estrutura, nao reescrever do zero |
| Home | Hub inteligente com busca, recentes, continuar ouvindo/assistindo e estado do NAS | Nao existe pagina Home | Falta a pagina mais importante do blueprint | Ja existem dados suficientes para uma Home v1 combinando arquivos, videos, analytics e fila de musica |
| Arquivos | Explorer puro com breadcrumb, tree, grid/lista e preview lateral | `fileProvider`, `FileContent` e `FileViewer` ja entregam arvore, tabs, grid e preview inline | UX ainda esta acoplada a selecao interna do explorer e nao a um shell mais claro | A base tecnica e boa para evolucao incremental |
| Abertura inteligente | Arquivos devem abrir no melhor viewer/player conforme tipo | `fileViewer.tsx` ja suporta imagem, audio, video e documento | Comportamento ainda preso ao preview do explorer; nao e transversal | Falta reaproveitar a mesma regra em favoritos, busca e Home |
| Favoritos | Agregador transversal de arquivos, pastas e midias | Hoje `starred` e so uma variante de `FilePage` | Favoritos nao e um dominio proprio | Pode nascer em v1 com o que ja existe no campo `starred` |
| Musica | Home propria, playlists automaticas, normalizacao de artista/album/genero, player global contextual | Frontend tem views por `all/artists/albums/genres/folders/playlists`; backend tem playlists manuais e player state | Falta produto de musica, nao so listagens | A fila global de musica e uma base forte para evolucao |
| Videos | Biblioteca tipo streaming, rotas por contexto, continuar assistindo robusto e player contextual | Backend ja tem playlists inteligentes, playback state e classificacao; frontend mostra playlists + biblioteca geral | IA de videos ainda nao acompanha a riqueza do backend | Este dominio deve ser usado como alavanca para a Home e para o shell final |
| Imagens | Biblioteca organizada por biblioteca/recentes/capturas/fotos/pastas/albuns, com viewer rico | Frontend tem tabs heuristicas e viewer modal; backend so entrega arquivos de imagem e metadados | Classificacao e agrupamento ainda vivem no frontend, sem persistencia | Precisa mover a inteligencia para o pipeline/backend |
| Analytics | Separar visao geral de biblioteca/indexacao | Existe apenas `/analytics` com overview unico | Falta segmentacao e falta integrar mais sinais do pipeline | O backend ja entrega parte do necessario para v1 |
| Configuracoes | Pagina de configuracao para biblioteca, indexacao, players, aparencia e idioma | Nao existe pagina nem endpoints alem de `translation` e `about` | Falta um dominio inteiro do sitemap final | Deve nascer junto com novos contratos em `configuration` |
| Sobre | Pagina tecnica compacta | Ja existe com versao, hash, runtime e path monitorado | Falta adequar ao novo shell e reduzir sobreposicao com configuracoes | E uma refatoracao leve em comparacao com Settings |
| Busca global | Command palette para arquivos, pastas, artistas, albuns, playlists, videos, imagens e acoes | `Header.tsx` tem apenas um input visual sem comportamento global | Falta um dos fluxos centrais da Home e do shell | Pode entrar depois da consolidacao das rotas e dominios |
| Diario de atividades | Sair da navegacao principal e virar historico tecnico interno ou ser removido | Existe frontend, provider, layout dedicado e endpoints backend | Legado ainda esta embutido no shell | Precisa sair do caminho critico sem apagar capacidade tecnica cedo demais |

## Maturidade por dominio

### 1. Fundacao visual e shell
- Ja existe dark theme em `frontend/src/components/providers/appProviders.tsx`.
- O layout base em `frontend/src/components/layout/Layout/Layout.tsx` resolve estrutura minima.
- Ainda faltam tokens centralizados, background global do blueprint, AppShell unico e desacoplamento de `ActivityDiaryProvider`.

### 2. Explorer
- O explorer atual ja tem provider, arvore, tabs e preview.
- O problema principal nao e ausencia tecnica; e enquadramento de produto.
- A task deve focar em UX estrutural, nao em reimplementar CRUD de arquivos.

### 3. Musica
- Existe um modulo funcional de biblioteca com playlists manuais.
- Falta camada de produto: Home de musica, playlists automaticas, normalizacao de genero e contexto de reproducao mais rico.
- O backend ainda nao modela playlists automaticas como entidade principal nesse dominio.

### 4. Videos
- E o dominio mais alinhado ao blueprint no backend.
- Ja ha classificacao, playback state, navegacao de playlist e catalogo inicial.
- O principal gap esta na navegacao e no refinamento da experiencia frontend.

### 5. Imagens
- O frontend ja entrega uma experiencia usavel de consumo.
- O backend ainda nao sustenta categorias persistidas como `capturas`, `fotos` e `albuns automaticos`.
- O risco aqui e crescer heuristica duplicada no frontend sem consolidar pipeline.

### 6. Sistema
- Analytics e About ja fornecem base de administracao.
- Settings e busca global ainda nao existem.
- `activity-diary` consome espaco mental e estrutural que o blueprint quer remover da camada principal.

## Conclusao

O caminho mais seguro nao e uma reescrita total. O backlog foi quebrado para:

1. Migrar a arquitetura de informacao primeiro.
2. Consolidar um AppShell unico.
3. Reaproveitar o que ja esta maduro em arquivos, musica e especialmente videos.
4. Subir a inteligencia de classificacao de imagens e musica para o backend.
5. Fechar o produto com Settings, busca global e remocao de legado visual.
