# 43 — Vídeos: escolher a faixa de áudio

**Tipo:** feature · **Prioridade:** P1 · **Spec:** [06](../especificacao/06-videos.md) `RN-06`, `RN-07`, `RN-20`, `RF-09`, `RF-13`, `D-02`

## Contexto

Animes costumam ter áudio em japonês e em português no mesmo arquivo. Hoje o player toca sempre a faixa padrão do arquivo, e nenhuma rota lista as faixas de áudio: `/video/metadata/:file_id` traz só `audio_codec` (reconferido em 2026-11-24). As rotas reais são `/video/stream/:file_id/remux|transcode` e `/video/metadata/:file_id` (sem o prefixo `/files`). A spec põe a escolha em escopo, com prioridade abaixo das demais regras de reprodução (`06/D-02`).

## Objetivo

Trocar a faixa de áudio durante a reprodução e lembrar o idioma escolhido para o próximo vídeo.

## O que fazer

1. **Backend**: `GET /video/metadata/:file_id/audio-tracks` (ffprobe com cache, no molde das legendas embutidas) lista as faixas; `/video/stream/:file_id/remux` e `/transcode` aceitam `audio_stream_index` (índice no contêiner, `-map 0:<n>`); índice que não é faixa de áudio do arquivo → `400` com mensagem própria.
2. **Web — menu "Faixa de áudio"** no player (`RF-09`), visível só quando o vídeo tem mais de uma faixa: lista idioma e título de cada faixa (de `/video/metadata/:file_id/audio-tracks`), com a atual marcada.
3. **Web — troca**: escolher outra faixa recarrega o vídeo pela rota de remux (ou transcode, quando o codec exige) com `audio_stream_index`, retomando na mesma posição e mantendo legenda, volume e velocidade.
4. **Web — lembrar**: o idioma escolhido fica no dispositivo (`RN-15`); ao abrir outro vídeo que tenha faixa nesse idioma, ela é usada.

## Critérios de aceite

- [x] Em um MKV com áudio em dois idiomas, trocar a faixa muda o áudio e a reprodução continua do mesmo ponto (diferença de até 2 s). *Backend local, `teste-task43/Anime Audio Duplo - Ep 1.mkv` (jpn 300 Hz, por 800 Hz): `audio-tracks` lista os índices 1 e 2; remux com `audio_stream_index=2` tem −23,5 dB acima de 600 Hz contra −45,7 dB do índice 1 (a faixa portuguesa chega); `remux/start?at=15` → 10 e o remux dura 20 s (começa no quadro-chave informado). Na web, o player carrega do quadro-chave e avança até o ponto pedido (teste do hook); o avanço por `currentTime` num navegador real não foi exercitado (sem navegador na máquina).*
- [x] Avançar e retroceder funcionam depois da troca (`06/RN-07`). *Seek depois da troca mantém `audio_stream_index` e usa o início real do remux (testes do hook); `at=10/15/25` → quadro-chave 10/10/20 no backend local.*
- [x] O episódio seguinte começa já no idioma escolhido, quando tem faixa nele. *Idioma guardado no dispositivo (`kuranas.videos.audioPreferences.v1`); próximo vídeo com faixa nesse idioma abre por remux com o índice dela (teste do hook).*
- [x] Vídeo com uma faixa só não mostra o menu. *Teste do `VideoControls` (uma faixa → sem menu; sem backend → sem menu).*
- [x] `audio_stream_index` inválido responde `400` com mensagem traduzida. *Backend local: `remux?audio_stream_index=0` (faixa de vídeo) e `transcode?audio_stream_index=7` → `400 {"error":"O stream de audio informado nao existe neste video"}`.*
- [x] Testes: montagem dos argumentos do ffmpeg com e sem o índice; teste do handler para o `400`; render sem mock do menu.
- [x] `make ci` verde.

## Sub-itens

- [x] **43.1** — o remux (`-ss` antes de `-i` com `-c copy`) começa no quadro-chave anterior ao `start`, mas a web trata o fluxo como se começasse exatamente em `start` (`remuxOffsetSecondsRef`). Com quadros-chave espaçados (MKV de teste com GOP de 10 s: `start=10` entrega os 30 s inteiros; `start=15` entrega a partir de 10 s), a posição exibida e o ponto de retomada erram até um GOP — fura a tolerância de 2 s do primeiro critério e afeta o avançar/retroceder (`06/RN-07`) em geral. Achado na verificação local da task 43; já existia antes dela. Caminho provável: o servidor localiza o quadro-chave ≤ `start` (ffprobe) e informa o início real ao cliente, ou o remux reposiciona com corte preciso.

## Fora de escopo

- Trocar a faixa sem recarregar o fluxo (exigiria streaming adaptativo, fora de escopo em 06 §4).
- Android (entra pelo board de revisão).
