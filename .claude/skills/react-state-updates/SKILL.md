---
name: react-state-updates
description: >-
  Ensina como o React processa atualizações de state: state é um snapshot fixo
  dentro de um render, o batching (várias chamadas setX(x+1) no mesmo handler só
  contam uma vez), e updater functions setX(n => n+1) para enfileirar múltiplas
  atualizações; mistura de replace e updater. Cobre "Render and Commit", "State
  as a Snapshot" e "Queueing a Series of State Updates" de react.dev/learn. Use
  quando o usuário disser "setState não atualiza na hora", "chamei setX 3x e só
  somou 1", "valor antigo depois do setState", "updater function", "n => n+1", ou
  "batching".
when_to_use: >-
  Disparar quando múltiplos setX no mesmo handler não acumulam, quando o valor
  lido logo após setX é o antigo, ou ao precisar enfileirar updates. NÃO usar para
  imutabilidade de objetos/arrays (react-state-objects-arrays).
---

# React: Atualizações de state (snapshot, batching, updater)

Fontes oficiais: react.dev/learn/render-and-commit,
react.dev/learn/state-as-a-snapshot,
react.dev/learn/queueing-a-series-of-state-updates

## State é um snapshot

Definir state **pede um novo render**, mas **não muda a variável no render atual**.
Dentro de um handler, o valor de state é **fixo** (snapshot daquele render):

```jsx
function handleClick() {
  setNumber(number + 1);
  setNumber(number + 1);
  setNumber(number + 1);
} // number era 0 → as três viram setNumber(0 + 1) → resultado: 1, não 3
```

## Batching

React **agrupa** (batches) as atualizações: espera todo o handler terminar antes
de re-renderizar. Evita renders pela metade. O batching é **dentro** de um
handler — cada clique/evento separado é processado por si.

## Updater functions (para acumular)

Passe uma **função** que recebe o state pendente e retorna o próximo:

```jsx
function handleClick() {
  setNumber(n => n + 1);
  setNumber(n => n + 1);
  setNumber(n => n + 1);
} // fila: 0→1→2→3 → resultado: 3 ✅
```

| Fila | Recebe | Retorna |
|---|---|---|
| `n => n + 1` | 0 | 1 |
| `n => n + 1` | 1 | 2 |
| `n => n + 1` | 2 | 3 |

Updater **deve ser puro** (só retorna o próximo valor; sem side effects).

## Misturando replace e updater

```jsx
setNumber(number + 5); // "replace por 5"
setNumber(n => n + 1); // aplica sobre 5 → 6
setNumber(42);         // "replace por 42" (descarta o resto) → 42
```

Um valor cru substitui a fila; um updater opera sobre o resultado pendente.

## Render and Commit (visão geral)

Renderizar é: (1) **trigger** (render inicial ou setState), (2) **render** —
React chama seus componentes (deve ser puro), (3) **commit** — React aplica as
mudanças no DOM. O browser então repinta.

## Armadilhas

- Ler `number` logo após `setNumber(...)` e esperar o novo valor — ainda é o snapshot antigo.
- Vários `setX(x+1)` esperando somar — use `setX(n => n+1)`.
- Side effect dentro do updater (ex: `setItems(items => { fetch(); return ... })`) — proibido.

## Verificação

- Atualizações que dependem do valor anterior usam updater `n => ...`?
- Nenhum código assume que o state mudou na mesma execução do handler?

Relacionados: [[react-state-usestate]], [[react-state-objects-arrays]].
