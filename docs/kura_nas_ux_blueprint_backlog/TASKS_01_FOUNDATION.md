# Tasks 01 - Foundation

## TASK-001 - Migrar IA global e rotas base
- Status: `DONE`
- Tamanho: `M`
- Objetivo: introduzir `Home`, `Arquivos`, `Favoritos` e `Configuracoes` como rotas de primeiro nivel, removendo `activity-diary` da navegacao principal.
- Estado atual: `frontend/src/app/App.tsx` ainda usa `/` como explorer; `Sidebar.tsx` e `Header.tsx` ainda exibem `activity-diary` e nao possuem `home`, `favorites` e `settings`.
- Escopo:
- Criar rotas `/home`, `/files`, `/favorites`, `/settings`.
- Manter redirects temporarios de `/` e `/starred`.
- Atualizar `ActivePageListener` e o tipo `pages` do provider de UI.
- Criterios de aceite:
- O produto abre em `Home` e nao mais em explorer.
- A navegacao principal reflete o sitemap macro do blueprint.
- `activity-diary` deixa de ser item principal sem quebrar acesso tecnico temporario.
- Dependencias: nenhuma.

## TASK-002 - Criar tokens visuais e AppShell unificado
- Status: `DONE`
- Tamanho: `M`
- Objetivo: consolidar background global, paleta, espacos, raio, sombras e estrutura unica de shell para todas as paginas.
- Estado atual: existe dark theme em `appProviders.tsx` e um layout base em `Layout.tsx`, mas sem tokens centralizados nem a direcao visual do blueprint.
- Escopo:
- Extrair tokens visuais centrais.
- Atualizar tema/base CSS para o background global e superficies do blueprint.
- Refatorar `Layout` para virar `AppShell` reutilizavel com sidebar, topbar e area scrollavel padrao.
- Criterios de aceite:
- Todas as paginas principais usam o mesmo shell estrutural.
- O visual base deixa de depender de overrides dispersos.
- O shell nao fica acoplado a `ActivityDiaryProvider`.
- Dependencias: `TASK-001`.

## TASK-003 - Entregar Home hub v1
- Status: `TODO`
- Tamanho: `M`
- Objetivo: criar a pagina inicial como hub de continuidade, biblioteca e estado do sistema.
- Estado atual: nao existe `Home`; o usuario cai direto em `FilePage`.
- Escopo:
- Criar pagina `Home` com hero/busca, recentes, continuar ouvindo, continuar assistindo e estado do NAS.
- Reaproveitar dados existentes de videos, analytics, arquivos recentes e fila global de musica sempre que possivel.
- Tratar estados vazios sem parecer erro.
- Criterios de aceite:
- `Home` responde "o que eu estava fazendo", "onde estao minhas midias" e "como esta o NAS".
- A pagina funciona em desktop e mobile.
- Nao introduz texto hardcoded fora de i18n.
- Dependencias: `TASK-001`, `TASK-002`.
