# 06 — Vídeos

Escopo: `backend/internal/api/v1/video/` (+ `playlist/`), queries `video/`, metadata/thumbnail/playlist no worker, `frontend/src/features/videos/**`, `service/videoPlayback.ts`, app Android (`feature/video`).

## Diagnóstico (2026-10-09)

- `.avi/.mkv/.wmv/.flv` estão em `VideoFormats` mas não são vídeo para o pipeline (sem metadata, thumbnail, playlist); stream cru sem remux/transcode; player fica em tela preta sem mensagem.
- Estado de reprodução é uma linha por IP: trocar de vídeo apaga a posição anterior; vídeo concluído reabre no fim; "Continuar" com no máximo 1 item.
- Catálogo só pela lista "Todos os vídeos" (botão manual); pastas com 1 vídeo/nome genérico e filmes soltos inacessíveis; "Filmes" com teto de 20; rebuild completo de playlists por arquivo novo; regex de episódio pega "1920x1080"; ordem lexicográfica de episódios.
- Home do catálogo faz chamada de IA síncrona (10 s) por request para um campo que ninguém lê.
- Player sem legendas, atalhos, PiP, persistência de volume/velocidade; tela cheia esconde os controles.
- Nomes de playlists e seções fixos em português no backend; Android sem playlists, progresso e paginação.

Decisões tomadas na revisão: `.ts` fica fora (conflita com TypeScript). Remux e transcode sob demanda via ffmpeg (como na música), sem HLS. "Concluído" a partir de 90%. Classificação de filme/série persistida por heurística (caminho + duração). Pôster só de imagem local (`poster.jpg`/`folder.jpg`), sem serviço externo. Motor de comportamento mantido e alimentado pelo progresso por vídeo. Fora do escopo: HLS adaptativo e seleção de faixa de áudio (dependem de pipeline de streaming dedicado).

## Lacunas (uma alteração cada)

### Conteúdo que não toca ou não aparece
- [x] 01. `.avi/.mkv/.wmv/.flv` reconhecidos como vídeo no pipeline
- [x] 02. Reconciliação de vídeos existentes sem `video_metadata`
- [x] 03. Extensões ausentes (`.m4v .mpg .mpeg .m2ts .3gp .vob`)
- [x] 04. MIME explícito para formatos de vídeo
- [x] 05. Erro de reprodução visível no player com ação "baixar original"
- [x] 06. Remux sob demanda (mkv/avi com H.264 → MP4 fragmentado)
- [x] 07. Transcode sob demanda de codecs não suportados (HEVC etc.) → H.264/AAC

### Retomada, continuar assistindo e assistido
- [x] 08. Estado do player por dispositivo (client id), não por IP
- [x] 09. Posição por vídeo (`video_watch_progress`)
- [x] 09.1. Próximo/anterior retomam a posição salva do vídeo de destino
- [x] 10. Rever vídeo concluído recomeça do início
- [x] 11. "Continuar assistindo" com vários itens
- [x] 12. Marcar assistido/não assistido manualmente
- [x] 13. Status por item da playlist lido do progresso por vídeo
- [x] 14. Progresso salvo ao fechar a aba
- [x] 15. "Concluído" a partir de 90%

### Navegação e catálogo
- [x] 16. Remover chamada de IA do `/video/catalog/home`
- [x] 17. Tela de vídeos não bloqueia esperando 3 queries
- [x] 18. Estados de erro com tentar de novo nas listas
- [x] 19. Navegação por hierarquia de pastas
- [x] 20. Classificação de vídeo persistida
- [x] 21. Seção "Filmes" paginada e completa
- [x] 22. Rebuild de playlists coalescido (sem rebuild completo por arquivo)
- [x] 23. Rebuild remove playlists órfãs/vazias
- [x] 24. Rebuild resiliente à unicidade de `order_index` com itens manuais
- [x] 25. Reordenar playlist sem violar a unicidade
- [x] 26. Ordenação natural dos episódios
- [x] 27. Regex de episódio com fronteira de palavra (backend)
- [x] 28. Regex de episódio com fronteira de palavra (frontend)
- [x] 29. Palavras-chave de classificação por segmento de caminho
- [x] 29.1. `coursePatternSpec` e `clipSpec` ("curso", "reel", "clip") também por token
- [x] 30. Detalhe da playlist paginado
- [x] 31. Lista de playlists paginada por seção
- [x] 32. Playlists de um vídeo (`by-video/:id`) no lugar de memberships completas
- [x] 33. Scroll infinito em "Todos os vídeos"
- [x] 34. Ordenação da biblioteca
- [x] 35. Filtros da biblioteca (formato, resolução, duração)
- [x] 36. Busca com fold de acento e índice trigram
- [x] 37. Duração/resolução/codec na listagem (contrato)
- [x] 38. Duração/resolução nos cards
- [x] 39. `?q=` sincronizado na URL
- [x] 40. "Adicionar à playlist" com menu e busca, sem padrão implícito
- [x] 41. Criar e excluir playlist personalizada (`create_playlist.sql` grava `is_auto = TRUE` mesmo para custom — corrigir junto)
- [x] 42. Ocultar/mostrar playlist na UI
- [x] 43. Painel de detalhes do vídeo com baixar original
- [x] 44. Pôster local de série/filme

### Thumbnails
- [x] 45. Frame do thumbnail em % da duração
- [x] 46. Fallback de thumbnail não fica em cache permanente
- [x] 47. Prévia GIF carregada só no hover
- [x] 48. Worker não gera GIF de prévia para todo vídeo
- [x] 48.1. Thumbnails derivados de um frame mestre por vídeo (hoje cada tamanho pedido — 240…1280 — roda um ffmpeg; o worker pré-gera só 320×180)
- [x] 49. Log/registro de acesso uma vez por reprodução, não por Range

### Player
- [x] 50. Atalhos de teclado
- [x] 51. Tela cheia no contêiner (controles visíveis)
- [x] 52. Picture-in-Picture
- [x] 52.1. Media Session API (título e controles do SO no PiP/notificação)
- [x] 53. Volume, mudo e velocidade persistidos
- [x] 53.1. Mudo independente do volume (hoje mudo = volume 0; ao recarregar mudo, o volume anterior se perde)
- [x] 54. Seek aplicado ao soltar o controle
- [x] 55. Legendas externas `.srt/.vtt`
- [x] 55.1. Desativar legenda também é lembrado (hoje o idioma lembrado religa a legenda no vídeo seguinte)
- [x] 56. Legendas embutidas
- [x] 57. Menu de configurações morto removido ou ligado ao player real
- [ ] 58. Autoplay do próximo com contagem regressiva e cancelar
- [ ] 59. Stream com 416, ETag e If-Range
- [ ] 59.1. `GetVideoMetadataByID` com SELECT de 22 colunas e Scan de 21 (consulta por id de metadata quebrada)

### i18n e limpeza
- [ ] 60. Nomes de playlists e seções do backend traduzidos (slug + i18n)
- [ ] 61. Badge de classificação via i18n
- [ ] 62. Código órfão do frontend de vídeos
- [ ] 63. Motor de comportamento alimentado pelo progresso por vídeo

### Android
- [ ] 64. Playlists visíveis no Android
- [ ] 65. "Recentes" sem duplicar itens
- [ ] 66. Reportar e retomar progresso
- [ ] 67. Biblioteca paginada com busca
- [ ] 68. Próximo/anterior e continuar assistindo
