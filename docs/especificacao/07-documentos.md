# 07 — Documentos

**Prioridade:** P0 · **Plataformas:** web, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Reunir os documentos da casa — contratos, contas, comprovantes, exames, manuais, planilhas — em uma biblioteca própria, organizada, onde o documento certo é achado em segundos pelo nome ou pelo que está escrito nele.

## 2. Cenários de uso

- O dono precisa da escritura da casa: abre Documentos, digita "escritura" e o PDF aparece, mesmo sem lembrar em que pasta está.
- Lembra só de um trecho ("IPTU 2024", um número de contrato): a busca encontra pelo conteúdo.
- Fotografa um comprovante com o celular e o envia; ele quer achá-lo depois como documento.
- Organiza por assunto — Casa, Carro, Saúde, Impostos — e navega por essas pastas.

## 3. Escopo

Biblioteca de documentos (lista, filtros, ordenação), busca por nome e por conteúdo, visualização, organização por pastas, indexação do texto, documentos digitalizados.

## 4. Fora de escopo

- Edição de documentos (texto, PDF, planilha) e assinatura digital.
- Digitalização pelo aplicativo (usar a câmera ou o scanner e enviar o arquivo).
- Classificação automática por assunto, extração de campos (valor, vencimento), lembretes de vencimento.
- Etiquetas manuais. *(ver D-02)*
- Versionamento de documentos.
- Conversão entre formatos.

## 5. Conceitos

- **Documento**: arquivo de um dos formatos de RN-01.
- **Pasta de documentos**: pasta configurada como biblioteca de documentos ([21](21-indexacao.md)); a organização por assunto é feita com subpastas dela.
- **Texto do documento**: conteúdo textual extraído e indexado para busca.
- **Documento digitalizado**: imagem ou PDF sem camada de texto (foto de um papel, scan).

## 6. Regras de negócio

- **RN-01** — São documentos: PDF; DOC/DOCX/ODT/RTF; XLS/XLSX/ODS/CSV; PPT/PPTX/ODP; TXT/MD; EPUB. Todo documento indexado em qualquer raiz aparece na biblioteca, esteja ou não na pasta de documentos.
- **RN-02** — Arquivos de código-fonte, logs e arquivos de configuração não são documentos (aparecem em Arquivos e na busca global).
- **RN-03** — Ordenação padrão: **data de modificação, mais recente primeiro**.
- **RN-04** — O texto é extraído de PDF com texto, DOCX, ODT, TXT, MD e CSV, em segundo plano, sem atrasar a entrada do documento na biblioteca. Limites: arquivo de até 50 MB e os primeiros 2 MB de texto extraído. Texto além de 2 MB não é indexado e o documento fica marcado como "indexado em parte"; arquivo acima de 50 MB fica como "só nome".
- **RN-05** — Documento digitalizado tem o texto reconhecido (OCR) quando o recurso de IA está ligado ([23](23-configuracoes-e-sistema.md)); sem ele, é encontrado só pelo nome, e a tela informa isso. *(ver D-01)*
- **RN-06** — A busca procura no nome e no texto, ignora maiúsculas e acentos, exige todas as palavras e ordena por relevância (nome antes de conteúdo).
- **RN-07** — Cada documento mostra a situação da indexação do texto: indexado, indexado em parte, só nome (sem texto extraível), na fila, falhou.
- **RN-08** — Mover, renomear e excluir um documento seguem [02](02-arquivos.md) e [10](10-lixeira.md); o texto indexado acompanha o documento.
- **RN-09** — O texto extraído fica só no servidor. Reconhecimento por IA usa o provedor configurado, local por padrão.

## 7. Requisitos por tela

### Documentos (`/documents`)

- **RF-01** — Lista paginada de todos os documentos, em lista (ícone do formato, nome, pasta, tamanho, modificado em, estrela) ou grade (miniatura da primeira página). Cabeçalho: total, busca, filtros, ordenação, visão, Recarregar, Enviar.
- **RF-02** — Busca no topo, por nome e conteúdo. Resultado encontrado pelo conteúdo mostra o trecho com o termo destacado e, em PDF, a página.
- **RF-03** — Filtros: formato (PDF, texto, planilha, apresentação, outros), pasta, período de modificação, só favoritos, situação de indexação.
- **RF-04** — Ordenação: modificação (padrão), nome, tamanho, data de adição, relevância (quando há busca).
- **RF-05** — Painel lateral de **pastas**: a árvore da pasta de documentos, para navegar por assunto; selecionar uma pasta restringe a lista a ela e às subpastas.
- **RF-06** — Clique abre o documento (RF-08). Menu do item e lote: os de Arquivos (Baixar, Renomear, Mover, Favoritar, Detalhes, Excluir).
- **RF-07** — "Enviar" manda arquivos para a pasta selecionada ([03](03-transferencia-entre-dispositivos.md)); "Nova pasta" cria um assunto.

### Visualizador

- **RF-08** — PDF: leitor com páginas, zoom, busca no documento, ir para página, imprimir, baixar. Ao abrir a partir de um resultado de busca, vai para a página do trecho com o termo destacado.
- **RF-09** — TXT/MD/CSV: texto formatado (Markdown renderizado, CSV em tabela).
- **RF-10** — Demais formatos: mostra o texto extraído, quando houver, com aviso de que é uma versão simplificada, e "Baixar original".
- **RF-11** — Anterior/próximo na ordem da lista; Favoritar; Detalhes (inclui a situação da indexação e "Reindexar texto"); Abrir pasta.

### Android

- **RF-12** — Lista de documentos paginada com busca por nome e conteúdo e filtro por formato; abrir PDF no app ou com "Abrir com…"; baixar.

## 8. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/documents/search` | busca por conteúdo |
| GET | `/files/search` | busca por nome |
| GET | `/files/blob/:id` · `/files/download/:id` | visualizar, baixar |
| GET | `/files/thumbnail/:id` | miniatura |

**A criar** (não existe hoje):

| Método | Rota | Uso |
|---|---|---|
| GET | `/documents` | biblioteca paginada (RF-01, RF-03, RF-04) |
| GET | `/documents/count` | total do filtro |
| GET | `/documents/:file_id/index-state` | situação da indexação do texto (RN-07) |
| POST | `/documents/:file_id/reindex` | reindexar o texto (RF-11) → `202` com `job_id` |

- **Parâmetros de `/documents` e `/documents/count`**: `format` (`pdf`, `text`, `spreadsheet`, `presentation`, `image`, `other`), `folder_id` (inclui subpastas — RF-05), `modified_from` e `modified_to` (data ISO), `favorites_only` (`true`), `index_state` (`indexed`, `partial`, `name_only`, `queued`, `failed`). `/documents` aceita ainda `page`, `page_size`, `sort` (`modified_at` — padrão —, `name`, `size`, `created_at`) e `order`. Valor inválido → `400` ([00/RN-52](00-padroes-transversais.md)).
- **Item de `/documents`** (envelope de [00 §10](00-padroes-transversais.md)): `id`, `name`, `format`, `parent_path` (caminho lógico da pasta), `size`, `modified_at`, `created_at`, `starred`, `index_state`.
- **`/documents/count`**: `{"total": 1234}`.
- **`/documents/:file_id/index-state`**: `{"state": "partial", "indexed_bytes": 2097152, "error": ""}`; `error` é a mensagem traduzida quando `state` é `failed`.
- **`/documents/search`** (já existe) passa a aceitar os mesmos filtros de `/documents` e cada resultado ganha `snippet` (trecho com o termo, quando o achado foi no conteúdo) e `page` (número da página no PDF, `null` quando não se aplica) — RF-02, RF-08.
- A árvore de pastas de RF-05 usa `/files/tree` a partir da pasta de documentos (`/libraries`); não há endpoint próprio.

## 9. Estados e erros

- Nenhum documento: estado vazio com "Enviar documentos" e a indicação de configurar a pasta de documentos.
- PDF protegido por senha: mostra que é protegido; pode ser baixado; não é indexado.
- Documento corrompido: aparece na lista; abrir explica e oferece baixar; indexação marcada como falha.
- IA desligada: aviso de que documentos digitalizados só são achados pelo nome.

## 10. Critérios de aceite

Além dos CA de [00](00-padroes-transversais.md) na lista e nos resultados de busca:

- [ ] **CA-01** — Documentos espalhados em pastas diferentes aparecem todos na biblioteca; o total confere.
- [ ] **CA-02** — Um PDF enviado agora aparece no topo da lista e, após a indexação, é encontrado por uma palavra do seu conteúdo.
- [ ] **CA-03** — Buscar um termo que só existe dentro de um DOCX o encontra, com o trecho.
- [ ] **CA-04** — Buscar "escritura" encontra `Escritura da casa.pdf`; buscar "iptu 2024" encontra o documento que tem as duas palavras em qualquer ordem.
- [ ] **CA-05** — Abrir um resultado encontrado na página 7 de um PDF abre o leitor na página 7.
- [ ] **CA-06** — Selecionar a pasta "Casa" mostra só os documentos dela e de suas subpastas.
- [ ] **CA-07** — Filtrar por "PDF" e por "este ano" mostra só o que atende aos dois.
- [ ] **CA-08** — Um PDF de 300 páginas abre em até 5 s e "ir para página" mostra qualquer página em até 2 s.
- [ ] **CA-09** — Um PDF digitalizado mostra a situação "só nome" com a IA desligada; com ela ligada, é encontrado pelo texto depois de processado.
- [ ] **CA-10** — Mover um documento de pasta não perde o texto indexado: continua sendo encontrado pelo conteúdo.
- [ ] **CA-11** — Um arquivo `.go` ou `.log` não aparece em Documentos.
- [ ] **CA-12** — No Android: buscar um documento pelo conteúdo, abrir e baixar funcionam.

## 11. Decisões em aberto

- **D-01** — Reconhecer texto de documentos digitalizados e de fotos de documentos (OCR). É o que torna "tirei foto do comprovante" pesquisável, mas depende de IA ou de ferramenta de OCR no servidor. *Provisório: em escopo usando a IA local já configurada; sem ela, só nome.*
- **D-02** — Etiquetas manuais (um documento em vários assuntos). *Provisório: fora de escopo; a organização é por pastas e favoritos.*
- **D-03** — Fotos classificadas pela IA como "documento" ([04](04-imagens.md)) aparecem também em Documentos. *Provisório: sim, com filtro de formato "Imagem".*

## 12. Rastreabilidade

`docs/revisao-produto/04-busca-global.md` lacuna 13 (conteúdo de documentos). A biblioteca de documentos é requisito novo, derivado da necessidade N4: `docs/melhorias/` tasks 44 (backend), 45 (web) e 46 (Android).
