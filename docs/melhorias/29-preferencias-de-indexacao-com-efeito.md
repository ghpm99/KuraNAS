# 29 — Indexação: as preferências da tela têm efeito

**Tipo:** bug · **Prioridade:** P1 · **Spec:** [21](../especificacao/21-indexacao.md) `RN-03`, `RN-14`, `RF-03`, §8, `CA-12`; [23](../especificacao/23-configuracoes-e-sistema.md) `RN-01`, `CA-01`

## Contexto

Configurações mostra três chaves — "Agendar varredura ao iniciar", "Extrair metadados durante o processamento", "Gerar thumbs e previews" — que **nenhum código lê**. As chaves são `scan_on_startup`, `extract_metadata` e `generate_previews` (`backend/internal/api/v1/configuration/model.go:22-24`, `settings.go:40-42`), gravadas em `system_preferences`. Fora do pacote `configuration` não há nenhuma leitura delas; a tela só as exibe (`frontend/src/components/settings/SettingsScreen.tsx:128-168`). Os testes existentes validam só a persistência.

## Objetivo

Cada preferência de indexação visível muda o que o servidor faz; o que não pode ter efeito sai da tela.

## O que fazer

Uma alteração por item.

1. **Leitura das preferências pelo worker**: o pipeline de processamento recebe um leitor das preferências de indexação (interface pequena, implementada por `configuration`), consultado a cada arquivo — a mudança vale na hora, sem reiniciar.
2. **`extract_metadata`**: desligada, a etapa de metadados do tipo não é enfileirada para arquivos novos ou alterados.
3. **`generate_previews`**: desligada, a etapa de miniatura/prévia não é enfileirada.
4. **`index_document_text`** (chave nova, padrão ligada): desligada, a etapa de texto de documento não é enfileirada.
5. **`scan_interval_hours`** (chave nova, 1 a 168, padrão 24): intervalo da varredura periódica; valor fora da faixa → `400` com mensagem própria.
6. **`scan_on_startup` sai da tela** (a varredura ao iniciar é obrigatória — `RN-03`); a chave continua aceita e devolvida pela API, sem efeito, por compatibilidade com os outros clientes.
7. **Web**: a seção Indexação mostra as quatro preferências com efeito, cada uma com o texto do que acontece ao desligar, incluindo que religar não reprocessa o passado (isso é a reindexação — task 30).

## Critérios de aceite

- [x] Teste do pipeline: com `generate_previews` desligada, um arquivo novo é registrado e não gera etapa de miniatura; religada, volta a gerar.
- [x] O mesmo para `extract_metadata` e `index_document_text`.
- [x] Teste: `scan_interval_hours` muda o intervalo usado pelo agendador; 0 e 169 são recusados com `400`.
- [ ] `21/CA-12` — no sistema rodando: desligar "gerar miniaturas" faz um arquivo novo aparecer sem miniatura; religar volta a gerar.
- [x] "Varrer ao iniciar" não aparece mais na tela; `GET /configuration/settings` ainda devolve `scan_on_startup`.
- [x] Seam test do `PUT /configuration/settings` com as chaves novas; render sem mock da seção.
- [x] `make ci` verde.

## Fora de escopo

- Reindexar por escopo (task 30).
- Unificar o modelo de salvar das seções de Configurações (`23/RN-02`) — item 16 do board de revisão.
- Liga/desliga da marcação por IA (fica em Configurações → IA).
