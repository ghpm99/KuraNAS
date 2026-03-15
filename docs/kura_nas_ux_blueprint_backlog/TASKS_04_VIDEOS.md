# Tasks 04 - Videos

## TASK-010 - Reorganizar IA e rotas do dominio Videos
- Status: `DONE`
- Tamanho: `M`
- Objetivo: alinhar videos ao mapa `Inicio`, `Continuar assistindo`, `Series`, `Filmes`, `Pessoais`, `Clipes` e `Pastas`.
- Estado atual: o frontend ja mostra playlists e biblioteca geral, mas ainda sem a segmentacao do blueprint.
- Escopo:
- Criar navegacao contextual do dominio.
- Reaproveitar o catalogo e as playlists inteligentes ja existentes.
- Definir o comportamento de `/videos` como Home do dominio.
- Criterios de aceite:
- O usuario nao precisa mais interpretar playlists cruas para navegar na biblioteca.
- O dominio passa a ter rotas/estados claros por categoria.
- Dependencias: `TASK-002`.

## TASK-011 - Refinar classificacao e paginas de detalhe de videos
- Status: `DONE`
- Tamanho: `M`
- Objetivo: aprofundar series, filmes, pessoais e clipes em paginas de detalhe coerentes.
- Estado atual: backend ja possui classificacao e playlists inteligentes; frontend ainda opera mais por playlist do que por experiencia de biblioteca.
- Escopo:
- Ajustar heuristicas e taxonomias de classificacao onde houver ruido.
- Criar detalhe de serie, agrupamento de episodios e paginas por contexto relevante.
- Aproveitar `continue watching` e `home catalog`.
- Criterios de aceite:
- Series e filmes deixam de parecer apenas listas renomeadas.
- O usuario encontra detalhes, progresso e continuidade por contexto.
- Dependencias: `TASK-010`.

## TASK-012 - Evoluir o player de video para contexto completo
- Status: `TODO`
- Tamanho: `M`
- Objetivo: transformar o player em extensao natural da biblioteca, com proximo item, retomada e contexto visivel.
- Estado atual: `videoPlayer.tsx` toca o video e possui controles; o fluxo ainda e muito isolado da biblioteca.
- Escopo:
- Exibir contexto de origem, proximos episodios ou relacionados e retorno consistente.
- Melhorar retomada automatica e navegacao lateral quando fizer sentido.
- Integrar melhor com a navegacao do dominio.
- Criterios de aceite:
- O usuario assiste e volta ao contexto sem se perder.
- O player oferece continuidade real, nao so reproducao.
- Dependencias: `TASK-011`.
