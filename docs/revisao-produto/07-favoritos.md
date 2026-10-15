# 07 — Favoritos

Escopo: coluna `home_file.starred` e tudo que a lê/escreve — `backend/internal/api/v1/files/` (toggle, listagem, tree `category=starred`, busca), jobs do worker que regravam `home_file`, filtros de imagens/música, `frontend/src/components/favorites/**`, toggles em arquivos/imagens/música, Home, app Android (`feature/files`).

## Diagnóstico (2026-10-15)

- Fonte única: `home_file.starred` (migration 0002), compartilhada por arquivos, imagens, música e busca. Sem `starred_at`.
- Escrita só por `POST /files/starred/:id` (toggle): lê a linha, inverte e regrava as 14 colunas via `update_file.sql` — não idempotente nem atômica; jobs do worker (scan, checksum, mark-deleted) regravam a linha inteira e podem desfazer uma estrela feita no meio. Id inexistente vira 500; item da lixeira pode ser favoritado.
- Lote = N toggles sequenciais no cliente.
- `/favorites` paginada e indexada, mas a navegação por pasta, breadcrumb e preview estão mortos (`extractFilePath` só reconhece `/files`); clicar sai de Favoritos; sem seleção/lote, ordenação, filtro de tipo no servidor; contagens por filtro parciais.
- Toggle de arquivos sem otimismo nem erro; Home não invalida `['home','favorites']`; overrides de estrela da música nunca são limpos.
- Sem estrela em detalhes/preview de arquivo, vídeo, e busca global (só indicador).
- Android: só a estrela na lista do explorer, sem atualizar a tela; sem tela de Favoritos.
- Mover/renomear/lixeira/restaurar preservam id e estrela (ok).

Decisões tomadas na revisão: contrato de escrita cresce com `PUT /files/starred/:id {"starred": bool}` idempotente e lote `POST /files/starred {"ids", "starred"}`, mantendo o toggle. Favoritar item inexistente ou na lixeira responde 404. Favoritos ordenáveis por nome e por data de favoritado (`starred_at`). Navegação por pasta dentro de Favoritos restaurada (`/favorites/<path>`, com breadcrumb e preview). Estrela global (sem favorito por dispositivo — não há autenticação).

## Lacunas (uma alteração cada)

### Backend
- [x] 01. Escrita da estrela atômica e idempotente (`update_file_starred.sql`, sem regravar a linha inteira)
- [x] 02. Jobs do worker (scan, checksum, mark-deleted) não regravam `starred`
- [x] 03. 404 para id inexistente ou na lixeira no toggle
- [ ] 04. `PUT /files/starred/:id` idempotente devolvendo o estado final
- [ ] 05. `POST /files/starred` em lote
- [ ] 06. Coluna `starred_at` preenchida ao favoritar
- [ ] 07. `GET /files/starred` com filtro de tipo e ordenação (nome, data de favoritado)
- [ ] 08. `GET /files/starred/counts` (pastas, arquivos, mídias)
- [ ] 09. `category` inválida em `/files/tree` responde 400 em vez de cair em "all"
- [ ] 10. Teste de integração do ciclo estrela → mover → lixeira → restaurar → reindexar

### Frontend
- [ ] 11. Navegação por pasta, breadcrumb e preview em `/favorites/<path>`
- [ ] 12. Toggle de arquivos otimista com rollback e erro visível
- [ ] 13. Invalidação das queries de favoritos (Home inclusive) após marcar
- [ ] 14. Seleção, ações em lote e atalhos na tela de Favoritos (usando o lote do servidor)
- [ ] 15. Ordenação e filtro de tipo pelo servidor na tela de Favoritos, contagens reais
- [ ] 16. Estrela nos detalhes/preview do arquivo
- [ ] 17. Estrela no player e no catálogo de vídeos
- [ ] 18. Marcar/desmarcar pelos resultados da busca global
- [ ] 19. Overrides de estrela da música reconciliados com o servidor
- [ ] 20. Contagem "N+ itens" via chave i18n; comentários removidos de `HomeScreen.tsx`

### Android
- [ ] 21. Estrela no explorer atualiza a tela (otimista com rollback, erro do servidor)
- [ ] 22. Tela de Favoritos paginada com navegação
- [ ] 23. Estrela em imagens, música e vídeo

### Decisões do dono
- 🚫 24. `prioritize_favorites` (campo só no Android, sem backend; `PriorityFavoritesStrategy` não lê estrela): remover a flag e renomear a estratégia, ou implementar peso de favoritos nas playlists de vídeo
- 🚫 25. Playlist automática "Favoritas" para vídeo, como na música: criar ou não
- 🚫 26. Arquivo movido fora do app perde a estrela (nova linha no scan): aceitar, ou herdar por checksum/nome ao reindexar
