# 44 — Documentos: biblioteca no backend

**Tipo:** feature · **Prioridade:** P0 (N4) · **Spec:** [07](../especificacao/07-documentos.md) `RN-01` a `RN-04`, `RN-07`, `RN-08`, `RF-01`, `RF-03`, `RF-04`, `CA-01`, `CA-11`

## Contexto

"Organizar documentos e achar o documento certo" é necessidade direta do dono (N4). Hoje o backend só tem a busca por conteúdo (`GET /documents/search`, domínio `documenttext`, tabela `document_text`). Não existe listagem de documentos, contagem, nem a situação da indexação do texto — a tela de Documentos não tem de onde ler (spec 07 §8, "A criar").

## Objetivo

O backend expõe a biblioteca de documentos: lista paginada com filtros e ordenação, total, situação da indexação e reindexação do texto de um documento.

## O que fazer

Tudo no domínio `documenttext` (extensão que importa `files`; `files` não passa a conhecer documentos). Um endpoint por informação, um `.sql` por consulta.

1. **Formatos** (`RN-01`, `RN-02`): função única que diz se uma extensão é documento e a qual grupo pertence (`pdf`, `text`, `spreadsheet`, `presentation`, `other`). Código-fonte, logs e configuração ficam de fora.
2. **Situação da indexação** (`RN-07`): a linha de `document_text` passa a registrar `state` (`indexed`, `partial`, `name_only`, `queued`, `failed`), `indexed_bytes` e o motivo da falha. `partial` quando o texto passa de 2 MB; `name_only` para arquivo acima de 50 MB, PDF sem texto e PDF com senha.
3. **`GET /documents`** — lista paginada no envelope padrão, com os filtros e a ordenação do contrato de 07 §8; desempate por `id` (`00/RN-17`); valor inválido → `400`.
4. **`GET /documents/count`** — mesmo filtro, `{"total"}`.
5. **`GET /documents/:file_id/index-state`**.
6. **`POST /documents/:file_id/reindex`** → `202` com `job_id`; `404` se não é documento.
7. **`GET /documents/search`** aceita os mesmos filtros e devolve `snippet` e `page` por resultado.

## Reconferência no código (2026-11-24)

`document_text` não guarda estado nem bytes indexados (só `truncated` e um código de erro); o extrator guarda no máximo 200 KB de texto (a spec fala em 2 MB); PDF digitalizado vira texto vazio sem erro; o extrator de PDF não sabe a página; DOC, ODT, RTF, XLS/XLSX, ODS, PPT/PPTX, ODP e EPUB não têm extrator; a busca só acha documentos com texto indexado (não acha pelo nome); `utils.DocumentFormats` não é a lista de `07/RN-01` e é usada por `/files/search` e pelo Analytics; não existe reindexação de um arquivo só.

## Decisões de desenho (2026-11-24)

- **Formatos**: função própria no `documenttext` com o mapa extensão → grupo de `RN-01`; `utils.DocumentFormats` fica como está.
- **Estado**: migração em `document_text` com `state`, `indexed_bytes` e `failure_reason`; documento sem linha (ou com linha antiga) é `queued`. Texto guardado até 2 MB (`RN-04`): acima disso `partial`. Acima de 50 MB, PDF sem texto, PDF com senha e formato sem extrator → `name_only`; erro de leitura → `failed` com o motivo traduzido.
- **Página**: o extrator de PDF passa a separar páginas (caractere de quebra de página entre elas, seguindo a árvore de páginas); a busca calcula `page` pela posição do trecho. Sem páginas conhecidas, `page` vem nulo.
- **Busca**: casa nome ou conteúdo (nome primeiro), aceita os filtros da lista; `snippet` só quando o achado foi no conteúdo.
- **Reindexar**: job do tipo de indexação de texto com escopo de um arquivo; pedido repetido devolve `202` com o job já em andamento. `folder_id` inexistente → `404`; `format=image` aceito e vazio (reservado).

## Critérios de aceite

- [ ] `07/CA-01` — documentos em pastas e raízes diferentes aparecem todos em `GET /documents`; `GET /documents/count` confere.
- [ ] `07/CA-11` — um `.go` e um `.log` não aparecem.
- [ ] Filtrar por `format=pdf` e por período devolve só o que atende aos dois; o total do filtro confere.
- [ ] `folder_id` devolve os documentos da pasta e das subpastas.
- [ ] Ordenar por `size` decrescente põe o maior documento da biblioteca em primeiro, com `page_size=1`.
- [ ] Um PDF digitalizado aparece com `index_state=name_only`; um DOCX de 5 MB de texto, com `partial`.
- [ ] Um resultado de busca achado na página 7 de um PDF traz `page: 7` e o trecho.
- [ ] Documento na lixeira não aparece.
- [ ] Testes contra Postgres real para cada consulta; teste de handler decodificando a resposta real e fixando os nomes de campo.
- [ ] `make ci-backend` verde (cobertura ≥ 80 %).

## Fora de escopo

- Tela web (task 45) e Android (task 46).
- OCR de digitalizados (`07/D-01`) e fotos classificadas como documento (`07/D-03`): entram depois, sem mudar este contrato (`format=image` já está reservado).
