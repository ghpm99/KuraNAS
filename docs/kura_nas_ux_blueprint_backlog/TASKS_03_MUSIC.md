# Tasks 03 - Music

## TASK-007 - Reorganizar IA e rotas do dominio Musica
- Status: `TODO`
- Tamanho: `M`
- Objetivo: alinhar musica ao mapa `Inicio`, `Playlists`, `Artistas`, `Albuns`, `Generos` e `Pastas`.
- Estado atual: `MusicSidebar.tsx` ja possui essas views, mas tudo acontece dentro de uma unica pagina sem rotas do dominio.
- Escopo:
- Criar subrotas ou estado de navegacao robusto para os contextos de musica.
- Padronizar cabecalhos e navegacao secundaria do dominio.
- Preparar o terreno para a Home de musica.
- Criterios de aceite:
- Usuario entende em que contexto de musica esta.
- Cada subarea e navegavel por URL ou estado restauravel.
- Dependencias: `TASK-002`.

## TASK-008 - Entregar Home de musica e contexto de reproducao
- Status: `TODO`
- Tamanho: `M`
- Objetivo: sair da listagem crua e criar uma experiencia de consumo proxima do blueprint.
- Estado atual: ha biblioteca, fila global e playlists, mas sem uma tela inicial do dominio.
- Escopo:
- Criar uma Home de musica com continuar ouvindo, playlists em destaque, artistas e albuns recentes.
- Integrar melhor a fila global e o contexto de origem da reproducao.
- Melhorar cards e secoes para consumo rapido.
- Criterios de aceite:
- A entrada em `/music` nao cai mais em uma lista generica.
- O player global preserva melhor o contexto de origem.
- Dependencias: `TASK-007`.

## TASK-009 - Normalizar metadados e playlists automaticas de musica
- Status: `TODO`
- Tamanho: `M`
- Objetivo: sustentar musica com dados tratados no backend, e nao apenas metadados crus.
- Estado atual: backend possui playlists manuais e player state; generos e agrupamentos ainda dependem de valores crus.
- Escopo:
- Criar camada de normalizacao para artista, album e genero.
- Introduzir playlists automaticas de primeiro nivel para musica.
- Expor contratos que suportem `continue ouvindo`, `recentes`, `favoritas` e agrupamentos confiaveis.
- Criterios de aceite:
- Generos deixam de explodir alias crus no frontend.
- O backend consegue responder playlists automaticas previsiveis.
- Dependencias: `TASK-008`.
