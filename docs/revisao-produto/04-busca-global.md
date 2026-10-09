# 04 — Busca global

Escopo: `backend/internal/api/v1/search/` + `pkg/database/queries/search/*.sql`, `frontend/src/components/search/**`, `service/search.ts`, e a ligação com as buscas por domínio (`/files/search`, `/image/library?q=&content=`).

## Diagnóstico (2026-10-09)

- Nenhuma `search_*.sql` usa o índice trigram (`hf.name ILIKE` ≠ `lower(name)`), e o `OR hf.path ILIKE '%…%'` força seq-scan: 8 varreduras de `home_file` por tecla, sem debounce nem cancelamento.
- A expansão por IA roda antes da resposta (até 10 s), ignora o limite por grupo e devolve uma `suggestion` que a UI nunca mostra.
- Sem faixas de música; várias palavras só casam como substring contígua; `%`/`_` não escapados; sem acento/typo; "ver todos" só para arquivos; Enter abre uma ação em vez do melhor resultado; caminho não codificado na navegação; sem erro, histórico ou destaque.
- Busca pelo conteúdo de documentos não existe.

Decisões tomadas na revisão: IA sai do caminho principal e vira ação explícita ("Buscar com IA"). Insensibilidade a acento via extensão `unaccent` criada defensivamente (como o `pg_trgm` da 0048) — sem ela a busca segue só case-insensitive. Conteúdo de documentos indexado para formatos de texto/código, Markdown, PDF e DOCX com limites de tamanho, em passo de worker de baixa prioridade.

## Lacunas

- [x] 01. Queries da busca global usam o índice trigram (`lower(name) LIKE`, sem `OR path`), curingas escapados, várias palavras como AND de termos
- [ ] 02. IA fora do caminho principal: resposta imediata, expansão sob demanda, merge respeitando o limite, `suggestion` exibida
- [ ] 03. Dialog: debounce, cancelamento, mantém o resultado anterior, estado de erro com tentar de novo
- [ ] 04. Enter abre o melhor resultado de dado; ações depois dos dados quando há consulta
- [ ] 05. Navegação com caminho codificado e playlist de vídeo sem colisão por nome
- [ ] 06. Faixas de música como grupo próprio (tocar/abrir álbum); mídia fora do grupo "Arquivos"
- [ ] 07. "Ver todos" para cada grupo (fotos, vídeos, música, arquivos)
- [ ] 08. Página completa de resultados de arquivos com filtros (tipo, data, tamanho, tier) e ordenação por relevância
- [ ] 09. Resultados com tamanho, data, favorito e tier
- [ ] 10. Insensível a acento e tolerante a erro de digitação
- [ ] 11. Histórico de buscas recentes e destaque do trecho encontrado
- [ ] 12. Ações rápidas completas (Lixeira, Downloads, Capturas, Assistente, Notificações, Takeout, Diário, subseções, seções de Configurações)
- [ ] 13. Busca pelo conteúdo de documentos (texto, Markdown, código, PDF, DOCX)
