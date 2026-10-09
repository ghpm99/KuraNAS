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
- [ ] 05. Erro de reprodução visível no player com ação "baixar original"
- [ ] 06. Remux sob demanda (mkv/avi com H.264 → MP4 fragmentado)
- [ ] 07. Transcode sob demanda de codecs não suportados (HEVC etc.) → H.264/AAC

### Retomada, continuar assistindo e assistido
- [ ] 08. Estado do player por dispositivo (client id), não por IP
- [ ] 09. Posição por vídeo (`video_watch_progress`)
- [ ] 10. Rever vídeo concluído recomeça do início
- [ ] 11. "Continuar assistindo" com vários itens
- [ ] 12. Marcar assistido/não assistido manualmente
- [ ] 13. Status por item da playlist lido do progresso por vídeo
- [ ] 14. Progresso salvo ao fechar a aba
- [ ] 15. "Concluído" a partir de 90%

### Navegação e catálogo
- [ ] 16. Remover chamada de IA do `/video/catalog/home`
- [ ] 17. Tela de vídeos não bloqueia esperando 3 queries
- [ ] 18. Estados de erro com tentar de novo nas listas
- [ ] 19. Navegação por hierarquia de pastas
- [ ] 20. Classificação de vídeo persistida
- [ ] 21. Seção "Filmes" paginada e completa
- [ ] 22. Rebuild de playlists coalescido (sem rebuild completo por arquivo)
- [ ] 23. Rebuild remove playlists órfãs/vazias
- [ ] 24. Rebuild resiliente à unicidade de `order_index` com itens manuais
- [ ] 25. Reordenar playlist sem violar a unicidade
- [ ] 26. Ordenação natural dos episódios
- [ ] 27. Regex de episódio com fronteira de palavra (backend)
- [ ] 28. Regex de episódio com fronteira de palavra (frontend)
- [ ] 29. Palavras-chave de classificação por segmento de caminho
- [ ] 30. Detalhe da playlist paginado
- [ ] 31. Lista de playlists paginada por seção
- [ ] 32. Playlists de um vídeo (`by-video/:id`) no lugar de memberships completas
- [ ] 33. Scroll infinito em "Todos os vídeos"
- [ ] 34. Ordenação da biblioteca
- [ ] 35. Filtros da biblioteca (formato, resolução, duração)
- [ ] 36. Busca com fold de acento e índice trigram
- [ ] 37. Duração/resolução/codec na listagem (contrato)
- [ ] 38. Duração/resolução nos cards
- [ ] 39. `?q=` sincronizado na URL
- [ ] 40. "Adicionar à playlist" com menu e busca, sem padrão implícito
- [ ] 41. Criar e excluir playlist personalizada
- [ ] 42. Ocultar/mostrar playlist na UI
- [ ] 43. Painel de detalhes do vídeo com baixar original
- [ ] 44. Pôster local de série/filme

### Thumbnails
- [ ] 45. Frame do thumbnail em % da duração
- [ ] 46. Fallback de thumbnail não fica em cache permanente
- [ ] 47. Prévia GIF carregada só no hover
- [ ] 48. Worker não gera GIF de prévia para todo vídeo
- [ ] 49. Log/registro de acesso uma vez por reprodução, não por Range

### Player
- [ ] 50. Atalhos de teclado
- [ ] 51. Tela cheia no contêiner (controles visíveis)
- [ ] 52. Picture-in-Picture
- [ ] 53. Volume, mudo e velocidade persistidos
- [ ] 54. Seek aplicado ao soltar o controle
- [ ] 55. Legendas externas `.srt/.vtt`
- [ ] 56. Legendas embutidas
- [ ] 57. Menu de configurações morto removido ou ligado ao player real
- [ ] 58. Autoplay do próximo com contagem regressiva e cancelar
- [ ] 59. Stream com 416, ETag e If-Range

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
