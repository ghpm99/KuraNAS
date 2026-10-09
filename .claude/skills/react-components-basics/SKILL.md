---
name: react-components-basics
description: >-
  Ensina os fundamentos de componentes React: definir um componente (função com
  nome capitalizado que retorna JSX), aninhar componentes, e import/export
  (default vs named). Cobre as páginas oficiais "Your First Component" e
  "Importing and Exporting Components" de react.dev/learn. Use quando o usuário
  disser "como crio um componente React", "default vs named export", "como
  importo um componente", "estrutura de arquivo de componente", ou estiver
  começando um componente do zero.
when_to_use: >-
  Disparar ao criar o primeiro componente de um arquivo, organizar componentes
  em arquivos separados, ou tirar dúvida sobre export default vs named. NÃO usar
  para props (react-props), JSX em detalhe (react-jsx) ou estado (react-state-usestate).
---

# React: Componentes (fundamentos)

Fonte oficial: https://react.dev/learn/your-first-component e
https://react.dev/learn/importing-and-exporting-components

Componentes são a base do React: elementos de UI reutilizáveis que combinam
markup, CSS e JavaScript. Você compõe, ordena e aninha componentes para montar
páginas inteiras.

## Definindo um componente

Um componente React é uma função JavaScript que retorna markup. Três passos:

```js
// 1. Exporte (export default = ponto de entrada do arquivo)
export default function Profile() {
  // 2. Defina a função (NOME COM LETRA MAIÚSCULA — obrigatório)
  // 3. Retorne JSX
  return (
    <img
      src="https://i.imgur.com/MK3eW3As.jpg"
      alt="Katherine Johnson"
    />
  );
}
```

Regras inegociáveis:
- **O nome DEVE começar com letra maiúscula.** `function profile()` não funciona —
  o React trata tags minúsculas como tags HTML e maiúsculas como componentes.
- Retorna **JSX**. Se o markup não estiver na mesma linha do `return`, **envolva em
  parênteses** — sem eles, tudo após o `return` é ignorado.

## Usando / aninhando componentes

Componentes podem renderizar outros componentes:

```js
function Profile() {
  return <img src="https://i.imgur.com/MK3eW3As.jpg" alt="Katherine Johnson" />;
}

export default function Gallery() {
  return (
    <section>
      <h1>Amazing scientists</h1>
      <Profile />
      <Profile />
      <Profile />
    </section>
  );
}
```

`Gallery` é o **parent**; cada `<Profile />` é um **child**. Tags minúsculas
(`<section>`) viram HTML; tags capitalizadas (`<Profile />`) viram componentes.

## Import / Export

Um arquivo tem **no máximo um `export default`**, mas **quantos named exports
quiser**.

| Sintaxe | Export | Import |
|---|---|---|
| Default | `export default function Button() {}` | `import Button from './Button.js';` |
| Named | `export function Button() {}` | `import { Button } from './Button.js';` |

- **Default import**: pode usar qualquer nome (`import Banana from './Button.js'` funciona).
- **Named import**: o nome **tem que bater** dos dois lados (por isso "named").

Convenção: use `default` quando o arquivo exporta só um componente; use `named`
quando exporta vários. `'./Gallery'` e `'./Gallery.js'` ambos funcionam.

## Armadilhas

- **Nunca defina um componente dentro de outro.** É lento e causa bugs (o filho
  é recriado a cada render, perdendo estado):
  ```js
  export default function Gallery() {
    function Profile() {} // 🔴 NUNCA
  }
  ```
  Defina sempre no **top level** do módulo. Se o filho precisa de dados do pai,
  passe via **props** ([[react-props]]), não por aninhamento de definição.
- Esquecer os parênteses após `return` em markup multilinha → retorna `undefined`.
- Esquecer a maiúscula no nome → React renderiza como tag HTML desconhecida, sem erro óbvio.
- Componente anônimo (`export default () => {}`) dificulta debug — prefira nomear.

## Verificação

- Nome capitalizado? Retorna JSX (com parênteses se multilinha)? Definido no top level?
- `import`/`export` casam (named bate o nome; default pode renomear)?

Próximos: [[react-jsx]], [[react-props]], [[react-pure-components]].
