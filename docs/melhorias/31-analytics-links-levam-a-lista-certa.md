# 31 — Analytics: "Ver duplicados", "Ver recentes" e "Ver maiores" abrem a lista certa

**Tipo:** bug · **Prioridade:** P2 · **Spec:** [16](../especificacao/16-analytics.md) `RN-06`, `RF-06` a `RF-09`, §7, `CA-05` a `CA-08`

## Contexto

Os três botões da visão geral e as linhas das tabelas fazem todos `navigate(appRoutes.files)` — abrem `/files`, sem filtro nem destino (`frontend/src/components/analytics/AnalyticsOverviewScreen.tsx:51-59`). Quem clica em "Ver duplicados" cai na raiz do explorer. Não existe tela de duplicados nem de maiores arquivos; os endpoints `/analytics/duplicates/groups` e `/files/top-files-by-size` não paginam.

## Objetivo

Cada número e cada linha do Analytics leva à lista que representa.

## O que fazer

1. **Backend — paginação** em `/analytics/duplicates/groups` (com filtro `type`) e em `/files/top-files-by-size`, opt-in por `page` (contrato de 16 §7; sem `page` a resposta atual é mantida).
2. **Tela `/analytics/largest`** (`RF-07`): lista paginada do maior para o menor, com nome, pasta, tamanho e tipo; cada linha abre o arquivo.
3. **Tela `/analytics/duplicates`** (`RF-09`): lista paginada de grupos, do maior desperdício para o menor, com filtro por tipo; cada grupo expande as cópias (caminho, data, tier) com abrir, abrir pasta e excluir (para a lixeira).
4. **Excluir em lote** nas cópias selecionadas, com confirmação extra quando a seleção inclui todas as cópias de um grupo.
5. **Ligações da visão geral**: "Ver maiores" → `/analytics/largest`; "Ver duplicados" → `/analytics/duplicates`; "Ver recentes" → a tela Recentes de Arquivos; linha de pasta → a pasta em Arquivos; linha de arquivo → o arquivo.

### Sub-itens

- [x] 31.1. A distribuição por tipo da visão geral (`analytics/type_distribution.sql`) usa listas de formatos fixas e desatualizadas (sem RAW, `.opus`, `.heif`…); passar a usar as listas canônicas de `pkg/utils` (as mesmas da indexação e do filtro de tipo dos duplicados), para as contagens por tipo baterem entre as telas.

## Critérios de aceite

- [x] `16/CA-05` — clicar em uma pasta da lista de maiores abre essa pasta em Arquivos. *Local 2026-11-17: a verificação achou o Analytics devolvendo caminho do disco (corrigido em `618e6f14`); `/analytics/top-folders` → `/Musicas/teste-task40/Generos`, que `/files/path` resolve para a pasta (id 871); a linha navega com `buildFilesUrl(path)`.*
- [x] `16/CA-06` — "ver todos" em maiores arquivos abre a lista ordenada por tamanho, com scroll infinito. *Local: `/files/top-files-by-size?page=1&page_size=3` em ordem de tamanho com `has_next: true`; scroll infinito coberto pelos testes da tela.*
- [x] `16/CA-07` — copiar um arquivo grande cria um grupo de duplicados; excluir uma cópia tira o grupo e atualiza o espaço recuperável. *Local: cópia de `rock1.mp3` → grupo novo, resumo 16.879 B recuperáveis; `DELETE /files/path {id}` (a chamada da tela) → grupo some, resumo 248 B.*
- [x] `16/CA-08` — a lista de duplicados com mais de 100 grupos rola até o fim. *Teste `_Postgres` caminha 137 grupos sem repetir nem pular; scroll até o fim coberto pelos testes da tela.*
- [x] Os CA de `00` (`CA-01` a `CA-12`, `CA-16`) nas duas telas novas.
- [x] Testes de handler decodificando o envelope paginado e a resposta sem `page`; render sem mock das telas; seam test do excluir.
- [x] `make ci` verde.

## Fora de escopo

- Blocos do Analytics que mostram zero no lugar de erro (`16/RN-03`), espaço total de todas as raízes (`RN-02`), espaço da lixeira e reprocessar falhas: item 12 do board de revisão.
- Remoção automática de duplicados.

## Depende de

Tasks 20 e 21 (as telas novas usam o adaptador de paginação e o erro no fim da lista).
