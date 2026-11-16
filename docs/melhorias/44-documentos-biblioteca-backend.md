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
