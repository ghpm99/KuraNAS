# 45 — Documentos: tela na web

**Tipo:** feature · **Prioridade:** P0 (N4) · **Spec:** [07](../especificacao/07-documentos.md) `RF-01` a `RF-11`, `CA-02` a `CA-10`; [01](../especificacao/01-shell-e-navegacao.md) `RF-01`; [00](../especificacao/00-padroes-transversais.md)

## Contexto

Não existe tela de Documentos: o menu não tem o item e os documentos só são alcançados navegando por pastas em Arquivos ou pela busca global. A task 44 entrega a API.

## Objetivo

Uma tela `/documents` onde o dono lista, filtra, busca por nome e conteúdo e abre qualquer documento do servidor.

## O que fazer

Uma alteração por item.

1. **Serviço** `service/documents.ts`: listagem, contagem, situação, reindexar e busca, com os tipos do contrato de 07 §8 e o adaptador de paginação compartilhado (`00/RN-53`).
2. **Rota `/documents` e item "Documentos"** no grupo Principal do menu, depois de Vídeos (`01/RF-01`).
3. **Lista** (`RF-01`): visão em lista (ícone do formato, nome, pasta, tamanho, modificado em, estrela) com scroll infinito, total no cabeçalho, Recarregar e os cinco estados de `00 §5`.
4. **Visão em grade** com miniatura da primeira página e alternância lembrada no dispositivo.
5. **Ordenação** (`RF-04`) e **filtros** (`RF-03`: formato, período, só favoritos, situação de indexação), todos na URL (`00/RN-21`), com estado "vazio de filtro".
6. **Painel de pastas** (`RF-05`): árvore a partir da pasta de documentos (`/libraries` + `/files/tree`); selecionar uma pasta aplica `folder_id`.
7. **Busca** (`RF-02`): campo no topo, espera de 300 ms, resultados com trecho destacado e página do PDF; ordenação por relevância enquanto há texto.
8. **Menu do item e lote** (`RF-06`): Baixar, Renomear, Mover, Favoritar, Detalhes, Excluir — reutilizando os diálogos de Arquivos.
9. **Enviar e Nova pasta** (`RF-07`) para a pasta selecionada.
10. **Visualizador** (`RF-08` a `RF-11`): PDF com ir para página e busca, abrindo na página do trecho quando vem de um resultado; TXT/MD/CSV formatados; demais formatos com o texto extraído e "Baixar original"; anterior/próximo na ordem da lista; Detalhes com a situação da indexação e "Reindexar texto".

## Reconferência no código (2026-11-24)

Só existe `searchDocuments` no cliente. As ações, a seleção, o menu e os detalhes de Arquivos chamam `useFile()` e exigem o `FileProvider`; nenhuma lista de invalidação conhece o cache de documentos. Não há leitor de PDF além do `<embed>` nativo, nem renderização de Markdown ou CSV. O backend não gera miniatura de PDF (devolve o ícone do formato) e não expõe o texto extraído. O menu lateral é uma lista plana, sem os grupos de `01/RF-01`.

## Decisões de desenho (2026-11-24)

- **Tela** em `features/documents/`, envolvida no `FilesLayout` (como Favoritos) para reaproveitar menu, lote, diálogos e detalhes de Arquivos; um adaptador converte o item de `/documents` em `FileData` (caminho = `parent_path` + nome). As chaves de cache de documentos entram nas listas de invalidação de arquivos, de favoritos e do envio.
- **Menu**: "Documentos" entra logo depois de Vídeos; os grupos de `01/RF-01` ficam para o board de revisão.
- **Leitor de PDF**: o leitor nativo do navegador em `<iframe>` (páginas, zoom, busca, imprimir e baixar já vêm dele), aberto com `#page=N` quando vem de um resultado de busca. Sem biblioteca nova.
- **Grade**: ícone do formato no lugar da miniatura da primeira página até o backend gerar essa miniatura (ver `pendencias-do-dono.md`).
- **Demais formatos (`RF-10`)**: rota nova `GET /documents/:file_id/text` (spec 07 §8) com o texto já extraído; Markdown renderizado por um conversor próprio e seguro (sem HTML cru) e CSV em tabela, sem dependência nova.

## Critérios de aceite

Verificados no sistema rodando, com pelo menos três páginas de documentos.

- [ ] `00/CA-01` a `CA-12` e `CA-15` na lista e nos resultados de busca.
- [ ] `00/CA-16` — teste que alimenta a tela com a resposta real do handler de `/documents`.
- [ ] `07/CA-02` — PDF enviado agora aparece no topo e, depois de indexado, é achado por palavra do conteúdo.
- [ ] `07/CA-03`, `CA-04` — busca por conteúdo de DOCX mostra o trecho; "iptu 2024" acha as duas palavras em qualquer ordem.
- [ ] `07/CA-05` — abrir resultado da página 7 abre o leitor na página 7.
- [ ] `07/CA-06`, `CA-07` — pasta e filtros combinados restringem a lista; o total acompanha.
- [ ] `07/CA-08` — PDF de 300 páginas abre em até 5 s e vai a qualquer página em até 2 s.
- [ ] `07/CA-10` — mover um documento não perde a busca por conteúdo.
- [ ] F5 com filtros, ordenação e busca reabre a mesma lista.
- [ ] Render sem mock de cada componente novo; seam test das mutações.
- [ ] `make ci` verde.

## Achados fora do escopo

- `POST /files/move` responde `{"path": "/home/.../nas-data/Documentos/..."}` com o caminho do disco, não o caminho lógico — o mesmo defeito corrigido nas respostas do envio (task 35). Achado na verificação local de `07/CA-10`; vai para o board de revisão de Arquivos.

## Fora de escopo

- OCR (`07/D-01`), etiquetas (`07/D-02`).
- Edição de documentos.

## Depende de

Task 44. Usa o adaptador de paginação e o erro no fim da lista das tasks 20 e 21.
