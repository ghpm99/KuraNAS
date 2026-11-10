# 08 — Busca global

**Prioridade:** P0 · **Plataformas:** web, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Achar qualquer coisa guardada no servidor — arquivo, pasta, foto, música, vídeo, documento — digitando o que o dono lembra dela, de qualquer tela.

## 2. Cenários de uso

- O dono lembra só de parte do nome: digita "escritura" e o PDF aparece, com a pasta onde está.
- Lembra do que está escrito dentro do documento, não do nome.
- Digita o nome de uma música e a toca direto do resultado.
- Quer ir rápido para uma tela: digita "lixeira" e Enter.

## 3. Escopo

Campo de busca no cabeçalho, painel de resultados agrupados, página de resultados completa com filtros, busca por conteúdo de documentos, expansão por IA sob demanda, histórico de buscas, ações rápidas de navegação.

## 4. Fora de escopo

- Busca na internet.
- Busca em e-mails (fica na tela de [19](19-email.md)).
- Operadores avançados (`AND`, `OR`, aspas, expressões regulares).
- Busca por conteúdo de imagem fora da galeria (a global mostra fotos por nome e etiquetas; a busca rica é de [04](04-imagens.md)).
- Busca semântica por vetores.

## 5. Regras de negócio

- **RN-01** — A busca procura em: nome de arquivos e pastas; título, artista, álbum e gênero de músicas; nome de vídeos, séries e playlists; nome, etiquetas e legenda de fotos; nome de álbuns de fotos; conteúdo de documentos ([07](07-documentos.md)).
- **RN-02** — A comparação ignora maiúsculas e acentos e tolera um erro de digitação em termos de 5 letras ou mais. Várias palavras são todas obrigatórias, em qualquer ordem.
- **RN-03** — Caracteres especiais (`%`, `_`, `\`) são tratados como texto.
- **RN-04** — Itens na lixeira nunca aparecem.
- **RN-05** — Dentro de cada grupo, a ordem é por relevância: nome igual ao termo, nome que começa com o termo, nome que contém o termo, conteúdo que contém o termo. Empate: mais recente primeiro.
- **RN-06** — A resposta principal não espera IA. A expansão por IA ("Buscar com IA") é uma ação explícita, que acrescenta resultados aos já exibidos.
- **RN-07** — Busca com menos de 2 caracteres não consulta o servidor.
- **RN-08** — O histórico de buscas é por dispositivo, guarda as 10 últimas e pode ser apagado.
- **RN-09** — A primeira resposta chega em até 500 ms em uma biblioteca de 100 mil itens.

## 6. Requisitos por tela

### Painel de busca (no cabeçalho, em qualquer tela)

- **RF-01** — Abre ao focar o campo ou com `/` / Ctrl+K. Vazio, mostra as buscas recentes e as ações rápidas.
- **RF-02** — Ao digitar (espera de 300 ms, pedido anterior cancelado), mostra resultados agrupados: Pastas, Arquivos, Documentos, Fotos, Músicas, Vídeos, Playlists e álbuns, Ações. Até 5 itens por grupo, cada grupo com "Ver todos (N)".
- **RF-03** — Cada resultado mostra ícone ou miniatura, nome com o trecho encontrado destacado, caminho ou contexto (artista — álbum; série — episódio), tamanho, data, estrela e tier. Documento encontrado pelo conteúdo mostra o trecho em volta do termo.
- **RF-04** — Enquanto uma nova busca carrega, os resultados anteriores continuam visíveis.
- **RF-05** — Setas navegam pelos resultados, Enter abre o selecionado (por padrão, o melhor resultado de dado), Esc fecha.
- **RF-06** — Abrir um resultado leva ao lugar dele: arquivo no visualizador dentro da pasta, pasta em Arquivos, foto no visualizador da galeria, música tocando, vídeo no player, playlist/álbum na tela dele.
- **RF-07** — Ações por resultado sem abrir: favoritar, "abrir pasta", tocar (música).
- **RF-08** — Ações rápidas: ir para cada tela e subseção do aplicativo, incluindo as seções de Configurações.
- **RF-09** — "Buscar com IA" aparece quando há provedor de IA configurado; mostra a consulta sugerida e acrescenta os resultados.

### Página de resultados (`/files?search=…` ou equivalente por tipo)

- **RF-10** — "Ver todos" abre a lista completa do grupo, paginada, com filtros: tipo, período de modificação, faixa de tamanho, tier, pasta, só favoritos; ordenação por relevância (padrão), data, nome, tamanho.
- **RF-11** — O termo, os filtros e a ordenação ficam na URL.

### Android

- **RF-12** — Tela de busca com os mesmos grupos, "ver todos" paginado e abertura do resultado na tela correspondente do app.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/search/global` | resultados agrupados |
| GET | `/search/global/ai` | expansão por IA |
| GET | `/files/search` | lista completa de arquivos/pastas com filtros |
| GET | `/documents/search` | conteúdo de documentos |
| GET | `/files/image/library` · `/music/search` · `/video/library/search` | "ver todos" por tipo |

## 8. Estados e erros

- Sem resultado: "nada encontrado para «termo»", com sugestões (conferir a grafia, buscar com IA quando disponível).
- Falha na busca: mensagem com "Tentar de novo", mantendo o que já estava na tela.
- Um grupo que falha não impede os outros de aparecer.
- IA indisponível ou lenta: a ação mostra o erro; os resultados normais não são afetados.

## 9. Critérios de aceite

Além dos CA de [00](00-padroes-transversais.md) na página de resultados e em cada "ver todos":

- [ ] **CA-01** — Buscar parte do nome de um arquivo que está em uma pasta nunca aberta o encontra.
- [ ] **CA-02** — "acao" encontra "Ação"; "FOTO" encontra "foto"; "relatorio 2023" encontra "2023 - Relatório final".
- [ ] **CA-03** — Buscar "100%" não devolve todos os arquivos.
- [ ] **CA-04** — Um termo que só existe dentro de um PDF encontra o PDF, com o trecho.
- [ ] **CA-05** — Buscar o título de uma música mostra a faixa no grupo Músicas e tocá-la funciona sem sair da tela.
- [ ] **CA-06** — Abrir um resultado de arquivo leva à pasta certa com o arquivo aberto, inclusive com espaços e acentos no caminho.
- [ ] **CA-07** — Digitar rápido dez letras gera resultado só da consulta final; resposta atrasada de consulta anterior não aparece.
- [ ] **CA-08** — Um arquivo na lixeira não aparece nos resultados.
- [ ] **CA-09** — "Ver todos" em Arquivos com mais de 200 resultados rola até o fim; filtrar por tipo e período reduz a lista e o total.
- [ ] **CA-10** — Com a IA desligada, a busca funciona igual e a ação de IA não aparece.
- [ ] **CA-11** — Com o painel aberto, é possível buscar, navegar e abrir um resultado só com o teclado.
- [ ] **CA-12** — As buscas recentes aparecem ao abrir o painel e podem ser apagadas.
- [ ] **CA-13** — No Android: buscar, abrir uma foto e uma música pelos resultados funciona.

## 10. Decisões em aberto

Nenhuma.

## 11. Rastreabilidade

`docs/revisao-produto/04-busca-global.md` (13 lacunas).
