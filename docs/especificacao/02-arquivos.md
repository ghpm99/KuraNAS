# 02 — Arquivos

**Prioridade:** P0 · **Plataformas:** web, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Navegar por tudo o que está guardado no servidor como em um gerenciador de arquivos — pastas, arquivos, abrir, organizar — de qualquer dispositivo da casa.

## 2. Cenários de uso

- O dono abre a pasta "Documentos/Casa" no PC, encontra a escritura em PDF e abre para ler.
- Reorganiza a bagunça: cria pastas, move e renomeia vários arquivos de uma vez.
- No celular, entra em uma pasta e abre uma foto ou baixa um arquivo ([03](03-transferencia-entre-dispositivos.md)).
- Quer saber onde um arquivo está fisicamente no disco do servidor.

## 3. Escopo

Listagem de pastas e arquivos, navegação (breadcrumb, árvore), abrir/pré-visualizar, detalhes, criar pasta, renomear, mover, copiar, excluir, seleção e lote, ordenação, busca na pasta, recentes, localização física.

Upload e download estão em [03](03-transferencia-entre-dispositivos.md); favoritar em [09](09-favoritos.md); exclusão recuperável em [10](10-lixeira.md).

## 4. Fora de escopo

- Edição de conteúdo de arquivos (editor de texto, de imagem, de planilha).
- Versionamento de arquivos.
- Permissões por arquivo ou por pasta; compartilhamento por link.
- Compactar e descompactar pelo navegador (exceto baixar pasta como zip).
- Etiquetas (tags) manuais em arquivos.

## 5. Conceitos

- **Raiz**: pasta do disco do servidor cadastrada como área de armazenamento ([20](20-armazenamento.md)). A listagem do nível mais alto mostra as raízes.
- **Caminho lógico**: o caminho que o usuário vê e pelo qual navega. Não muda quando o arquivo é migrado de tier.
- **Caminho físico**: onde os bytes estão no disco naquele momento.
- **Item**: arquivo ou pasta.

## 6. Regras de negócio

- **RN-01** — A árvore mostrada é a do índice do servidor ([21](21-indexacao.md)), que reflete o disco. Item excluído ou na lixeira não aparece em nenhuma listagem.
- **RN-02** — Pastas e arquivos de sistema do próprio KuraNAS (lixeira, temporários) nunca aparecem.
- **RN-03** — Ordenação padrão: pastas primeiro; dentro de cada grupo, **data de modificação, mais recente primeiro**. Opções: nome, tamanho, data de modificação, tipo; crescente e decrescente. Ordenação por nome é natural (`foto2` antes de `foto10`) e ignora maiúsculas e acentos.
- **RN-04** — Toda operação (criar, renomear, mover, copiar, excluir) altera o disco **e** o índice na mesma requisição: quando a resposta chega, a listagem já reflete o resultado, sem esperar reindexação.
- **RN-05** — Nome de item: não vazio, sem `/` nem `\`, sem `..`, no máximo 255 bytes, sem caracteres que o sistema de arquivos do servidor recusa. Espaços no início e no fim são removidos. Violação responde `400` dizendo o motivo.
- **RN-06** — Nenhuma operação sobrescreve em silêncio. Destino já existente responde `409`; o cliente oferece: manter os dois (sufixo ` (2)`), substituir, pular. Em lote, a escolha pode valer para todos os conflitos.
- **RN-07** — Mover e renomear preservam a identidade do item (`id`), e com ela favorito, histórico de reprodução, participação em playlists e álbuns.
- **RN-08** — Uma pasta não pode ser movida ou copiada para dentro de si mesma ou de uma descendente (`400`).
- **RN-09** — Mover entre raízes é cópia verificada seguida de remoção da origem; se a cópia falhar, a origem fica intacta.
- **RN-10** — Excluir move para a lixeira ([10](10-lixeira.md)). Exclusão definitiva é opção explícita, com confirmação própria.
- **RN-11** — Toda operação é confinada às raízes cadastradas. Caminho que escape de uma raiz responde `400`.
- **RN-12** — Operação sobre item que não existe mais responde `404` e a listagem é atualizada.
- **RN-13** — Arquivo no tier frio se comporta como qualquer outro (abre, copia, move, baixa); a única diferença visível é o indicador de tier.
- **RN-14** — Contagem de filhos e tamanho de uma pasta consideram só itens existentes (não os da lixeira); o tamanho de uma pasta é a soma recursiva do conteúdo.

## 7. Requisitos por tela

### Listagem (`/files/<caminho>`)

- **RF-01** — Mostra o conteúdo da pasta em grade (ícone/miniatura, nome) ou lista (nome, tamanho, tipo, data de modificação; colunas ordenáveis pelo cabeçalho). A escolha de visão é lembrada no dispositivo.
- **RF-02** — Cada item mostra: ícone por tipo ou miniatura (imagem, vídeo, capa de áudio), nome, e os indicadores de favorito e de tier frio. Pasta mostra a contagem de itens.
- **RF-03** — Cabeçalho: breadcrumb clicável da raiz até a pasta atual, total de itens da pasta, busca na pasta, ordenação, alternância grade/lista, Recarregar, Nova pasta, Enviar.
- **RF-04** — Paginação, scroll infinito, recarregar e estados conforme [00](00-padroes-transversais.md). Estado vazio de pasta oferece "Enviar arquivos" e "Nova pasta".
- **RF-05** — Clique seleciona; duplo clique ou Enter abre. Pasta abre na listagem; arquivo abre no visualizador (RF-10). No toque, um toque abre e toque longo seleciona.
- **RF-06** — Menu de contexto (botão direito, toque longo ou "⋯") no item: Abrir, Baixar, Renomear, Mover para…, Copiar para…, Favoritar/Desfavoritar, Detalhes, Excluir.
- **RF-07** — Com seleção, a barra de lote oferece: Baixar (zip quando mais de um), Mover, Copiar, Favoritar, Excluir.
- **RF-08** — Atalhos: Enter abre, F2 renomeia, Delete exclui, Ctrl+A seleciona tudo, Backspace sobe um nível, `/` foca a busca.

### Árvore de pastas

- **RF-09** — No menu lateral, árvore das raízes e subpastas, expandindo sob demanda. A pasta atual fica destacada e visível, inclusive ao abrir a tela por endereço direto.

### Visualizador (`/files/<caminho do arquivo>`)

- **RF-10** — Abre o arquivo conforme o tipo:

| Tipo | Como abre |
|---|---|
| Imagem | visualizador de imagem ([04/RF-12](04-imagens.md)) |
| Vídeo | player de vídeo ([06](06-videos.md)) |
| Áudio | toca no player de música ([05](05-musica.md)) |
| PDF | leitor embutido |
| Texto, Markdown, código, log | texto com quebra de linha; arquivos grandes mostram o início e avisam |
| Outros | cartão com ícone, nome, tamanho e "Baixar" |

- **RF-11** — O visualizador tem: nome, anterior/próximo dentro da mesma pasta (na ordenação atual), Baixar, Favoritar, Detalhes, Excluir, Fechar (volta à pasta na posição em que estava).
- **RF-12** — Arquivo que não pode ser exibido (formato sem suporte, arquivo corrompido) mostra a explicação e "Baixar" — nunca tela preta ou em branco.

### Painel de detalhes

- **RF-13** — Para arquivo: nome, tipo, tamanho, caminho lógico, caminho físico e tier, criado em, modificado em, último acesso, checksum, situação de backup, e os metadados do tipo (dimensões e EXIF de imagem; duração e tags de áudio; duração, resolução e codec de vídeo).
- **RF-14** — Para pasta: nome, caminho, quantidade de arquivos e de subpastas (recursivo), tamanho total, modificado em.
- **RF-15** — O caminho físico pode ser copiado com um clique.

### Diálogos de operação

- **RF-16** — **Nova pasta**: campo de nome com validação (RN-05); ao criar, a pasta aparece na listagem e fica selecionada.
- **RF-17** — **Renomear**: campo preenchido com o nome atual, com apenas o nome (sem a extensão) selecionado. Mudar a extensão pede confirmação.
- **RF-18** — **Mover / Copiar**: seletor de pasta em árvore, paginado, com "Nova pasta" dentro do seletor; não permite escolher destino inválido (RN-08).
- **RF-19** — **Excluir**: informa quantos itens vão para a lixeira e oferece a opção "excluir definitivamente".
- **RF-20** — Operação em muitos itens ou em item grande mostra progresso e pode continuar em segundo plano ([00/RN-39](00-padroes-transversais.md)).

### Busca na pasta

- **RF-21** — Busca por nome dentro da pasta atual e de suas subpastas, no servidor, paginada. Cada resultado mostra o caminho relativo; "abrir local" leva à pasta do item.

### Recentes

- **RF-22** — Lista global dos arquivos abertos ou adicionados recentemente, paginada, ordenada do mais recente para o mais antigo.

### Android

- **RF-23** — Explorer com listagem paginada, navegação por pastas, abrir por tipo, favoritar, renomear, mover, excluir, nova pasta, detalhes, baixar e enviar ([03](03-transferencia-entre-dispositivos.md)). Paridade completa em [24](24-app-android.md).

## 8. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/files/tree` | listagem de uma pasta (paginada, ordenável) |
| GET | `/files/:id` | um item |
| GET | `/files/path` | item por caminho |
| GET | `/files/ancestors/:id` | cadeia de pastas para breadcrumb/árvore |
| GET | `/files/folder-stats/:id` | contagem e tamanho recursivos |
| GET | `/files/location/:id` | caminho físico e tier |
| GET | `/files/by-disk-path` | item a partir do caminho no disco |
| GET | `/files/search` | busca na pasta |
| GET | `/files/recent-files` | recentes |
| GET | `/files/thumbnail/:id` | miniatura |
| GET | `/files/blob/:id` | conteúdo para visualização |
| POST | `/files/folder` | criar pasta |
| POST | `/files/rename` | renomear |
| POST | `/files/move` | mover |
| POST | `/files/copy` | copiar |
| DELETE | `/files/path` | excluir (`permanent=true` para definitivo) |

## 9. Estados e erros

| Situação | Resposta | O que o usuário vê |
|---|---|---|
| Pasta do endereço não existe | `404` | "Pasta não encontrada" + voltar para a raiz |
| Nome inválido | `400` | motivo, no próprio diálogo |
| Destino já existe | `409` | opções de RN-06 |
| Item sumiu durante a operação | `404` | aviso e listagem atualizada |
| Disco cheio / sem permissão no disco | `500` com mensagem específica | a mensagem do servidor |
| Volume da raiz desconectado | — | a raiz aparece marcada como indisponível; nada é removido do índice |

## 10. Critérios de aceite

Além de todos os CA de [00](00-padroes-transversais.md) aplicados à listagem, ao seletor de pasta, à busca na pasta e a Recentes:

- [ ] **CA-01** — Abrir uma pasta com mais de 500 itens e rolar até o fim mostra todos; o total do cabeçalho confere com o disco.
- [ ] **CA-02** — Ordenar por tamanho decrescente coloca o maior arquivo da pasta em primeiro, mesmo que ele estivesse na última página.
- [ ] **CA-03** — Criar pasta, renomear, mover e copiar aparecem na listagem imediatamente e estão no disco do servidor.
- [ ] **CA-04** — Mover um arquivo favoritado mantém a estrela.
- [ ] **CA-05** — Copiar para uma pasta onde já existe o mesmo nome oferece as opções de conflito; "manter os dois" cria ` (2)`.
- [ ] **CA-06** — Tentar mover uma pasta para dentro dela mesma é impedido com mensagem clara.
- [ ] **CA-07** — Excluir manda para a lixeira e o item some de Arquivos, Imagens/Música/Vídeos e Favoritos.
- [ ] **CA-08** — Selecionar 20 itens e mover todos funciona e informa o resultado; se 2 falharem, os 18 são movidos e os 2 são listados com o motivo.
- [ ] **CA-09** — Abrir por endereço direto uma pasta três níveis abaixo mostra breadcrumb completo e a árvore expandida até ela.
- [ ] **CA-10** — PDF, imagem, vídeo, áudio, `.txt` e `.md` abrem no visualizador; um `.zip` mostra o cartão com "Baixar".
- [ ] **CA-11** — No visualizador, "próximo" percorre os arquivos da pasta na ordenação atual, inclusive além da página carregada.
- [ ] **CA-12** — Fechar o visualizador volta à pasta na mesma posição de rolagem.
- [ ] **CA-13** — Detalhes de uma pasta mostram tamanho e contagem recursivos corretos.
- [ ] **CA-14** — Um arquivo no tier frio abre, copia e baixa normalmente e mostra o caminho físico do tier frio nos detalhes.
- [ ] **CA-15** — Buscar na pasta encontra um arquivo que está em subpasta e em página ainda não carregada.
- [ ] **CA-16** — Um arquivo colocado direto no disco do servidor aparece na listagem após a indexação, sem ação do usuário ([21](21-indexacao.md)).
- [ ] **CA-17** — No Android: navegar três níveis, abrir uma foto, renomear um arquivo e excluir outro funcionam e refletem na web.

## 11. Decisões em aberto

- **D-01** — Ordenação padrão: o dono citou "data desc"; hoje a listagem pode estar por nome. *Provisório: RN-03 (pastas primeiro, modificação mais recente primeiro).*
- **D-02** — Clique simples abre ou seleciona na web. *Provisório: RF-05 (clique seleciona, duplo clique abre), como gerenciadores de arquivos de desktop.*

## 12. Rastreabilidade

`docs/revisao-produto/01-arquivos.md` (26 lacunas); `docs/melhorias/` tasks 01, 02, 05, 08 (índice e consistência), 10 (raízes), 13 (tiering).
