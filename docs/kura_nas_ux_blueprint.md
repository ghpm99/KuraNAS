# KuraNAS — UX Blueprint Completo

## Objetivo
Redesenhar o KuraNAS para consolidar três mundos que hoje estão misturados no produto — gerenciador de arquivos, biblioteca de mídia e dashboard do sistema — em uma experiência coerente, previsível, escalável e com identidade visual única.

---

# 1. Arquitetura completa de UX

## 1.1 Problema atual
Hoje o produto mistura no mesmo nível de navegação:
- exploração de arquivos (modelo Dropbox/Explorer)
- consumo de mídia (modelo Spotify/Netflix/Google Photos)
- administração do sistema (modelo dashboard técnico)

Isso quebra o modelo mental do usuário. A consequência é que o sistema parece vários produtos unidos sem uma lógica única.

## 1.2 Princípio central da nova UX
Separar claramente o produto em três domínios:
1. **Explorar** → arquivos e pastas
2. **Consumir** → imagens, músicas e vídeos
3. **Administrar** → analytics, configurações e informações do sistema

## 1.3 Novo mapa de navegação

### Sidebar principal
- Home
- Arquivos
- Favoritos
- Imagens
- Músicas
- Vídeos
- Analytics
- Configurações
- Sobre

### Regras
- **Home** vira a página inicial padrão.
- **Arquivos** concentra a navegação por filesystem.
- **Favoritos** vira agregador transversal.
- **Imagens / Músicas / Vídeos** são bibliotecas de mídia independentes.
- **Analytics** e **Sobre** ficam em Sistema.
- **Diário de Atividades** sai da navegação principal e deve ser removido do produto, salvo se reaproveitado como histórico técnico interno.

## 1.4 Novo modelo de navegação por camadas

### Camada 1: Navegação global
Sidebar com grandes áreas do produto.

### Camada 2: Navegação contextual
Cada área possui sua própria subnavegação.

Exemplos:
- Arquivos → Tudo / Recentes / Compartilhados no futuro / Favoritos
- Imagens → Biblioteca / Pessoas no futuro / Álbuns / Pastas / Capturas
- Músicas → Início / Playlists / Artistas / Álbuns / Gêneros / Pastas
- Vídeos → Início / Continuar assistindo / Séries / Filmes / Pessoais / Pastas

### Camada 3: Estado de consumo
Player global, viewer global e overlays de ação rápida.

## 1.5 Página inicial ideal: Home
A Home não deve ser um explorer. Ela deve ser um hub inteligente.

### Seções da Home
- Busca global
- Continuar ouvindo
- Continuar assistindo
- Imagens recentes
- Arquivos recentes
- Favoritos recentes
- Biblioteca de mídia
- Estado do NAS

### Objetivo da Home
Dar ao usuário três respostas imediatas:
- o que eu estava fazendo
- onde estão minhas mídias principais
- como está meu sistema

## 1.6 Jornada principal por tipo de usuário

### Usuário explorador de arquivos
Home → Arquivos → Pasta → Arquivo → Preview/Viewer

### Usuário de música
Home → Músicas → Playlist/Álbum/Artista → Player global

### Usuário de vídeo
Home → Vídeos → Continuar assistindo / Série / Filme / Pasta → Player

### Usuário de fotos
Home → Imagens → Biblioteca / Capturas / Pastas → Viewer

### Usuário técnico
Home → Analytics → detalhes de storage/processamento/indexação

## 1.7 Regras de consistência
- Toda página usa a mesma estrutura de layout.
- O usuário nunca perde a noção de onde está.
- Mídia sempre abre no player/viewer adequado.
- Arquivos, favoritos e busca respeitam o tipo do arquivo e usam a melhor experiência disponível.

---

# 2. Design system completo

## 2.1 Direção visual
Visual escuro, premium, técnico e moderno, com sensação de produto local sofisticado.
Referências conceituais:
- Plex / Jellyfin para mídia
- Spotify para música
- Linear / Vercel para consistência visual
- Dropbox para clareza em arquivos

## 2.2 Background global
Usar como background padrão de toda a aplicação:

```css
background:
  radial-gradient(1200px 400px at 20% -10%, rgba(225, 29, 72, 0.18), transparent 60%),
  radial-gradient(900px 320px at 95% 2%, rgba(14, 116, 144, 0.22), transparent 62%),
  linear-gradient(180deg, #070a10 0%, #0a1018 100%);
```

## 2.3 Paleta de cores

### Base
- Background Root: `#070A10`
- Background Elevated: `#0A1018`
- Surface 1: `#0F1623`
- Surface 2: `#121A29`
- Surface 3: `#182234`
- Border Subtle: `#1B2637`
- Border Strong: `#263349`

### Texto
- Text Primary: `#F3F7FF`
- Text Secondary: `#B8C2D9`
- Text Muted: `#7E8AA3`
- Text Disabled: `#5B6579`

### Marca / ação
- Primary: `#6D5DF6`
- Primary Hover: `#7C70FF`
- Cyan Accent: `#06B6D4`
- Pink Accent: `#E11D48`
- Success: `#22C55E`
- Warning: `#F59E0B`
- Danger: `#EF4444`

### Overlays
- Overlay Soft: `rgba(7,10,16,0.62)`
- Overlay Strong: `rgba(7,10,16,0.82)`

## 2.4 Gradientes utilitários
- Hero Gradient: pink → transparent / cyan → transparent
- Music Gradient: roxo + azul
- Video Gradient: magenta + cyan escuro
- Image Gradient: neutro escuro com leve brilho

## 2.5 Grid e espaçamento
Escala base de 4px.
- 4 / 8 / 12 / 16 / 20 / 24 / 32 / 40 / 48 / 64

Container principal:
- desktop wide: 1440+ px
- content max width opcional para dashboards: 1600 px

## 2.6 Bordas e cantos
- radius sm: 10px
- radius md: 14px
- radius lg: 18px
- radius xl: 24px
- radius pill: 999px

## 2.7 Sombras
- card: `0 8px 30px rgba(0,0,0,.24)`
- floating: `0 12px 40px rgba(0,0,0,.35)`
- active glow primary: `0 0 0 1px rgba(109,93,246,.6), 0 8px 30px rgba(109,93,246,.18)`

## 2.8 Tipografia
Fonte principal: Inter.

Escala:
- Display 32/40 semibold
- H1 28/36 semibold
- H2 24/32 semibold
- H3 20/28 semibold
- Title 18/26 semibold
- Body 14/22 regular
- Body Small 13/20 regular
- Caption 12/18 medium

## 2.9 Ícones
Lucide ou equivalente linear.
Regras:
- traço consistente
- tamanhos padronizados 16 / 18 / 20 / 24
- evitar misturar famílias de ícone

## 2.10 Componentes base

### Estruturais
- AppShell
- Sidebar
- Topbar
- ContentHeader
- SectionHeader
- PageContainer
- SplitLayout
- EmptyState
- ErrorState

### Inputs
- SearchInput
- FilterChip
- Select
- SegmentedControl
- Toggle
- Checkbox
- Radio
- Slider
- TextField
- TextArea

### Feedback
- Toast
- Badge
- ProgressBar
- Skeleton
- Loader
- StatusPill

### Cards
- FileCard
- FolderCard
- MediaCard
- AlbumCard
- PlaylistCard
- MetricCard
- StatCard
- RecentItemCard

### Navegação
- NavItem
- Breadcrumb
- Tabs
- SecondarySidebar
- Pagination / LoadMore

### Modais e overlays
- Dialog
- Drawer
- CommandPalette
- ContextMenu
- DropdownMenu
- LightboxViewer
- QuickPreviewPanel

### Player / viewer
- GlobalPlayer
- MiniPlayer
- VideoPlayerShell
- ImageViewerShell
- QueuePanel

## 2.11 Estados de componentes
Todo componente deve ter:
- default
- hover
- focus visible
- active
- selected
- disabled
- loading
- empty quando aplicável

## 2.12 Motion
Animações discretas.
- hover cards: elevar 2–4px
- transições: 150–220ms
- drawers/modals: ease-out suave
- evitar exagero em áreas de arquivo e grade densa

---

# 3. Layout das telas

## 3.1 Estrutura global padrão

### Shell
- Sidebar fixa à esquerda
- Topbar superior com busca global, ações contextuais e avatar
- Área de conteúdo com header local e corpo scrollável
- Player global fixo inferior

## 3.2 Home

### Blocos
1. Hero de boas-vindas com busca
2. Continuar ouvindo
3. Continuar assistindo
4. Imagens recentes
5. Biblioteca
6. Arquivos recentes
7. Estado do sistema

### Comportamento
- cards horizontais para consumo rápido
- seções com “ver tudo”
- adaptação para conteúdo vazio

## 3.3 Arquivos

### Objetivo
Ser o modo explorer puro do produto.

### Estrutura
- header: breadcrumb + ações (criar pasta, enviar, ordenar)
- painel esquerdo opcional: árvore de diretórios
- painel central: grade ou lista
- preview lateral opcional em telas largas

### Tabs
- Todos
- Recentes
- Favoritos

### Regras de abertura
- imagem → viewer
- vídeo → player
- áudio → player
- documentos simples → preview quando suportado
- pasta → navegação normal

## 3.4 Favoritos
Dois modos:
- Tudo
- Pastas
- Arquivos
- Mídias

Permitir filtro por tipo.
Usar preview real em vez de ícones genéricos quando houver thumbnail.

## 3.5 Imagens

### Problema a resolver
Hoje é uma grade caótica.

### Nova estrutura
Subnav:
- Biblioteca
- Recentes
- Capturas
- Fotos
- Pastas
- Álbuns automáticos

### Biblioteca
Grade densa com agrupamento por data.

### Capturas
Agrupa screenshots, prints e capturas de tela.

### Fotos
Prioriza fotos pessoais e imagens de câmera.

### Pastas
Modo baseado em filesystem.

### Álbuns automáticos
- Viagens
- Documentos visuais
- Wallpapers
- Memes
- Outros

### Viewer
- imagem central
- strip inferior opcional
- metadados laterais
- ações: favoritar, baixar no futuro, abrir pasta, slideshow

## 3.6 Músicas

### Objetivo
Trazer uma experiência próxima de Spotify, mas ancorada no conteúdo local.

### Navegação
- Início
- Playlists
- Artistas
- Álbuns
- Gêneros
- Pastas

### Início
- Continue ouvindo
- Mixes automáticos
- Tocadas recentemente
- Álbuns adicionados recentemente
- Artistas favoritos
- Gêneros principais

### Playlists
- automáticas
- manuais no futuro
- por pasta
- por humor no futuro

### Artistas
- grid com capa gerada
- nome padronizado
- contagem de faixas, álbuns e duração

### Álbuns
- capa principal
- metadados agregados
- ações de play / fila / favorito

### Gêneros
- apenas gêneros normalizados
- nunca exibir a explosão de aliases crus

### Pastas
- visão orientada a filesystem para quem organizou música por diretórios

### Página de detalhe do artista
- hero compacto com nome e stats
- músicas principais
- álbuns
- participações no futuro

### Página de álbum
- capa grande
- lista de faixas
- duração total
- informações técnicas

## 3.7 Vídeos

### Objetivo
Parecer biblioteca de streaming, não lista caótica de arquivos.

### Navegação
- Início
- Continuar assistindo
- Séries
- Filmes
- Pessoais
- Clipes
- Pastas

### Início
- continuar assistindo
- adicionados recentemente
- séries em andamento
- filmes
- vídeos pessoais recentes

### Séries
Agrupar temporadas e episódios automaticamente.

### Filmes
Agrupar filmes únicos ou pastas de filmes.

### Pessoais
Vídeos gravados pelo usuário, celular, câmera, DVR, gravações de tela.

### Clipes
Conteúdo curto, recortes, memes e vídeos avulsos.

### Pastas
Modo orientado a diretório.

### Página de série
- banner / poster gerado
- temporadas
- episódios
- progresso por episódio

### Player de vídeo
- tela principal limpa
- fila/playlist lateral opcional
- próximos episódios sugeridos
- retomada automática

## 3.8 Analytics
Separar em duas áreas:
- Visão geral
- Biblioteca e indexação

### Visão geral
- uso de storage
- distribuição por tipo de mídia
- arquivos maiores
- duplicados no futuro
- saúde do sistema no futuro

### Biblioteca e indexação
- quantos arquivos indexados
- quantas mídias categorizadas
- pendências de thumbnail / metadata
- erros de processamento

## 3.9 Sobre
Página técnica compacta.
- versão
- commit hash
- host info
- caminho raiz monitorado
- links úteis futuros

## 3.10 Configurações
- biblioteca e pastas observadas
- indexação
- players
- aparência
- idioma
- metadados / automação futura

---

# 4. Modelagem das playlists automáticas

## 4.1 Princípio
Playlists automáticas devem ser entidades de primeiro nível do sistema. Não devem existir só no frontend.

## 4.2 Tipos de playlists automáticas

### Música
- Recentemente adicionadas
- Mais tocadas
- Tocadas recentemente
- Favoritas
- Por artista
- Por álbum
- Por gênero normalizado
- Por pasta
- Descobertas recentes
- Faixas sem metadados completos

### Vídeo
- Continuar assistindo
- Séries em andamento
- Por série
- Por temporada
- Por pasta
- Vídeos pessoais recentes
- Clipes curtos
- Filmes adicionados recentemente

### Imagens
Mais próximo de álbuns dinâmicos do que playlists:
- Recentes
- Capturas
- Fotos
- Pastas
- Memes
- Wallpapers
- Documentos visuais

## 4.3 Entidades conceituais

### Playlist
- id
- type (manual, automatic)
- media_type (audio, video, image)
- title
- slug
- description
- source_rule
- artwork
- visibility
- created_at
- updated_at

### PlaylistItem
- id
- playlist_id
- media_id
- position
- derived_score
- derived_reason
- added_at

### PlaybackSession
- id
- media_type
- media_id
- playlist_id opcional
- current_time
- duration
- status
- last_played_at
- device no futuro

## 4.4 Regras de geração

### Música — exemplos
- **Recentemente adicionadas**: order by created_at desc
- **Mais tocadas**: order by play_count desc
- **Continue ouvindo**: status in paused/in_progress
- **Por álbum**: agrupar faixas por normalized_album
- **Por artista**: agrupar por normalized_artist
- **Mix automático**: combinar artista, gênero e recência

### Vídeo — exemplos
- **Continuar assistindo**: sessões com progresso > 3% e < 95%
- **Série em andamento**: episódios com pelo menos um item em progresso
- **Próximo episódio**: detectar ordenação sequencial
- **Pasta de vídeo pessoal**: paths + metadados + duração + padrão de nome

## 4.5 Regras de ordenação
- manual quando usuário editar no futuro
- automática por score
- automática por sequência natural
- automática por recência

## 4.6 Continuidade de reprodução
O sistema deve salvar:
- última mídia
- ponto de reprodução
- fila atual
- contexto de origem (playlist, pasta, álbum, série)

Assim o retorno do usuário é sempre contextual.

---

# 5. Modelo de classificação de mídia

## 5.1 Objetivo
Parar de exibir arquivos crus como biblioteca final. A mídia precisa de enriquecimento e classificação.

## 5.2 Pipeline conceitual
1. Descoberta do arquivo
2. Extração de metadados
3. Normalização
4. Classificação
5. Agrupamento
6. Geração de thumbnails/artwork
7. Publicação na biblioteca

## 5.3 Classificação de imagens

### Sinais usados
- extensão
- pasta
- nome do arquivo
- resolução
- EXIF
- data de captura
- origem da câmera
- proporção

### Categorias iniciais
- foto
- captura de tela
- wallpaper
- meme
- documento visual
- ilustração/anime
- ícone/asset
- outras

### Heurísticas práticas
- nomes contendo screenshot, captura, print → captura
- imagens pequenas com fundo transparente e proporções de asset → ícone/asset
- EXIF de câmera/celular → foto
- pastas DCIM/Camera → foto
- pastas downloads/memes → meme potencial

## 5.4 Classificação de áudio

### Sinais usados
- ID3 tags
- artista/álbum/gênero
- duração
- bitrate
- pasta
- nome do arquivo

### Etapas
1. limpar valores vazios
2. normalizar variações de artista/álbum/gênero
3. detectar tracks soltas vs álbuns
4. detectar conteúdo não musical no futuro (podcast, aula, gravação)

### Gêneros normalizados
Criar uma camada de alias.
Exemplos:
- Alt. Rock → Rock
- Alternative Rock → Rock
- pop rock → Rock / Pop-Rock conforme política
- christian rock → Rock cristão ou Rock se simplificar

## 5.5 Classificação de vídeo

### Sinais usados
- duração
- resolução
- pasta
- nome do arquivo
- padrões S01E01, 1080p, WEBRip, BluRay etc
- origem do arquivo

### Categorias iniciais
- episódio de série
- filme
- vídeo pessoal
- gravação de tela
- clipe curto
- anime
- conteúdo avulso

### Heurísticas práticas
- SxxExx → série
- nome com episode/ep → série
- duração longa única sem episódio → filme potencial
- pastas de DVR / câmera / celular → vídeo pessoal
- duração muito curta → clipe

## 5.6 Níveis de confiança
Cada classificação deve registrar confidence score.
- high
- medium
- low

Itens com baixa confiança podem cair em “Outros”.

## 5.7 Correção futura pelo usuário
No futuro, permitir corrigir categoria, artista, álbum, gênero ou tipo de vídeo. Isso retroalimenta o sistema.

---

# 6. Estrutura de componentes React

## 6.1 Princípios
- layout e estado desacoplados
- providers para estado de domínio
- hooks para lógica simples e consumo derivado
- componentes pequenos e reutilizáveis
- páginas compostas por blocos previsíveis

## 6.2 Estrutura sugerida

```txt
src/
  app/
    home/
    files/
    images/
    music/
    videos/
    favorites/
    analytics/
    settings/
    about/

  components/
    app-shell/
    navigation/
    feedback/
    cards/
    media/
    file-browser/
    player/
    viewer/
    analytics/
    forms/
    common/

  features/
    home/
    files/
    images/
    music/
    videos/
    favorites/
    analytics/
    search/
    player/
    library/

  providers/
    AppShellProvider/
    SearchProvider/
    PlayerProvider/
    FilesProvider/
    ImagesProvider/
    MusicProvider/
    VideosProvider/
    FavoritesProvider/
    AnalyticsProvider/

  hooks/
    useDebounce.ts
    useKeyboardShortcuts.ts
    useMediaActions.ts
    useFormattedDuration.ts
    useNormalizedSearch.ts
    useBreakpoint.ts

  lib/
    utils/
    constants/
    mappers/
    formatters/
    classifiers/

  types/
    media.ts
    file.ts
    player.ts
    analytics.ts
```

## 6.3 Shell da aplicação

### AppShell
Responsável por:
- sidebar global
- topbar
- área scrollável
- player global
- overlays globais

### Exemplo conceitual
- `AppShell`
  - `Sidebar`
  - `Topbar`
  - `MainContent`
  - `GlobalPlayer`
  - `CommandPalette`

## 6.4 Providers principais

### PlayerProvider
Gerencia:
- fila atual
- item atual
- tipo de mídia
- play/pause/seek
- origem do contexto
- miniplayer/fullscreen

### FilesProvider
Gerencia:
- diretório atual
- árvore
- grid/list view
- ordenação
- seleção
- preview lateral

### MusicProvider
Gerencia:
- home da música
- playlists
- artistas
- álbuns
- gêneros
- contexto de reprodução

### VideosProvider
Gerencia:
- continuar assistindo
- séries
- filmes
- vídeos pessoais
- player context

### ImagesProvider
Gerencia:
- biblioteca
- filtros
- agrupamento
- viewer

## 6.5 Hooks úteis
- `usePlayer()`
- `useFilesExplorer()`
- `useImageLibrary()`
- `useMusicLibrary()`
- `useVideoLibrary()`
- `useGlobalSearch()`
- `useSelection()`
- `useMediaThumbnail()`
- `useOpenMedia()`

## 6.6 Componentes por domínio

### File browser
- `FileExplorerHeader`
- `FileTreeSidebar`
- `FileGrid`
- `FileList`
- `FileCard`
- `FolderCard`
- `FilePreviewPanel`
- `FileContextMenu`

### Music
- `MusicHero`
- `MusicSectionRow`
- `AlbumCard`
- `ArtistCard`
- `GenreCard`
- `TrackRow`
- `PlaylistCard`
- `QueueDrawer`

### Video
- `VideoHero`
- `ContinueWatchingRow`
- `SeriesCard`
- `MovieCard`
- `EpisodeRow`
- `SeasonSelector`
- `VideoPlayerLayout`

### Images
- `ImageMasonryGrid` ou `ImageGrid`
- `PhotoCard`
- `CaptureCard`
- `ImageViewer`
- `MetadataPanel`
- `DateGroupHeader`

### Shared media
- `MediaCard`
- `MediaRow`
- `Thumbnail`
- `Artwork`
- `ProgressBadge`
- `FavoriteButton`

## 6.7 Padrão de composição das páginas

### Exemplo: página de música
- `MusicPage`
  - `MusicLayout`
    - `MusicSidebar`
    - `MusicContent`
      - `MusicHero`
      - `MusicSectionRow`
      - `MusicSectionRow`
      - `TrackTable`

### Exemplo: página de arquivos
- `FilesPage`
  - `FilesLayout`
    - `FileTreeSidebar`
    - `FilesContent`
      - `FileExplorerHeader`
      - `FileGrid`
      - `QuickPreviewPanel`

## 6.8 Busca global
Criar uma `CommandPalette` com:
- arquivos
- pastas
- artistas
- álbuns
- playlists
- vídeos
- imagens
- ações rápidas

## 6.9 Tokens de design no frontend
Idealmente criar tokens centralizados:
- color tokens
- spacing tokens
- radius tokens
- shadow tokens
- z-index scale
- motion durations

## 6.10 Convenções de naming
- componentes visuais: PascalCase
- hooks: `useX`
- providers: `XProvider`
- DTOs e modelos em `types/`
- mapeamentos e normalizações em `lib/mappers`

---

# 7. Regras de comportamento e produto

## 7.1 Todos os arquivos não pode ser mídia burra
A visão de arquivos continua sendo explorer, mas abrir um item sempre respeita a melhor experiência de consumo.

## 7.2 Favoritos precisa ser transversal
Favoritos devem aceitar:
- arquivos
- pastas
- álbuns
- playlists
- séries no futuro

## 7.3 Biblioteca não pode refletir metadado cru sem tratamento
Toda biblioteca precisa usar dados normalizados.

## 7.4 O sistema precisa parecer um produto único
Mesmos tokens, mesmos cards, mesma estrutura, mesma linguagem visual.

---

# 8. Roadmap de implementação recomendado

## Fase 1 — Fundação
- definir arquitetura de navegação
- criar AppShell único
- padronizar background, cores, tipografia, cards e topbar
- mover Home para página inicial

## Fase 2 — Explorer e consumo integrado
- ajustar Arquivos para abrir viewers/players corretos
- padronizar Favoritos
- implementar preview consistente

## Fase 3 — Música
- normalização de artista/álbum/gênero
- home de música
- playlists automáticas
- cards de artista, álbum e gênero

## Fase 4 — Vídeos
- continuar assistindo robusto
- classificação por série, filme, pessoal e clipe
- páginas de detalhe e player contextual

## Fase 5 — Imagens
- classificação foto/captura/documento/meme/asset
- biblioteca organizada
- viewer e metadados

## Fase 6 — Analytics e refinamento
- consolidar visão técnica
- remover legado visual restante
- adicionar busca global

---

# 9. Resultado esperado
Ao final, o KuraNAS deixa de ser “um explorer com telas paralelas” e passa a ser:
- um **hub local de arquivos**
- uma **biblioteca moderna de mídia**
- um **painel técnico coerente do NAS**

O produto ganha clareza, valor percebido e identidade própria.


---

# 10. Sitemap final do produto

```
/
  home

/files
  /files
  /files/recent
  /files/favorites

/images
  /images
  /images/recent
  /images/captures
  /images/photos
  /images/folders
  /images/albums

/music
  /music
  /music/playlists
  /music/artists
  /music/albums
  /music/genres
  /music/folders

/videos
  /videos
  /videos/continue
  /videos/series
  /videos/movies
  /videos/personal
  /videos/clips
  /videos/folders

/favorites

/analytics

/settings

/about
```

---

# 11. Wireframes conceituais das telas

Os wireframes abaixo descrevem estrutura e hierarquia visual das telas.

---

# 11.1 Home

```
+-------------------------------------------------------------+
| Sidebar |                Topbar (Search)                    |
|         +---------------------------------------------------+
|         | Hero / Busca Global                               |
|         | "Buscar arquivos, músicas, vídeos..."             |
|         +---------------------------------------------------+
|         | Continuar ouvindo                                 |
|         | [album][album][album][album]                      |
|         +---------------------------------------------------+
|         | Continuar assistindo                              |
|         | [serie][filme][video pessoal]                     |
|         +---------------------------------------------------+
|         | Imagens recentes                                  |
|         | [img][img][img][img][img]                          |
|         +---------------------------------------------------+
|         | Arquivos recentes                                 |
|         | [arquivo][arquivo][arquivo]                       |
|         +---------------------------------------------------+
|         | Estado do NAS                                     |
|         | Storage | CPU | RAM | Indexação                   |
+-------------------------------------------------------------+
```

Objetivo: dar visão imediata do sistema e do consumo recente.

---

# 11.2 Arquivos

```
+-------------------------------------------------------------+
| Sidebar | Topbar | Breadcrumb / Ações                       |
|         +---------------------------------------------------+
|         | TreeSidebar | FileGrid                            |
|         |             |                                     |
|         | Pastas      | [folder] [folder] [file]             |
|         |             | [file]   [file]   [file]              |
|         |             |                                     |
|         |             |                                     |
+-------------------------------------------------------------+
```

Opcional em telas largas:

```
| Tree | Grid | Preview Panel |
```

Preview abre:
- imagem
- vídeo
- áudio
- metadados

---

# 11.3 Imagens

```
+-------------------------------------------------------------+
| Sidebar | Topbar                                            |
|         +---------------------------------------------------+
|         | Tabs: Biblioteca | Recentes | Capturas | Pastas   |
|         +---------------------------------------------------+
|         | Agosto 2025                                     |
|         | [img][img][img][img][img][img]                    |
|         |                                                 |
|         | Julho 2025                                      |
|         | [img][img][img][img]                              |
+-------------------------------------------------------------+
```

Viewer:

```
+-------------------------------------------------------------+
| imagem fullscreen                                           |
|                                                             |
| <  previous | next  >                                       |
|                                                             |
| metadata panel lateral opcional                             |
+-------------------------------------------------------------+
```

---

# 11.4 Música

Página inicial da música:

```
+-------------------------------------------------------------+
| Sidebar | Topbar                                            |
|         +---------------------------------------------------+
|         | Hero: Continue ouvindo                            |
|         | [album][album][album]                              |
|         +---------------------------------------------------+
|         | Playlists automáticas                              |
|         | [playlist][playlist][playlist]                     |
|         +---------------------------------------------------+
|         | Artistas populares                                 |
|         | [artist][artist][artist]                           |
|         +---------------------------------------------------+
|         | Álbuns recentes                                    |
|         | [album][album][album]                              |
+-------------------------------------------------------------+
```

Página de artista:

```
+-------------------------------------------------------------+
| Artist Hero                                                 |
| nome + stats                                                |
+-------------------------------------------------------------+
| Top Tracks                                                  |
| Track list                                                  |
+-------------------------------------------------------------+
| Albums                                                      |
| [album][album][album]                                       |
+-------------------------------------------------------------+
```

---

# 11.5 Vídeos

```
+-------------------------------------------------------------+
| Sidebar | Topbar                                            |
|         +---------------------------------------------------+
|         | Continuar assistindo                              |
|         | [video][video][video]                              |
|         +---------------------------------------------------+
|         | Séries                                            |
|         | [serie][serie][serie]                              |
|         +---------------------------------------------------+
|         | Filmes                                            |
|         | [filme][filme][filme]                              |
|         +---------------------------------------------------+
|         | Vídeos pessoais                                   |
|         | [video][video][video]                              |
+-------------------------------------------------------------+
```

Player:

```
+-------------------------------------------------------------+
| Video Player                                                |
|                                                             |
| controls                                                    |
|                                                             |
| Next episode / related videos                               |
+-------------------------------------------------------------+
```

---

# 11.6 Analytics

```
+-------------------------------------------------------------+
| Storage usage                                               |
| [pie chart]                                                 |
+-------------------------------------------------------------+
| Media distribution                                          |
| images | videos | audio                                     |
+-------------------------------------------------------------+
| Largest files                                               |
| table                                                       |
+-------------------------------------------------------------+
| Library processing                                          |
| indexed | pending | errors                                  |
+-------------------------------------------------------------+
```

---

# 12. Backlog técnico de implementação

## Fase 1 — Fundação visual

Implementar:

- AppShell unificado
- Sidebar nova
- Topbar padrão
- Background global
- Tokens de design
- Componentes base

Componentes:

- AppShell
- Sidebar
- Topbar
- PageContainer
- SectionHeader
- Card base

---

## Fase 2 — Explorer

Implementar:

- FileGrid
- FileTreeSidebar
- FilePreviewPanel
- FileContextMenu
- abertura inteligente de mídia

---

## Fase 3 — Player global

Implementar:

- PlayerProvider
- GlobalPlayer
- Queue system
- Media session

---

## Fase 4 — Música

Implementar:

- MusicProvider
- ArtistCard
- AlbumCard
- GenreCard
- PlaylistCard
- TrackRow

Lógica backend:

- normalização de gênero
- playlists automáticas

---

## Fase 5 — Vídeos

Implementar:

- VideoProvider
- ContinueWatching
- Series detection
- Episode grouping

---

## Fase 6 — Imagens

Implementar:

- ImageProvider
- classificação automática
- viewer

---

## Fase 7 — Analytics

Implementar:

- métricas de biblioteca
- gráficos

---

# 13. Roadmap estimado

Se implementado por um dev principal:

- Fundação visual: 1–2 semanas
- Explorer refinado: 1 semana
- Player global: 1 semana
- Música: 2–3 semanas
- Vídeos: 2 semanas
- Imagens: 1–2 semanas
- Analytics: 1 semana

Total estimado: **9–12 semanas de evolução do produto**.

---

# 14. Próxima evolução possível

Futuras melhorias do KuraNAS:

- reconhecimento facial em fotos
- detecção de duplicados
- recomendações inteligentes
- sincronização com dispositivos
- mobile companion
- streaming remoto opcional

Essas evoluções transformam o produto em um **media OS pessoal completo**.

