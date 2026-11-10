# 08 — Lixeira

Escopo: `backend/internal/api/v1/trash/` (service, handler, purger), tabela `trash_item` (migration 0036), exclusão em `files/operations.go`, `restore_subtree.sql`, `frontend/src/components/trash/**`, `frontend/src/service/trash.ts`, diálogo de exclusão, app Android (`feature/files`).

## Diagnóstico (2026-11-10)

- Excluir move os bytes para `.kuranas-trash/` na raiz dona do caminho e faz soft-delete da subárvore em `home_file`; `?permanent=true` apaga direto. Rotas: `GET /trash`, `POST /trash/:id/restore`, `DELETE /trash/:id`, `DELETE /trash`, `GET|PUT /trash/retention`. Purgador no boot e a cada 12 h, retenção padrão de 30 dias.
- Isolamento ok: scan, watcher, watch folders, WebDAV e backup ignoram a lixeira; listagens, busca e mídia filtram `deleted_at`; erros via i18n; id e estrela preservados ao restaurar.
- Restaurar arquivo do tier frio só zera `deleted_at`: `physical_path` continua apontando para o caminho frio de onde os bytes já saíram (leitura de código, a confirmar em teste).
- Purgar pasta com descendentes frios deixa as cópias frias órfãs; linhas soft-deleted de `home_file` nunca são removidas depois do purge.
- Filho e pai excluídos separadamente: restaurar o filho recria o pai vazio e o pai passa a dar 409 sem saída.
- Conflito 409 não informa o caminho ocupado; bytes sumidos da lixeira viram 500 genérico; "esvaziar" não informa falhas.
- `trash_item` não guarda tipo; DTO só tem `id, original_path, size, deleted_at`; tamanho de pasta é suspeito (entrada do FS, não o conteúdo); listagem sem total, ordenação ou busca; sem lote.
- Operações da lixeira são registradas pelo `logger` nos handlers; o purgador automático só escreve no log de console. Chaves `ANALYTICS_TRASH_*` existem sem endpoint nem uso.
- Frontend: tela mínima — sem nome/ícone, "expira em", seleção/lote, ordenação, busca, total; `window.confirm`; `['trash']` não é invalidada ao excluir arquivos; página vazia após remover o último item; fallback de erro de mutação usa a mensagem de carregamento; campo de retenção não ressincroniza nem acusa valor inválido; `isMutating` trava a lista inteira; data no locale do navegador.
- Android: sem lixeira; `FilesViewModel.deleteFile` existe mas nenhuma tela chama — o app não exclui.
- Comentários dentro de funções em `trash/` e `useTrashScreen.ts`.

## Lacunas (uma alteração cada)

### Backend
- [x] 01. Restaurar arquivo do tier frio deixa a linha consistente (`physical_path` limpo), com teste Postgres do ciclo excluir → restaurar
- [x] 02. Purgar pasta remove as cópias frias dos descendentes
- [x] 03. Purge (item, esvaziar, expirado) remove as linhas soft-deleted da subárvore em `home_file`
- [x] 04. Tamanho de pasta na lixeira é a soma real do conteúdo (confirmar antes o que `home_file.size` guarda para diretórios)
- [x] 05. Restaurar com bytes ausentes responde erro específico em vez de 500 genérico; item continua removível
- [x] 06. 409 de restauração informa o caminho ocupado (campo `path`)
- [x] 07. `trash_item` guarda o tipo; `GET /trash` devolve `name`, `is_directory` e `expires_at`
- [x] 08. `GET /trash/summary` (itens, bytes)
- [x] 09. Ordenação e busca em `GET /trash` (parâmetro inválido → 400)
- [x] 10. Restaurar e excluir em lote com resultado por id
- [x] 11. "Esvaziar" devolve `{purged, failed}`
- [x] 12. Descartada: o "diário" (`diary`) é o diário pessoal de atividades do usuário, não um registro de operações; excluir/restaurar/purgar/esvaziar já são registrados pelo `logger` nos handlers. Notificar o expurgo automático virou decisão do dono (36)
- [x] 13. Teste de integração Postgres do ciclo completo (arquivo, pasta, reindexação)
- [x] 14. Comentários removidos do pacote `trash/`

### Frontend
- [x] 15. `['trash']` invalidada após excluir arquivos (explorer e imagens)
- [x] 16. Página atual vazia após restaurar/excluir volta para a anterior
- [x] 17. Erro de restaurar/excluir com mensagem própria (`TRASH_ACTION_ERROR`)
- [x] 18. Diálogo de confirmação do app no lugar de `window.confirm`
- [x] 19. Campo de retenção ressincroniza, acusa valor inválido e sobrevive a falha do `GET`
- [x] 20. Item com nome, ícone de pasta/arquivo, origem e "expira em N dias" (depende da 07)
- [x] 21. Resumo no topo: total de itens e bytes (depende da 08)
- [x] 22. Ordenação e busca pelo servidor (depende da 09)
- [ ] 23. Seleção e ações em lote (depende da 10)
- [ ] 24. Botões travam por item, não a lista inteira
- [ ] 25. Data formatada no locale do app
- [ ] 26. Após restaurar, ação "ver na pasta"
- [ ] 27. Render sem mock e seam test para cada mutação da lixeira (conferir o que já existe)
- [x] 28. Comentários removidos de `useTrashScreen.ts`

### Android
- [ ] 29. Excluir pelo explorer com confirmação, aviso "vai para a Lixeira" e erro do servidor
- [ ] 30. Tela de Lixeira: listar paginado, restaurar, excluir definitivo, esvaziar

### Decisões do dono
- 🚫 31. Retenção: permitir "nunca expirar" e/ou teto de espaço (quota) da lixeira
- 🚫 32. Conflito ao restaurar: manter 409, restaurar com sufixo, escolher outro destino ou sobrescrever
- 🚫 33. Pai ausente ou também na lixeira: recriar pastas faltantes (hoje), exigir o pai antes, ou mesclar pastas ao restaurar
- 🚫 34. Espaço da lixeira no Analytics (chaves `ANALYTICS_TRASH_*` órfãs): criar `/analytics/trash` ou remover as chaves
- 🚫 35. Toast "Desfazer" após excluir (exige a exclusão devolver o id do item da lixeira)
- 🚫 36. Notificação in-app quando o expurgo automático remove itens (ou falha): criar ou não
