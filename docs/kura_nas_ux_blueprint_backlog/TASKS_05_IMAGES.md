# Tasks 05 - Images

## TASK-013 - Criar pipeline de classificacao persistida para imagens
- Status: `DONE`
- Tamanho: `M`
- Objetivo: mover a inteligencia de imagens do frontend para o backend/pipeline.
- Estado atual: `imageContent.tsx` classifica `recent`, `screenshots` e `camera` por heuristica local; nao ha categoria persistida.
- Escopo:
- Definir categorias iniciais de imagem e score de confianca.
- Persistir classificacao no backend junto do pipeline de metadados.
- Expor contratos para filtros confiaveis no frontend.
- Criterios de aceite:
- Frontend deixa de depender de heuristicas locais para categorias principais.
- O backend consegue devolver imagens por categoria consistente.
- Dependencias: nenhuma obrigatoria, mas ajuda ja estar no AppShell final.

## TASK-014 - Reorganizar IA da biblioteca de imagens
- Status: `TODO`
- Tamanho: `M`
- Objetivo: alinhar imagens ao mapa `Biblioteca`, `Recentes`, `Capturas`, `Fotos`, `Pastas` e `Albuns`.
- Estado atual: existe uma boa grade com viewer, mas a IA atual ainda usa filtros tecnicos como `portrait` e `landscape`.
- Escopo:
- Criar navegacao contextual do dominio.
- Priorizar agrupamento por data e categorias com semantica de produto.
- Introduzir base para albuns automaticos.
- Criterios de aceite:
- `/images` deixa de parecer apenas uma grade filtravel.
- O dominio passa a refletir linguagem do blueprint.
- Dependencias: `TASK-013`.

## TASK-015 - Evoluir viewer e acoes de imagem
- Status: `TODO`
- Tamanho: `S`
- Objetivo: transformar o viewer atual em um viewer de produto, com contexto, metadados e acoes relevantes.
- Estado atual: o modal ja exibe imagem, navegacao e detalhes; ainda falta consolidar a experiencia final.
- Escopo:
- Melhorar painel de metadados, navegacao entre itens e acoes rapidas.
- Preparar suporte para abrir pasta, favoritar e slideshow quando o backend suportar.
- Ajustar comportamento para telas largas e mobile.
- Criterios de aceite:
- Viewer funciona como uma experiencia principal de consumo.
- O usuario nao precisa sair do viewer para acao basica de navegacao.
- Dependencias: `TASK-014`.
