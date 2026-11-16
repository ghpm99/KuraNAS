# 01 — Shell e navegação

**Prioridade:** P0 · **Plataformas:** web (a navegação do Android está em [24](24-app-android.md))

## 1. Objetivo

Dar a todas as telas a mesma moldura — menu, cabeçalho, área de conteúdo, player persistente — de modo que o dono chegue a qualquer feature em um clique, em qualquer tamanho de tela, e nunca veja uma tela em branco.

## 2. Cenários de uso

- No PC, o dono alterna entre Arquivos, Imagens, Música e Vídeos pelo menu lateral enquanto a música continua tocando.
- No tablet e no celular (navegador), usa a barra inferior e o conteúdo ocupa a tela inteira.
- O servidor está desligado: a tela avisa e oferece tentar de novo, em vez de ficar carregando para sempre.

## 3. Escopo

Menu lateral, barra inferior, cabeçalho (busca global, sino de notificações), área de conteúdo com rolagem, player persistente de música, página 404, tratamento de erro por rota, aviso de servidor fora do ar, tema, atalhos globais, título da aba.

## 4. Fora de escopo

- Conteúdo de cada tela (specs próprias).
- Menu configurável pelo usuário (ordem, esconder itens).
- Múltiplas janelas/abas sincronizadas entre si.

## 5. Regras de negócio

- **RN-01** — Toda rota da aplicação renderiza dentro do shell. Não existe tela sem menu, exceto o player de vídeo em tela cheia e o assistente de configuração inicial.
- **RN-02** — O shell monta mesmo sem resposta do backend. Falha ao carregar traduções, configurações ou qualquer dado global não impede o menu e a navegação de aparecerem.
- **RN-03** — Só a área de conteúdo rola. Menu, cabeçalho, barra inferior e player ficam fixos. Não há rolagem horizontal da página em nenhuma largura a partir de 320 px.
- **RN-04** — A navegação é por URL: toda tela e todo estado relevante (pasta aberta, aba, busca, item aberto) tem endereço próprio; F5 e "voltar" do navegador reproduzem o que estava na tela.
- **RN-05** — Erro em uma tela fica contido nela: o menu continua funcional e navegar para outra rota limpa o erro.
- **RN-06** — A música não para ao trocar de rota. O player persistente nunca cobre conteúdo, barra inferior ou avisos.
- **RN-07** — Preferências de aparência valem em todos os dispositivos (salvas no servidor); o estado recolhido do menu é por dispositivo.

## 6. Requisitos por tela

### Menu lateral (telas largas — a partir do breakpoint `tablet` do tema)

- **RF-01** — Itens em dois grupos, nesta ordem. **Principal:** Home, Arquivos, Imagens, Música, Vídeos, Documentos, Favoritos, Lixeira. **Ferramentas:** Baixar de URL, Capturas, Takeout, Tarefas, Assistente, E-mails, Analytics, Diário, Robô aspirador, Aplicativos, Configurações, Sobre. O item da rota atual fica destacado. Toda tela que existe tem item no menu — nenhuma é alcançável só digitando o endereço.
- **RF-02** — Recolhível para modo só ícones, com rótulo em tooltip. O estado persiste no dispositivo.
- **RF-03** — Quando os itens não cabem na altura, o próprio menu rola.
- **RF-04** — Em `/files`, o menu mostra a árvore de pastas abaixo dos itens ([02/RF-09](02-arquivos.md)).

### Barra inferior (telas estreitas — abaixo do breakpoint `tablet`)

- **RF-05** — Cinco destinos: Home, Arquivos, Imagens, Música, Mais. "Mais" abre os demais itens do menu.
- **RF-06** — Alvos de toque de no mínimo 44 × 44 px; respeita a área segura do aparelho.

### Cabeçalho

- **RF-07** — Campo de busca global ([08](08-busca-global.md)), sino de notificações com contador de não lidas ([15](15-notificacoes-e-jobs.md)) e, em tela estreita, o título da tela atual.

### Área de conteúdo

- **RF-08** — Ao navegar para uma rota nova, a rolagem começa no topo; ao voltar pelo histórico, a posição anterior é restaurada.
- **RF-09** — Enquanto o código da rota carrega, aparece indicador de carregamento dentro do shell.
- **RF-10** — Rota inexistente mostra página "não encontrada" com link para a Home.

### Player persistente

- **RF-11** — Com música na fila, aparece uma barra de player fixa: capa, título, artista, anterior, tocar/pausar, próxima, progresso. Em tela estreita fica acima da barra inferior e expande em painel com os controles completos ([05/RF-20](05-musica.md)).

### Servidor fora do ar

- **RF-12** — Sem resposta do backend, um aviso fixo no topo informa a falha e oferece "Tentar de novo". Quando a conexão volta, o aviso some e as telas recarregam sozinhas.

### Tema e acessibilidade

- **RF-13** — Tema claro, escuro ou do sistema; cor de destaque configurável ([23](23-configuracoes-e-sistema.md)).
- **RF-14** — Navegação completa por teclado, com foco visível; link "pular para o conteúdo"; ao trocar de rota o foco vai para o conteúdo; respeita `prefers-reduced-motion`.
- **RF-15** — Título da aba do navegador reflete a tela atual.
- **RF-16** — Atalhos globais: `/` foca a busca, `?` abre a ajuda de atalhos, `g` seguido de uma letra navega para a seção. Atalhos não disparam enquanto o foco está em campo de texto.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/configuration/translation` | catálogo de textos |
| GET | `/configuration/settings` | aparência e preferências |
| GET | `/health` | detecção de servidor fora do ar |
| GET | `/notifications/unread-count` | contador do sino |

## 8. Estados e erros

- Falha ao buscar traduções: o shell usa o catálogo embutido de reserva; nenhuma chave crua (`FILES_TITLE`) aparece na tela.
- Erro de renderização em uma tela: mensagem traduzida, botão "Tentar de novo" e detalhes técnicos recolhidos.

## 9. Critérios de aceite

- [ ] **CA-01** — Com o backend desligado, abrir a aplicação mostra menu, cabeçalho e o aviso de servidor fora do ar; nenhuma tela em branco, nenhuma chave de tradução crua.
- [ ] **CA-02** — Religar o backend faz o aviso sumir e a tela atual carregar sem recarregar a página.
- [ ] **CA-03** — Cada item do menu abre a tela correspondente e fica destacado; F5 em qualquer uma delas reabre a mesma tela.
- [ ] **CA-04** — Em 360 × 640, 768 × 1024 e 1920 × 1080 não há rolagem horizontal em nenhuma tela do índice.
- [ ] **CA-05** — Em 360 × 640, com música tocando, barra inferior, player e conteúdo são todos alcançáveis; o último item de qualquer lista não fica escondido atrás do player.
- [ ] **CA-06** — Em janela de 500 px de altura, o último item do menu lateral é alcançável por rolagem.
- [ ] **CA-07** — Tocar uma música, navegar por cinco telas diferentes e voltar: a música não foi interrompida.
- [ ] **CA-08** — Rolar uma lista, abrir um item, voltar pelo navegador: a lista está na mesma posição e com os mesmos itens carregados.
- [ ] **CA-09** — Endereço inexistente mostra a página "não encontrada" dentro do shell.
- [ ] **CA-10** — Um erro de renderização provocado em uma tela não derruba o menu; navegar para outra tela funciona.
- [ ] **CA-11** — É possível ir da Home até tocar uma música usando só o teclado.
- [ ] **CA-12** — Trocar o tema em um dispositivo reflete nos outros após recarregar.

## 10. Decisões em aberto

- **D-01** — Itens da barra inferior no celular: os quatro propostos em RF-05 ou outra seleção. *Provisório: RF-05.*
- **D-02** — Ordem e agrupamento do menu (RF-01). Hoje o menu é uma lista única (Home, Arquivos, Favoritos, Imagens, Música, Vídeos, Assistente, Takeout, Capturas, Downloads, Lixeira, Analytics, Configurações, Sobre), sem Documentos, Tarefas, E-mails e Diário; "Downloads" é a tela de aplicativos cliente e não existe tela de download por URL. *Provisório: RF-01.*

## 11. Rastreabilidade

`docs/revisao-produto/02-shell-layout.md` (24 lacunas, todas marcadas como resolvidas — reverificar contra os CA acima).
