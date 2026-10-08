# 03 — Imagens

Escopo: `backend/internal/api/v1/image/`, queries `image/` e `search_images.sql`, pipeline de metadata/miniatura/classificação no worker, `frontend/src/pages/images`, `components/imageContent/**`, `imageProvider`, app Android (`feature/images`).

## Diagnóstico (2026-10-08)

- **"Só 200 e não carrega mais"**: busca, abas (Capturas/Fotos/Recentes), pastas e álbuns filtram no cliente só o que já foi carregado; o gatilho de próxima página é desligado quando o filtro zera (`useImageContent.ts:441-446`). Não existe busca de imagens no servidor (a global devolve 6–12).
- Extensão gravada sem normalizar (`IMG_0001.JPG`) e consulta case-sensitive (`format = ANY(['.jpg',…])`) — fotos de câmera/Windows podem estar fora da galeria. HEIC/AVIF/TIFF/RAW não são imagem para o sistema.
- `LEFT JOIN image_metadata` com Scan em tipos não anuláveis: uma imagem sem metadata derruba a página com 500.
- Data EXIF (`AAAA:MM:DD HH:MM:SS`) não é parseável no JS: a timeline agrupa pelo mtime.
- IA só classifica categoria/nome sugerido; nada é pesquisável; sem tags/legenda/OCR. Roda síncrona dentro do passo de metadata.
- Miniatura PNG 960×720 com letterbox, sem orientação EXIF, gerada sob demanda; o worker pré-gera 320 que a galeria não usa.
- Sem álbuns reais, seleção em lote, favoritas na galeria, salto por data, total de fotos; viewer navega só no carregado e dá a volta.
- Android pede `limit=50` (ignorado; usa 15) e nunca pagina.

Decisões tomadas na revisão: marcação por conteúdo usa a cadeia de IA já configurada (default Ollama local, como o e-mail), em passo assíncrono de baixa prioridade, respeitando o toggle existente. Sem chamadas externas automáticas: GPS aparece como coordenadas com link para o OpenStreetMap acionado pelo usuário (sem tiles embutidos nem geocodificação). Rotação só visual no viewer. Fora do escopo: link de compartilhamento (sem autenticação por decisão registrada) e reconhecimento facial.

## Lacunas

- [x] 01. Extensão em maiúsculas: normalizar `format` na gravação + migração dos dados + consultas tolerantes
- [x] 02. Página quebra com 500 quando a imagem não tem linha de metadata; `rows.Err()`
- [x] 03. Listagem enxuta no servidor com busca/filtros (q, categoria, favoritas, formato, período, câmera, ordenação) e scroll infinito real com estado de erro
- [x] 04. `taken_at` indexado (EXIF → mtime), timeline agrupada por ele, total de fotos e salto por ano/mês
- [x] 05. Android: paginação da galeria (`page_size`)
- [x] 06. Categorias da IA mapeadas (10 valores, i18n, Fotos inclui retrato/paisagem, Capturas inclui screenshot de app)
- [x] 07. Favoritas: aba na galeria e estrela no card
- [x] 08. Pastas a partir do servidor (contagem e capa reais)
- [ ] 09. Miniaturas: JPEG sem letterbox, orientação EXIF, resampler rápido, cache por largura×altura, tamanho da grade pré-gerado, preview para o viewer
- [ ] 10. Formatos: WebP/BMP/TIFF/HEIC/HEIF/AVIF/JFIF/RAW reconhecidos, com miniatura/preview
- [ ] 11. Seleção múltipla e ações em lote na galeria
- [ ] 12. Viewer: navegar além do carregado (vizinhos), pan/pinça/swipe, baixar original, girar, local no disco, GPS, atalhos ignoram campos de texto
- [ ] 13. Álbuns reais (criar, renomear, apagar, adicionar/remover, capa)
- [ ] 14. Busca por conteúdo: tags, legenda e OCR pela IA em passo assíncrono, pesquisáveis
- [ ] 15. Indexação: IA fora do caminho crítico do metadata
- [ ] 16. Facetas de câmera e formato como filtros
- [x] 17. Desempenho do cliente: sem recomputar pastas/álbuns sobre tudo, refetch no foco limitado
- [ ] 18. Duplicadas acessíveis a partir da galeria
- [x] 19. Código morto (`ImageCategoryTabs`, chaves `IMAGES_CATEGORY_*`, `fileId`/`createdAt` desalinhados)
