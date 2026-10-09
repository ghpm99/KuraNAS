---
name: react-sharing-state
description: >-
  Ensina compartilhar state entre componentes em React via "lifting state up"
  (mover o state pro pai comum mais próximo e passar por props + handlers),
  componentes controlled vs uncontrolled, single source of truth; e preservar/
  resetar state (React guarda state por POSIÇÃO na árvore — mesma posição preserva,
  posição/tipo diferente reseta; resetar com key). Cobre "Sharing State Between
  Components" e "Preserving and Resetting State" de react.dev/learn. Use quando o
  usuário disser "dois componentes precisam do mesmo state", "lift state up",
  "controlled component", "o state some/persiste quando troco de aba", "resetar
  state com key", ou "componente mantém valor antigo ao trocar".
when_to_use: >-
  Disparar ao sincronizar state entre irmãos, decidir controlled/uncontrolled, ou
  quando state persiste/reseta inesperadamente ao reordenar/trocar componentes.
  NÃO usar para Context (react-context) nem reducer (react-reducer).
---

# React: Compartilhando / preservando state

Fontes: react.dev/learn/sharing-state-between-components,
react.dev/learn/preserving-and-resetting-state

## Lifting state up

Para dois componentes compartilharem state, mova-o para o **pai comum mais
próximo** e passe para baixo via props; passe **handlers** para os filhos
mudarem o state do pai.

```jsx
export default function Accordion() {
  const [activeIndex, setActiveIndex] = useState(0);
  return (
    <>
      <Panel isActive={activeIndex === 0} onShow={() => setActiveIndex(0)} />
      <Panel isActive={activeIndex === 1} onShow={() => setActiveIndex(1)} />
    </>
  );
}
function Panel({ isActive, onShow, title, children }) {
  return <section>{isActive ? <p>{children}</p> : <button onClick={onShow}>Show</button>}</section>;
}
```

**Single source of truth:** cada peça de state tem **um** componente dono. Não
duplique state compartilhado.

- **Uncontrolled**: tem state local próprio; o pai não controla.
- **Controlled**: dirigido por props do pai (mais flexível p/ coordenar).

## Preservar e resetar state

React **guarda o state pela POSIÇÃO** na árvore de UI, não pela identidade JSX:

- **Mesmo componente, mesma posição** → state **preservado** (mesmo que props mudem).
- **Posição diferente** ou **tipo de componente diferente** naquela posição → state **resetado**.
- Renderizar `<Counter/>` vs `null` na mesma posição destrói o state quando vira null.

Para **forçar reset** de um componente na mesma posição, dê uma **`key`** diferente:

```jsx
{isPlayerA
  ? <Counter key="Taylor" person="Taylor" />
  : <Counter key="Sarah" person="Sarah" />}
```

key diferente → React trata como componente distinto → state reinicia.

## Armadilhas

- Definir um componente **dentro** de outro → ele troca de "identidade" a cada render
  do pai e **perde o state**. Defina no top level.
- Renderização condicional que troca o **tipo** na mesma posição reseta o state sem querer.
- Usar índice de lista como key + reordenar → state vaza pro item errado ([[react-rendering-lists]]).
- Esquecer de passar o handler pra baixo → filho não consegue atualizar o pai.

## Verificação

- O state vive no pai comum mais próximo que precisa dele? Uma única fonte da verdade?
- Componente certo preserva/reseta? Usou `key` quando quis reiniciar?

Relacionados: [[react-state-structure]], [[react-context]], [[react-props]].
