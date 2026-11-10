# 04 — Imagens

**Prioridade:** P0 · **Plataformas:** web, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Ver todas as fotos da casa em uma galeria única, organizada por data, e encontrar uma foto antiga em poucos segundos.

## 2. Cenários de uso

- O dono tira fotos no celular, envia ([03](03-transferencia-entre-dispositivos.md)) e minutos depois elas estão no topo da galeria no PC.
- Procura "a foto do aniversário de 2019": pula para 2019 na linha do tempo e rola o mês.
- Procura uma foto pelo que aparece nela ("cachorro", "praia", o texto de um recibo).
- Monta um álbum com as fotos de uma viagem que estão espalhadas em várias pastas.

## 3. Escopo

Galeria com linha do tempo, visualizador, busca e filtros, pastas, álbuns, favoritas, seleção e lote, metadados (EXIF), miniaturas, marcação de conteúdo por IA.

## 4. Fora de escopo

- Edição de imagem (recorte, filtros, ajuste). Girar é apenas visual, no visualizador.
- Reconhecimento facial e agrupamento por pessoa.
- Mapa embutido e geocodificação; GPS aparece como coordenadas com link externo acionado pelo usuário.
- Compartilhamento por link.
- Detecção e remoção automática de duplicadas (a lista de duplicadas é de [16](16-analytics.md)).

## 5. Conceitos

- **Data da foto**: data em que a foto foi tirada (EXIF). Sem EXIF, a data de modificação do arquivo.
- **Álbum**: coleção manual de fotos, independente das pastas. Uma foto pode estar em vários álbuns; o álbum não move nem copia arquivos.
- **Categoria**: classificação automática (foto, captura de tela, documento, …) usada nas abas.

## 6. Regras de negócio

- **RN-01** — É imagem todo arquivo com extensão de imagem reconhecida, sem diferenciar maiúsculas: JPG/JPEG/JFIF, PNG, GIF, WebP, BMP, TIFF, HEIC/HEIF, AVIF e RAW de câmera. Toda imagem indexada aparece na galeria, com ou sem metadados.
- **RN-02** — A galeria é ordenada pela **data da foto, mais recente primeiro**, e agrupada por dia (com cabeçalhos de mês e ano).
- **RN-03** — Imagem sem metadados ou com metadados inválidos aparece normalmente; falta de metadado nunca gera erro nem tira a foto da lista.
- **RN-04** — Miniaturas respeitam a orientação EXIF, não distorcem a proporção e são geradas antes de serem pedidas (na indexação). Miniatura que falha mostra um ícone no lugar, sem quebrar a grade.
- **RN-05** — A galeria usa miniaturas; o original só é carregado no visualizador (em tamanho de tela) e no download.
- **RN-06** — Busca e filtros são do servidor e valem para a biblioteca inteira ([00/RN-19](00-padroes-transversais.md)).
- **RN-07** — A marcação por IA (categoria, etiquetas, legenda, texto da imagem) é opcional, roda em segundo plano com baixa prioridade e nunca atrasa a foto de aparecer na galeria. Com a IA desligada, a galeria funciona inteira; só a busca por conteúdo fica indisponível, e a tela diz isso.
- **RN-08** — A IA de imagens usa o provedor configurado, por padrão local ([23](23-configuracoes-e-sistema.md)). Nenhuma imagem sai da rede sem escolha explícita do dono.
- **RN-09** — Excluir uma foto a manda para a lixeira e a remove de todos os álbuns; restaurar a devolve aos álbuns.
- **RN-10** — Nome de álbum é obrigatório e único. Excluir um álbum não exclui as fotos.
- **RN-11** — A capa do álbum é a foto escolhida pelo dono ou, sem escolha, a mais recente do álbum.

## 7. Requisitos por tela

### Galeria (`/images`)

- **RF-01** — Grade de miniaturas agrupada por dia, com scroll infinito. Cabeçalho com o total de fotos, busca, filtros, ordenação, Recarregar e Enviar.
- **RF-02** — Abas: Todas, Fotos, Capturas de tela, Recentes (adicionadas nos últimos 30 dias, por data de adição), Favoritas, Pastas, Álbuns. Cada aba é uma lista do servidor com paginação própria.
- **RF-03** — Linha do tempo lateral com anos e meses; clicar em um mês leva a galeria direto àquele ponto, sem precisar rolar por tudo o que vem antes.
- **RF-04** — Cada miniatura mostra, ao passar o mouse ou em seleção: estrela de favorita, caixa de seleção e indicador de vídeo/ao vivo quando houver.
- **RF-05** — Clique abre o visualizador (RF-12). Seleção por caixa, Ctrl/Shift+clique ou toque longo.
- **RF-06** — Lote: Baixar, Favoritar, Adicionar a álbum, Mover para pasta, Excluir.

### Busca e filtros

- **RF-07** — Busca por nome de arquivo e, com a IA ligada, por conteúdo (etiquetas, legenda, texto na imagem).
- **RF-08** — Filtros combináveis: período (de/até), câmera, formato, categoria, só favoritas, pasta. Os valores de câmera e formato vêm do servidor, com contagem.
- **RF-09** — Ordenação: data da foto (padrão, decrescente), data de adição, nome, tamanho.

### Pastas

- **RF-10** — Lista das pastas que contêm imagens, com capa e contagem reais. Abrir uma pasta mostra a galeria restrita a ela.

### Álbuns

- **RF-11** — Lista de álbuns com capa, nome e contagem. Criar, renomear, escolher capa, excluir. Dentro do álbum: galeria das fotos, adicionar (pelo seletor ou a partir da seleção na galeria) e remover do álbum.

### Visualizador

- **RF-12** — Mostra a imagem ajustada à tela, com: anterior/próxima na ordem da lista de origem (continuando além do que estava carregado, sem dar a volta), zoom (roda do mouse, pinça, duplo clique), arrastar, girar, tela cheia, apresentação de slides.
- **RF-13** — Ações: Favoritar, Baixar original, Adicionar a álbum, Detalhes, Abrir pasta, Excluir, Fechar.
- **RF-14** — Detalhes: nome, data da foto, dimensões, tamanho, câmera e lente, exposição (abertura, tempo, ISO, distância focal), coordenadas GPS com link, caminho lógico e físico, álbuns em que está, etiquetas e legenda da IA.
- **RF-15** — Teclado: setas navegam, Esc fecha, `f` favorita, Delete exclui, `i` abre detalhes. No toque: deslizar navega, pinça dá zoom.
- **RF-16** — Fechar volta à galeria na posição da foto que estava aberta.

### Android

- **RF-17** — Galeria paginada com scroll infinito, agrupada por dia; visualizador com deslizar, pinça, favoritar, baixar, compartilhar e excluir; álbuns (ver e adicionar). Busca por nome e conteúdo.

## 8. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/files/image/library` | galeria paginada (busca, filtros, ordenação) |
| GET | `/files/image/library/count` | total do filtro |
| GET | `/files/image/library/timeline` | anos/meses com contagem |
| GET | `/files/image/library/neighbors/:file_id` | anterior/próxima no visualizador |
| GET | `/files/image/library/facets/cameras` · `/formats` | valores de filtro |
| GET | `/files/image/library/folders` | pastas com imagens |
| GET | `/files/image/metadata/:file_id` | metadados de uma imagem |
| GET, POST | `/files/image/albums` | listar, criar |
| GET, PUT, DELETE | `/files/image/albums/:id` | detalhe, editar, excluir |
| GET, POST, DELETE | `/files/image/albums/:id/items` | fotos do álbum |
| GET | `/files/thumbnail/:id` | miniatura e prévia |
| GET | `/files/images/classification/pending-count` | fila da IA |
| POST | `/files/images/classification/backfill` | reprocessar marcação |

## 9. Estados e erros

- Biblioteca sem nenhuma imagem: estado vazio com "Enviar fotos".
- Imagem cujo arquivo sumiu do disco: miniatura com ícone de indisponível; abrir explica; o registro não é apagado enquanto o volume estiver desconectado.
- IA desligada: campo de busca informa que a busca é só por nome.
- Formato que o navegador não exibe (HEIC, RAW): o servidor entrega prévia convertida; se a conversão não for possível, o visualizador mostra a explicação e "Baixar original".

## 10. Critérios de aceite

Além dos CA de [00](00-padroes-transversais.md) em cada aba, nas pastas, nos álbuns e no conteúdo do álbum:

- [ ] **CA-01** — Em uma biblioteca com mais de 1 000 fotos, rolar a galeria até o fim mostra todas; o total do cabeçalho confere.
- [ ] **CA-02** — Uma foto enviada agora aparece no topo de "Recentes" e, se tirada hoje, no topo de "Todas", com miniatura, sem ação manual.
- [ ] **CA-03** — `IMG_0001.JPG`, `foto.HEIC` e `scan.tiff` aparecem na galeria com miniatura.
- [ ] **CA-04** — Uma foto tirada na vertical aparece em pé na miniatura e no visualizador.
- [ ] **CA-05** — Clicar em um mês de três anos atrás na linha do tempo mostra as fotos daquele mês em menos de 2 s.
- [ ] **CA-06** — Filtrar por período e câmera mostra só o que atende aos dois, com o total do filtro; limpar os filtros restaura a galeria completa.
- [ ] **CA-07** — Buscar pelo nome de um arquivo que não foi carregado ainda o encontra.
- [ ] **CA-08** — Com a IA ligada e a marcação concluída, buscar por algo que aparece em uma foto a encontra; com a IA desligada, a tela informa que a busca é por nome.
- [ ] **CA-09** — No visualizador, "próxima" a partir da última foto carregada segue para a seguinte da biblioteca; na última foto de todas, "próxima" fica desabilitada.
- [ ] **CA-10** — Fechar o visualizador volta à galeria na foto que estava aberta.
- [ ] **CA-11** — Criar álbum, adicionar 30 fotos de pastas diferentes, renomear, trocar a capa e remover 5 fotos funciona; os arquivos não se moveram no disco.
- [ ] **CA-12** — Excluir uma foto que está em um álbum a tira do álbum e a manda para a lixeira; restaurar a devolve ao álbum.
- [ ] **CA-13** — Selecionar 50 fotos e favoritar marca as 50; a aba Favoritas mostra as 50.
- [ ] **CA-14** — Uma imagem sem EXIF aparece na galeria, pela data do arquivo, e seus detalhes abrem sem erro.
- [ ] **CA-15** — Detalhes de uma foto de câmera mostram data, câmera, exposição e dimensões.
- [ ] **CA-16** — No Android: rolar a galeria além de três páginas, abrir uma foto, deslizar para a próxima, favoritar e baixar funcionam.

## 11. Decisões em aberto

- **D-01** — "Recentes" como aba própria ou apenas uma ordenação ("data de adição"). *Provisório: aba própria (RF-02).*
- **D-02** — Vídeos curtos do celular aparecem na galeria junto das fotos. *Provisório: não; vídeos ficam em [06](06-videos.md).*

## 12. Rastreabilidade

`docs/revisao-produto/03-imagens.md` (19 lacunas).
