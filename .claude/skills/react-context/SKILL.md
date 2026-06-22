---
name: react-context
description: >-
  Ensina Context em React para passar dados pela árvore sem prop drilling: 3
  passos (createContext, useContext no filho, prover com <Context value=>),
  Context atravessa componentes intermediários, providers aninhados sobrescrevem,
  e quando NÃO usar (preferir props ou passar JSX como children antes). Cobre
  "Passing Data Deeply with Context" de react.dev/learn. Use quando o usuário
  disser "prop drilling", "useContext", "createContext", "passar dado fundo na
  árvore", "Context Provider", "tema/usuário logado global", ou "evitar passar
  props por vários níveis".
when_to_use: >-
  Disparar ao introduzir Context, sofrer prop drilling, ou decidir entre context e
  props/children. NÃO usar para a lógica de state em si (react-reducer) nem para o
  padrão reducer+context combinado (react-reducer-context).
---

# React: Context

Fonte: react.dev/learn/passing-data-deeply-with-context

**Prop drilling**: passar props por muitos níveis intermediários só para chegar
no filho fundo. Context deixa um pai disponibilizar um valor para **toda a árvore
abaixo**, sem passar prop a prop.

## 3 passos

```jsx
// 1. Criar (com valor default):
// LevelContext.js
import { createContext } from 'react';
export const LevelContext = createContext(1);

// 2. Usar no filho (no topo do componente, nunca em loop/condição):
import { useContext } from 'react';
const level = useContext(LevelContext);

// 3. Prover a partir do pai:
<LevelContext value={level}>
  {children}
</LevelContext>
```

> Nota de versão: React 19 permite usar `<LevelContext value=>` direto como
> provider. Em versões anteriores use `<LevelContext.Provider value=>`.

O filho lê o valor do **provider mais próximo acima** dele. Um componente pode
**usar e prover** o mesmo context (ex: cada `Section` faz `value={level + 1}`).

## Atravessa intermediários; aninhados sobrescrevem

Componentes intermediários **não precisam saber** do context — ele passa por eles
automaticamente. Providers aninhados do mesmo context sobrescrevem o de cima.
Contexts diferentes são independentes.

## Antes de usar Context, considere alternativas

1. **Passar props explicitamente** — fluxo de dados claro; vale a verbosidade.
2. **Extrair componente e passar JSX como `children`** — se o intermediário não
   usa o dado, talvez falte uma extração: `<Layout><Posts posts={posts} /></Layout>`.

Use context só quando nenhuma das duas serve bem.

Bons casos: tema (dark mode), usuário logado, routing, e state complexo via
[[react-reducer-context]]. Funciona melhor com valores que mudam **pouco**.

## Armadilhas

- `useContext` em loop/condição → Rules of Hooks ([[react-state-usestate]]).
- Esquecer o `<...value=>` provider → o filho recebe o **valor default** silenciosamente.
- Recriar o objeto `value` a cada render → re-renderiza todos os consumidores
  (memoize o value quando custar caro).
- Usar context onde props/children resolveriam → acopla demais e dificulta o rastreio.

## Verificação

- Há provider acima de todo consumidor? `useContext` no topo?
- Realmente precisa de context, ou props/children bastam?

Relacionados: [[react-sharing-state]], [[react-reducer]], [[react-reducer-context]].
