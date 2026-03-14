# Tasks 02 - Explorer And Favorites

## TASK-004 - Refatorar a estrutura da pagina Arquivos
- Status: `DONE`
- Tamanho: `M`
- Objetivo: alinhar `Arquivos` ao papel de explorer puro, com header claro, breadcrumb, tree opcional, grid/lista e preview lateral.
- Estado atual: `FilePage` ja possui action bar, tabs, grid e sidebar de detalhes, mas a estrutura ainda e muito presa ao layout legado.
- Escopo:
- Reorganizar `FilePage` para um shell de explorer mais explicito.
- Destacar breadcrumb, acoes, modos de visualizacao e preview.
- Preservar CRUD e carregamento existentes.
- Criterios de aceite:
- A pagina deixa claro que esta em "Arquivos" e nao em Home.
- Em telas largas, preview lateral funciona sem poluir o grid.
- A arvore continua acessivel no fluxo de arquivos.
- Dependencias: `TASK-002`.

## TASK-005 - Tornar a abertura de midia transversal e consistente
- Status: `DONE`
- Tamanho: `S`
- Objetivo: garantir que imagem, audio, video e documento sempre abram na melhor experiencia disponivel em qualquer ponto do produto.
- Estado atual: `fileViewer.tsx` ja sabe exibir tipos de arquivo, mas a regra ainda esta concentrada no explorer.
- Escopo:
- Extrair uma regra unica de abertura por tipo.
- Reusar essa regra em Arquivos, Favoritos, Home e futuras buscas.
- Evitar que midia seja tratada como arquivo bruto quando ja houver player/viewer melhor.
- Criterios de aceite:
- Abertura de item e consistente em todos os pontos cobertos.
- Video abre no fluxo de player, imagem no viewer e audio no player global quando aplicavel.
- Dependencias: `TASK-004`.

## TASK-006 - Criar Favoritos v1 como dominio proprio
- Status: `DONE`
- Tamanho: `M`
- Objetivo: transformar favoritos em pagina propria e nao mais em variante escondida do explorer.
- Estado atual: `/starred` reaproveita `FilePage`; o dominio aceita apenas arquivos marcados com `starred`.
- Escopo:
- Criar pagina `/favorites` com filtros iniciais por `Tudo`, `Pastas`, `Arquivos` e `Midias`.
- Reaproveitar o campo `starred` existente como fonte de v1.
- Mostrar thumbnail/preview real quando houver.
- Criterios de aceite:
- Favoritos deixa de ser so um alias de Arquivos.
- Usuario consegue filtrar o conjunto sem perder contexto.
- Dependencias: `TASK-001`, `TASK-005`.
