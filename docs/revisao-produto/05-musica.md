# 05 — Música

Escopo: `backend/internal/api/v1/music/` + queries `music/`, metadata de áudio (`scripts/audio_metadata.py`, worker), streaming (`/files/stream/:id`), `frontend/src/features/music/**`, `service/{music,playlist,playerState}.ts`, app Android (`feature/music`).

## Diagnóstico (2026-10-09)

- Só `.mp3/.wav/.aac/.flac` são áudio (`.m4a/.ogg/.opus/.wma` somem); tags lidas só como ID3 — FLAC/M4A ficam sem artista/álbum/título e somem de Artistas/Álbuns/Gêneros; ano lido de `TYER` (vazio em ID3 v2.4); sem disco, capa, letras.
- Web lê `metadata.duration`, backend envia `length`: duração nunca aparece.
- Chaves com `/` (pastas, "AC/DC") dão 404 no Gin.
- Todo endpoint de catálogo carrega a biblioteca inteira e agrupa em memória por request.
- Fila "lembrada" nunca é gravada; estado do player chaveado por IP; playlists automáticas (ids negativos) não abrem; "tocar tudo" trunca em 200; "carregar mais" manual e inacessível; sem página de álbum/artista, ordenação, favoritos, mais tocadas, transcodificação.
- Android sem paginação (50 itens por lista).

Decisões tomadas na revisão: `.ogg/.oga/.opus` são áudio (`.ogv` vídeo); `.wma`/`.alac` entram como áudio e tocam via transcodificação ffmpeg sob demanda quando o navegador não suporta (sem ffmpeg, servem os bytes crus). Chaves com `/` resolvidas no roteador (raw path), sem mudar contrato. Estado do player por dispositivo (id do cliente gerado no navegador, IP como fallback). Uma reprodução conta após 30 s ou metade da faixa. Coletânea = álbum com artistas divergentes na mesma pasta → "Vários artistas". Aba "Faixas" com todas as músicas.

## Lacunas

- [x] 01. Formatos de áudio (`.m4a/.ogg/.oga/.opus/.wma/.alac`) reconhecidos
- [x] 02. Tags por formato (Vorbis/MP4/ID3 via mutagen easy) com disco, ano (`date`/`TDRC`), letras; reprocessar faixas sem tags
- [x] 03. Duração aparece na web (`length`), tipos alinhados ao contrato real
- [x] 04. Chaves com `/` funcionam (pastas, artistas, álbuns)
- [x] 05. Catálogo agregado no SQL com paginação e índices (sem carregar a biblioteca por request)
- [x] 06. Ordenação de faixas por disco/faixa numéricos; opções de ordenação nas listas
- [x] 07. Playlists automáticas abrem
- [ ] 08. "Tocar tudo" sem truncar: fila por contexto (todas as faixas, payload leve)
- [ ] 09. Scroll infinito acessível em todas as listas de música + estados de erro/vazio
- [ ] 10. Persistência da fila e do estado (posição, volume, shuffle, repeat) por dispositivo
- [ ] 11. Fila: tocar a seguir, adicionar ao fim, reordenar, salvar como playlist
- [ ] 12. Playlists: renomear, reordenar, confirmar exclusão, aviso de faixa repetida, IA separada
- [ ] 13. Capa de álbum (embutida e `folder.jpg`) em listas, player, fila e Media Session
- [ ] 14. Páginas de álbum e de artista
- [ ] 15. Favoritos dentro da música, histórico e mais tocadas
- [ ] 16. Coletâneas e álbuns homônimos agrupados corretamente; gênero com acento inicial
- [ ] 17. Transcodificação sob demanda para formatos não suportados
- [ ] 18. Player: erro sem laço, seek ao soltar, shuffle por permutação, atalhos de teclado, Media Session completa, registro de acesso uma vez por reprodução
- [ ] 19. Letras no player expandido
- [ ] 20. Aba "Faixas", busca dentro da Música, código morto e i18n (`Unknown Artist`)
- [ ] 21. Android: paginação e faixa por id
