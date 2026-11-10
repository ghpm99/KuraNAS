# 06 — Vídeos

**Prioridade:** P0 · **Plataformas:** web, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Assistir aos filmes, séries e animes guardados no servidor, sem anúncio, em qualquer dispositivo da casa, continuando de onde parou e emendando o próximo episódio.

## 2. Cenários de uso

- O dono abre "Séries", escolhe um anime, e a tela mostra em qual episódio parou; "Continuar" retoma no minuto certo.
- Termina um episódio e o próximo começa sozinho, na ordem certa.
- Começa um filme no PC e termina no tablet.
- O arquivo é um MKV com legenda embutida e áudio em codec que o navegador não toca: assiste assim mesmo, com a legenda.

## 3. Escopo

Catálogo (continuar assistindo, séries, filmes, pessoais, clipes, pastas), organização de séries em temporadas e episódios, playlists, player, legendas, progresso e "assistido", conversão sob demanda, miniaturas e pôster local, detalhes do vídeo.

## 4. Fora de escopo

- Qualquer forma de anúncio, trailer automático ou conteúdo promovido.
- Metadados de serviços externos (sinopse, elenco, nota, pôster da internet).
- Streaming adaptativo (HLS/DASH) com troca de qualidade durante a reprodução.
- Transmissão para TV (Chromecast, DLNA, AirPlay). *(ver D-03)*
- Download de vídeos da internet ([11](11-download-por-url.md)) e gravação de streaming ([12](12-capturas.md)).
- Edição ou corte de vídeo.
- Baixar para assistir offline no Android. *(ver D-04)*

## 5. Conceitos

- **Classificação**: série, filme, pessoal (vídeos de câmera/celular), clipe (curtos), outros. Atribuída automaticamente por caminho, nome e duração; o dono pode corrigir, e a correção prevalece.
- **Série**: conjunto de episódios de uma mesma pasta-mãe, organizados em temporadas.
- **Episódio**: vídeo com número de temporada e de episódio, extraídos do nome (`S01E02`, `1x02`, `Ep 02`, `- 02`) ou da ordem natural dentro da pasta.
- **Progresso**: posição em que o dono parou em um vídeo. Global (vale em todos os dispositivos).
- **Assistido**: vídeo cujo progresso passou de 90 % ou que foi marcado à mão.
- **Playlist**: sequência de vídeos. Automática (uma por série/pasta) ou criada pelo dono.

## 6. Regras de negócio

### Catálogo

- **RN-01** — É vídeo todo arquivo com extensão de vídeo reconhecida: MP4, M4V, MKV, AVI, MOV, WebM, WMV, FLV, MPG/MPEG, M2TS, 3GP, VOB, OGV. Todo vídeo indexado aparece em "Todos os vídeos" e em "Pastas", qualquer que seja a classificação.
- **RN-02** — Episódios de uma série são ordenados por temporada e número (ordem natural: episódio 2 antes do 10). Números que fazem parte de resolução, ano ou codec (`1080p`, `2019`, `x264`) não são confundidos com número de episódio.
- **RN-03** — Séries, filmes e demais listas do catálogo são ordenadas por nome; "Adicionados recentemente" e "Todos os vídeos" por data de adição, mais recente primeiro.
- **RN-04** — Pôster: imagem `poster`/`folder`/`cover` na pasta; senão, um quadro do vídeo. Miniaturas são de um quadro a ~10 % da duração, nunca tela preta.
- **RN-05** — Nomes de seções e de playlists automáticas são traduzidos; nada chega em português fixo do servidor.

### Reprodução

- **RN-06** — O servidor entrega cada vídeo da forma mais leve que o dispositivo consegue tocar: original quando compatível; troca de contêiner quando só o contêiner é o problema; conversão de vídeo/áudio quando o codec não é suportado. A escolha é automática.
- **RN-07** — Avançar e retroceder funcionam em qualquer um dos três modos, para qualquer ponto do vídeo.
- **RN-08** — Vídeo que não pode ser reproduzido mostra o motivo e a ação "Baixar original" — nunca tela preta sem explicação.
- **RN-09** — Nenhum anúncio, vinheta ou conteúdo não pedido é inserido antes, durante ou depois de um vídeo.

### Progresso

- **RN-10** — O progresso é salvo por vídeo, a cada 10 s de reprodução, ao pausar, ao sair do player e ao fechar a aba.
- **RN-11** — Abrir um vídeo com progresso entre 2 % e 90 % retoma do ponto salvo, com a opção visível "Começar do início". Abaixo de 2 % ou já assistido, começa do início.
- **RN-12** — Ao passar de 90 % o vídeo vira "assistido" e sai de "Continuar assistindo"; se for episódio de série, o próximo não assistido entra no lugar.
- **RN-13** — O dono pode marcar e desmarcar "assistido" em um vídeo, em uma temporada ou em uma série inteira.
- **RN-14** — "Continuar assistindo" lista todos os vídeos em andamento e o próximo episódio de cada série em andamento, do mais recente para o mais antigo.
- **RN-15** — Progresso e "assistido" são globais; volume, velocidade, legenda escolhida e preferência de reprodução automática são por dispositivo.

### Sequência

- **RN-16** — Ao terminar um vídeo que está em um contexto (série, playlist, pasta), o próximo começa após contagem regressiva de 10 s, que pode ser cancelada. Fora de contexto (filme avulso), a reprodução termina e a tela oferece voltar ao catálogo.
- **RN-17** — "Próximo" e "anterior" seguem o contexto de onde o vídeo foi aberto. No último episódio da temporada, "próximo" vai para o primeiro da temporada seguinte.

### Legendas e áudio

- **RN-18** — Legendas disponíveis: arquivos `.srt`/`.vtt`/`.ass` com o mesmo nome do vídeo (com ou sem sufixo de idioma) na mesma pasta ou em subpasta `Subs`, e faixas embutidas de texto. Legendas de imagem (PGS/VobSub) aparecem na lista como indisponíveis, com explicação.
- **RN-19** — A escolha de legenda (idioma ou desligada) é lembrada e aplicada ao próximo vídeo.
- **RN-20** — Vídeo com mais de uma faixa de áudio permite escolher a faixa; a escolha de idioma é lembrada. *(ver D-02)*

### Playlists

- **RN-21** — Cada série/pasta tem playlist automática, mantida pelo sistema quando arquivos entram e saem. Playlists do dono têm nome obrigatório e ordem definida por ele.
- **RN-22** — Ocultar uma playlist a tira do catálogo sem afetar os arquivos; ela continua acessível em "Ocultas".
- **RN-23** — Vídeo movido ou renomeado mantém progresso, estrela e posição nas playlists do dono.

## 7. Requisitos por tela

Todas as listas seguem [00](00-padroes-transversais.md).

### Catálogo (`/videos`)

- **RF-01** — Abas: Início, Continuar, Séries, Filmes, Pessoais, Clipes, Pastas, Ocultas.
- **RF-02** — **Início**: fileiras de Continuar assistindo, Adicionados recentemente, Séries, Filmes. Cada fileira carrega e falha por conta própria e tem "ver todos".
- **RF-03** — Card de vídeo: miniatura/pôster, título, duração, resolução, barra de progresso, selo de assistido, estrela. Passar o mouse mostra prévia animada. Card de série: pôster, nome, temporadas, episódios assistidos/total.
- **RF-04** — Clique no card de vídeo abre o player; no card de série abre a tela da série.
- **RF-05** — **Todos os vídeos**: lista completa com busca, ordenação (data de adição, nome, duração, tamanho) e filtros (classificação, formato, resolução, duração, assistido/não assistido).
- **RF-06** — **Pastas**: navegação pela hierarquia de pastas que contêm vídeo.

### Tela da série

- **RF-07** — Pôster, nome, total de episódios e progresso; botão principal "Continuar T1E5" (ou "Assistir" / "Rever"); seletor de temporada; lista de episódios com miniatura, número, título, duração, progresso e selo de assistido.
- **RF-08** — Ações: marcar temporada/série como assistida, favoritar, ocultar, corrigir classificação, corrigir número de temporada/episódio de um vídeo.

### Player (`/video/:id`)

- **RF-09** — Controles: tocar/pausar, barra de progresso com tempo e prévia de quadro ao passar o mouse, ±10 s, volume e mudo, velocidade (0,5× a 2×), legendas, faixa de áudio, anterior/próximo, lista de episódios, tela cheia, picture-in-picture.
- **RF-10** — Os controles somem após 3 s sem interação durante a reprodução e reaparecem ao mover o mouse ou tocar; ficam visíveis em tela cheia.
- **RF-11** — Atalhos: espaço/`k` toca-pausa, ←/→ ±10 s, ↑/↓ volume, `m` mudo, `f` tela cheia, `c` legenda, `n`/`p` próximo/anterior, 0–9 pula para a porcentagem.
- **RF-12** — Título do vídeo (e "T1E5 — nome" quando episódio) visível; botão de voltar leva à tela de origem.
- **RF-13** — Painel de detalhes: nome do arquivo, caminho, tamanho, duração, resolução, codecs, faixas de áudio e legenda, Baixar original, Favoritar, Marcar como assistido, Adicionar a playlist.
- **RF-14** — Enquanto o vídeo carrega ou converte, mostra indicador; conversão em andamento é sinalizada.
- **RF-15** — Controles do sistema (teclas de mídia, notificação) funcionam, inclusive em picture-in-picture.

### Playlists

- **RF-16** — Criar, renomear, reordenar, excluir playlist do dono; adicionar vídeo a playlist por menu com busca; ver em quais playlists um vídeo está.

### Android

- **RF-17** — Catálogo com as mesmas seções, paginado; tela de série; player com retomada, próximo/anterior, legendas, tela cheia em paisagem, picture-in-picture, gestos (duplo toque ±10 s) e envio de progresso ao servidor.

## 8. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/video/catalog/home` · `/video/catalog/section/:section` | início e seções |
| GET | `/video/library/files` · `/folders` · `/folders/videos` · `/movies` | biblioteca |
| GET | `/video/library/search` | busca |
| GET | `/video/continue` | continuar assistindo |
| PUT | `/video/progress/:file_id/watched` | marcar assistido |
| POST | `/video/playback/start` · `/next` · `/previous` | sessão de reprodução |
| GET, PUT | `/video/playback/state` | posição atual |
| GET | `/files/video-stream/:id` | original (com Range) |
| GET | `/files/video/stream/:file_id/remux` · `/transcode` | conversão sob demanda |
| GET | `/files/video/metadata/:file_id` | metadados |
| GET | `/files/video/subtitles/:id[/vtt]` · `/embedded/:stream_index/vtt` | legendas |
| GET | `/files/video-thumbnail/:id` · `/files/video-preview/:id` | miniatura e prévia |
| GET, POST | `/video/playlists` | listar, criar |
| GET, PUT, DELETE | `/video/playlists/:id` · GET `/:id/items` · `/:id/poster` | playlist |
| PUT | `/video/playlists/:id/hidden` · `/:id/reorder` | ocultar, reordenar |
| POST, DELETE | `/video/playlists/:id/videos[/:videoId]` | itens |
| GET | `/video/playlists/by-video/:file_id` · `/hidden` · `/unassigned` | consultas |

**A criar** (não existem hoje): corrigir classificação e número de temporada/episódio de um vídeo; marcar temporada/série como assistida; escolha de faixa de áudio.

## 9. Estados e erros

- Biblioteca sem vídeo: estado vazio explicando como adicionar.
- Servidor sem ferramenta de conversão instalada: vídeos compatíveis tocam; os demais mostram o motivo e "Baixar original"; Configurações avisa da pendência.
- Arquivo ausente do disco: o card indica indisponível; abrir explica.
- Falha ao salvar progresso: a reprodução continua; nova tentativa na próxima gravação.

## 10. Critérios de aceite

Além dos CA de [00](00-padroes-transversais.md) em cada seção, em Todos os vídeos, nos episódios de uma série, nas playlists e na busca:

- [ ] **CA-01** — Um MP4 (H.264/AAC) toca com avanço e retrocesso instantâneos.
- [ ] **CA-02** — Um MKV com H.264 toca no navegador, com avanço e retrocesso para qualquer ponto.
- [ ] **CA-03** — Um vídeo em HEVC ou com áudio AC3/DTS toca (convertido), com avanço e retrocesso.
- [ ] **CA-04** — Um arquivo corrompido mostra a mensagem de erro e "Baixar original".
- [ ] **CA-05** — Assistir 10 minutos, fechar a aba, abrir o mesmo vídeo em outro dispositivo: retoma perto dos 10 minutos e oferece "Começar do início".
- [ ] **CA-06** — Assistir até 95 %: o vídeo fica marcado como assistido, sai de "Continuar" e o próximo episódio aparece lá.
- [ ] **CA-07** — Em uma série de 24 episódios nomeados `Ep 1`…`Ep 24`, a ordem é 1, 2, 3 … 24.
- [ ] **CA-08** — Um arquivo `Serie.S02E03.1080p.x264.mkv` é o episódio 3 da temporada 2.
- [ ] **CA-09** — Ao fim de um episódio, a contagem regressiva aparece e o próximo começa; cancelar mantém a tela no vídeo terminado.
- [ ] **CA-10** — No último episódio de uma temporada, "próximo" abre o primeiro da seguinte; no último da série, "próximo" fica desabilitado.
- [ ] **CA-11** — Um vídeo com `video.pt-BR.srt` ao lado mostra a legenda na lista e a exibe sincronizada; a legenda embutida de texto de um MKV também.
- [ ] **CA-12** — Escolher legenda em um episódio a mantém no seguinte; desligar também é lembrado.
- [ ] **CA-13** — Volume e velocidade escolhidos valem no próximo vídeo e após recarregar.
- [ ] **CA-14** — "Continuar assistindo" mostra três vídeos diferentes em andamento ao mesmo tempo.
- [ ] **CA-15** — Marcar uma temporada como assistida marca todos os episódios dela.
- [ ] **CA-16** — "Filmes" e "Séries" com mais de 100 itens rolam até o fim.
- [ ] **CA-17** — Um vídeo solto em uma pasta com nome genérico aparece em "Todos os vídeos" e em "Pastas".
- [ ] **CA-18** — Miniaturas mostram um quadro do vídeo, não tela preta.
- [ ] **CA-19** — Em tela cheia os controles e a legenda aparecem; os atalhos de teclado funcionam.
- [ ] **CA-20** — Do começo ao fim de um episódio, nenhum conteúdo além do próprio vídeo é exibido.
- [ ] **CA-21** — No Android: abrir uma série, continuar o episódio em andamento, pular para o próximo, ligar legenda e girar para paisagem funcionam; o progresso aparece na web.

## 11. Decisões em aberto

- **D-01** — Reprodução automática do próximo: preferência global no servidor ou por dispositivo (lacuna 58.1 do board). *Provisório: por dispositivo (RN-15), com o valor do servidor como padrão inicial.*
- **D-02** — Escolha de faixa de áudio (animes com áudio japonês e português): exige conversão no servidor a cada troca. *Provisório: em escopo (RN-20), prioridade abaixo das demais regras de reprodução.*
- **D-03** — Assistir na TV (Chromecast/DLNA). *Provisório: fora de escopo; a TV usa o navegador.*
- **D-04** — Assistir offline no Android. *Provisório: fora de escopo.*
- **D-05** — Peso de favoritos nas playlists e playlist automática "Favoritos" de vídeo (lacunas 24 e 25 de Favoritos). *Provisório: não criar.*

## 12. Rastreabilidade

`docs/revisao-produto/06-videos.md` (85 lacunas). Requisitos sem item no backlog: tela de série com temporadas (RF-07, RF-08), correção manual de classificação e numeração, faixa de áudio (RN-20), regra de ausência de anúncio (RN-09).
