# 43 — Vídeos: escolher a faixa de áudio

**Tipo:** feature · **Prioridade:** P1 · **Spec:** [06](../especificacao/06-videos.md) `RN-06`, `RN-07`, `RN-20`, `RF-09`, `RF-13`, `D-02`

## Contexto

Animes costumam ter áudio em japonês e em português no mesmo arquivo. Hoje o player toca sempre a faixa padrão do arquivo; a lista de faixas aparece só nos detalhes. A spec põe a escolha em escopo, com prioridade abaixo das demais regras de reprodução (`06/D-02`).

## Objetivo

Trocar a faixa de áudio durante a reprodução e lembrar o idioma escolhido para o próximo vídeo.

## O que fazer

1. **Backend**: `/files/video/stream/:file_id/remux` e `/transcode` aceitam `audio_stream_index`; o ffmpeg passa a mapear a faixa pedida; índice que não é faixa de áudio do arquivo → `400` com mensagem própria.
2. **Web — menu "Faixa de áudio"** no player (`RF-09`), visível só quando o vídeo tem mais de uma faixa: lista idioma e título de cada faixa (de `/files/video/metadata/:file_id`), com a atual marcada.
3. **Web — troca**: escolher outra faixa recarrega o vídeo pela rota de remux (ou transcode, quando o codec exige) com `audio_stream_index`, retomando na mesma posição e mantendo legenda, volume e velocidade.
4. **Web — lembrar**: o idioma escolhido fica no dispositivo (`RN-15`); ao abrir outro vídeo que tenha faixa nesse idioma, ela é usada.

## Critérios de aceite

- [ ] Em um MKV com áudio em dois idiomas, trocar a faixa muda o áudio e a reprodução continua do mesmo ponto (diferença de até 2 s).
- [ ] Avançar e retroceder funcionam depois da troca (`06/RN-07`).
- [ ] O episódio seguinte começa já no idioma escolhido, quando tem faixa nele.
- [ ] Vídeo com uma faixa só não mostra o menu.
- [ ] `audio_stream_index` inválido responde `400` com mensagem traduzida.
- [ ] Testes: montagem dos argumentos do ffmpeg com e sem o índice; teste do handler para o `400`; render sem mock do menu.
- [ ] `make ci` verde.

## Fora de escopo

- Trocar a faixa sem recarregar o fluxo (exigiria streaming adaptativo, fora de escopo em 06 §4).
- Android (entra pelo board de revisão).
