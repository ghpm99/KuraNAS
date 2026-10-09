---
name: react-state-structure
description: >-
  Ensina a estruturar bem o state em React: agrupar state relacionado, evitar
  contradições (use um status em vez de vários booleans), evitar state redundante
  (calcule derivados no render), não espelhar props no state, evitar duplicação
  (guarde IDs e derive o objeto), e achatar/normalizar state aninhado. Cobre
  "Choosing the State Structure" e "Reacting to Input with State" de
  react.dev/learn. Use quando o usuário disser "como organizo o state", "tenho
  vários booleans de status", "state derivado", "preciso guardar isso no state?",
  "props no state não atualiza", "state aninhado difícil de atualizar", ou
  "estados impossíveis".
when_to_use: >-
  Disparar ao modelar/refatorar a forma do state, ou ao ver booleans contraditórios,
  valores derivados guardados, props copiadas pra state, ou objetos duplicados.
  NÃO usar para lift state up (react-sharing-state).
---

# React: Estruturando o state

Fontes: react.dev/learn/choosing-the-state-structure,
react.dev/learn/reacting-to-input-with-state

Pense de forma **declarativa**: enumere os *estados visuais* da UI (vazio,
digitando, enviando, sucesso, erro) e dispare transições — não manipule a tela
imperativamente.

## Princípios

1. **Agrupe state relacionado.** Se sempre atualiza junto, una num objeto:
   `useState({ x: 0, y: 0 })` em vez de dois `useState`.
2. **Evite contradições.** Não deixe `isSending`+`isSent` poderem ser ambos true.
   Use um `status`: `'typing' | 'sending' | 'sent'`, e derive `isSending = status === 'sending'`.
3. **Evite state redundante.** Não guarde o que dá pra **calcular no render**:
   `const fullName = firstName + ' ' + lastName;` (não um `useState` separado).
4. **Não espelhe props no state.** `useState(messageColor)` "congela" a prop. Use a
   prop direto (`const color = messageColor`). Exceção: valor **inicial** intencional
   (`initialColor`).
5. **Evite duplicação.** Guarde a peça essencial (ex: `selectedId`) e **derive** o
   objeto (`items.find(i => i.id === selectedId)`), em vez de copiar o objeto inteiro.
6. **Evite aninhamento profundo.** Achate/normalize como um banco (mapa por id +
   `childIds`) para facilitar updates.

> "Make your state as simple as it can be — but no simpler."

## Armadilhas

- Guardar `fullName`/total/filtrado no state → vira fonte de bugs de sincronia. Derive.
- `useState(prop)` esperando atualizar quando a prop muda — não atualiza (só roda no 1º render).
- Selecionar guardando o **objeto** em vez do **id** → fica stale quando a lista muda.
- Booleans demais que criam combinações "impossíveis" → troque por um enum `status`.

## Verificação

- Algum valor do state poderia ser **calculado** a partir de outro state/props? Remova.
- Existe combinação de state que nunca deveria acontecer? Modele com `status`.
- State aninhado fundo? Considere normalizar.

Relacionados: [[react-state-usestate]], [[react-sharing-state]], [[react-state-objects-arrays]].
