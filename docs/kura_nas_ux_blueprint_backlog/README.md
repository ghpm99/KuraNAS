# KuraNAS UX Blueprint Backlog

Este diretorio transforma o blueprint de UX em um backlog incremental, orientado pelo estado real do sistema hoje.

## Fontes analisadas
- `docs/kura_nas_ux_blueprint.md`
- `frontend/src/app/App.tsx`
- `frontend/src/components/layout/Layout/Layout.tsx`
- `frontend/src/components/layout/Layout/index.tsx`
- `frontend/src/components/layout/Sidebar/Sidebar.tsx`
- `frontend/src/components/layout/Header/Header.tsx`
- `frontend/src/components/fileContent/fileContent.tsx`
- `frontend/src/components/fileContent/components/fileViewer/fileViewer.tsx`
- `frontend/src/components/musicContent/musicContent.tsx`
- `frontend/src/components/music/MusicSidebar.tsx`
- `frontend/src/components/imageContent/imageContent.tsx`
- `frontend/src/components/providers/videoContentProvider/videoContentProvider.tsx`
- `frontend/src/components/videos/videoContent/components/VideoContentScreen.tsx`
- `backend/internal/app/routes.go`
- `backend/internal/api/v1/video/service.go`
- `backend/internal/api/v1/video/playlist/classifier.go`
- `backend/internal/api/v1/music/service.go`

## Estrutura
- `CURRENT_STATE_GAP_ANALYSIS.md`: diagnostico entre blueprint e produto atual.
- `MASTER_STORY.md`: story principal para TODO, andamento e historico curto.
- `TASKS_01_FOUNDATION.md`: navegacao, shell e home.
- `TASKS_02_EXPLORER_AND_FAVORITES.md`: arquivos, preview e favoritos.
- `TASKS_03_MUSIC.md`: arquitetura e evolucao de musica.
- `TASKS_04_VIDEOS.md`: arquitetura e evolucao de videos.
- `TASKS_05_IMAGES.md`: pipeline e experiencia de imagens.
- `TASKS_06_SYSTEM_AND_SEARCH.md`: analytics, settings, about, busca e fechamento.

## Regras de uso
- Sempre atualizar `MASTER_STORY.md` ao concluir uma task.
- Cada conclusao deve registrar uma linha curta em `Historico`.
- Toda task de produto deve validar i18n, testes e impacto de rotas.
- Quando uma task depender de backend e frontend, manter entrega vertical, sem deixar contrato quebrado no meio.

## Legenda
- `TODO`: ainda nao iniciada.
- `IN_PROGRESS`: em execucao.
- `BLOCKED`: dependente de decisao externa ou pre-requisito.
- `DONE`: concluida e registrada na story principal.

## Ordem recomendada
1. Fundacao
2. Explorer e favoritos
3. Musica
4. Videos
5. Imagens
6. Sistema e busca
