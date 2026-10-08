# 01 — Arquivos

Escopo: navegador genérico de arquivos (`backend/internal/api/v1/files/`, `frontend/src/features/files/`, `service/files.ts`, `components/actionBar`, `components/folderPicker`) e sua relação com tiering/raízes.

## Diagnóstico (2026-10-08)

- A listagem pagina certo no backend (`hasNext`), mas o frontend só lê `data.pages[0]` — toda pasta mostra no máximo 200 itens.
- Toda operação age sobre o "item aberto" (pasta onde se está ou arquivo aberto): não há seleção, multi-seleção, menu de contexto.
- O caminho físico (`physical_path`) é `json:"-"`; arquivo demovido ao tier frio é inencontrável fora do app.
- Favoritos/Recentes filtram por `parent_path` (favorito em subpasta não aparece em `/favorites`); Recentes usa `LIMIT 10` global.
- Download lê o arquivo inteiro em RAM (`os.ReadFile` + `c.Data`), sem Range nem `Content-Disposition`; pasta e múltiplos não baixam.
- Sem ordenação, sem busca na pasta, upload sem progresso/drag&drop/pasta e abortando o lote no primeiro conflito; copiar arquivo frio falha.

Decisões tomadas na revisão (owner pediu execução contínua): expor o local físico via endpoint próprio (`/files/location/:id`) — coerente com a decisão "sem auth, whitelist de IP"; `/files/blob/:id` e demais rotas mantidas por contrato (Android). Busca por conteúdo fica para a funcionalidade 04 (Busca) / 03 (Imagens). Upload retomável (chunked/tus) fora do escopo — upload passa a ser um arquivo por requisição com progresso.

## Lacunas

- [x] 01. Infinite scroll na listagem (todas as páginas) e no FolderPicker
- [x] 02. Contagem de filhos ignora `deleted_at` + N+1 por diretório + total real no cabeçalho
- [x] 03. Teto de `page_size` e validação de `page` (handler segue após `AbortWithError`)
- [x] 04. Ordenação no servidor (nome/tamanho/data/tipo, asc/desc) + índices + seletor na UI
- [x] 05. Favoritos e Recentes globais e paginados (`/files/starred`, `/files/recent-files`)
- [x] 06. Local físico: `/files/location/:id`, busca por caminho de disco, tier visível (quente e frio) em detalhes e listagem
- [x] 07. Copiar arquivo do tier frio falha (`os.Stat` no caminho lógico)
- [x] 08. Estado inválido após apagar/renomear/mover o item aberto
- [x] 09. Download por streaming com Range, 404 em arquivo ausente, `Content-Disposition`, pasta como zip, múltiplos
- [x] 10. Seleção (clique seleciona, duplo clique abre), multi-seleção, ações em lote, menu de contexto
- [x] 11. Upload completo: drag&drop, pasta, progresso por arquivo, política de conflito, destino correto
- [x] 12. Breadcrumb e árvore funcionando com deep link/F5 (`/files/ancestors/:id`); spinner `'pending'`
- [x] 13. Busca na pasta (recursiva, paginada) + índice trigram em `name`
- [x] 14. Barra de ações responsiva; "Criar" que na verdade reescaneia; título do FolderPicker
- [x] 15. Visão lista com colunas ordenáveis; itens como link (ctrl+click/nova aba)
- [x] 16. Painel de detalhes completo (checksum, interação, backup, metadados) e para pasta (tamanho/contagem recursivos)
- [ ] 17. Preview: mais tipos (md, código, log, mkv, m4a…), fallback com Baixar, anterior/próximo
- [ ] 18. Diálogo de apagar informa lixeira + opção definitiva; renomear seleciona só o nome
- [ ] 19. Atalhos de teclado (Enter, Delete, F2, Ctrl+A, Backspace, `/`)
- [ ] 20. Estados vazio/erro com tentar de novo e mensagem do backend
- [ ] 21. Thumbnail apaga registro do banco quando o arquivo some do disco (volume desmontado)
- [ ] 22. WebDAV ciente de tier (arquivo frio some do DAV)
- [ ] 23. "Trazer para o quente agora" a partir do detalhe
- [ ] 24. Mover entre raízes (cópia verificada + remoção)
- [ ] 25. Código morto e i18n (`FILES_EXPLORER_EYEBROW`, `.page`, subtítulo longo)
- [ ] 26. Acessibilidade de cards/linhas/breadcrumb
