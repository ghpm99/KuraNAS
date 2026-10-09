---
name: react-jsx
description: >-
  Ensina JSX em React: as 3 regras (um único elemento raiz / Fragment <>, fechar
  todas as tags, camelCase nos atributos como className) e como embutir
  JavaScript com chaves {} — variáveis, expressões, e {{ }} para objetos/inline
  style. Cobre "Writing Markup with JSX" e "JavaScript in JSX with Curly Braces"
  de react.dev/learn. Use quando o usuário disser "como escrevo JSX", "className
  vs class", "erro de JSX", "Fragment", "como uso variável no JSX", "style inline
  no React", ou "double curly braces".
when_to_use: >-
  Disparar ao escrever markup JSX, ver erro de JSX (tag não fechada, múltiplos
  root elements, atributo HTML cru), ou interpolar JS no markup. NÃO usar para
  passar dados entre componentes (react-props).
---

# React: JSX

Fonte oficial: https://react.dev/learn/writing-markup-with-jsx e
https://react.dev/learn/javascript-in-jsx-with-curly-braces

JSX é uma extensão de sintaxe que deixa você escrever markup parecido com HTML
dentro do JavaScript. Mantém a lógica de renderização junto do markup (eles são
relacionados). JSX é mais **estrito** que HTML.

## As 3 regras do JSX

1. **Retorne um único elemento raiz.** Para vários elementos, envolva num pai
   único — um `<div>` ou um **Fragment** `<>...</>` (não adiciona nó ao DOM):
   ```jsx
   return (
     <>
       <h1>Todos</h1>
       <img src="..." alt="..." />
     </>
   );
   ```
   Motivo: JSX vira objetos JS; não dá pra retornar dois objetos sem envolvê-los.

2. **Feche todas as tags.** `<img>` → `<img />`; `<li>item` → `<li>item</li>`.

3. **camelCase na maioria dos atributos.** Atributos viram chaves de objeto JS,
   então `class` → `className`, `stroke-width` → `strokeWidth`, `onclick` → `onClick`.
   **Exceção:** `aria-*` e `data-*` mantêm os hífens.

Dica: para migrar HTML grande, use https://transform.tools/html-to-jsx.

## JavaScript com chaves `{}`

Chaves embutem **expressões JS** no JSX. Funcionam em **exatamente dois lugares**:

```jsx
const name = 'Gregorio';
const avatar = 'https://.../img.jpg';

// 1. Como texto dentro de uma tag:
<h1>{name}'s To Do List</h1>
<h1>To Do List for {formatDate(today)}</h1>   // chamadas de função também

// 2. Como atributo logo após o =:
<img src={avatar} alt={name} />
```

⚠️ `src="{avatar}"` (com aspas) passa a **string literal** `"{avatar}"`, não o
valor da variável. Aspas = string; chaves = valor dinâmico.

## Double curlies `{{ }}` — objetos e inline style

`{{ }}` **não é sintaxe especial**: é um objeto JS `{...}` dentro das chaves `{}`
do JSX. Usado tipicamente para `style` inline:

```jsx
<ul style={{ backgroundColor: 'black', color: 'pink' }}>
  <li>...</li>
</ul>
```

Propriedades CSS em **camelCase** (`background-color` → `backgroundColor`). Você
também pode passar um objeto vindo de uma variável: `<div style={person.theme}>`.

## Armadilhas

- **Múltiplos elementos raiz** sem wrapper → erro. Use `<>...</>`.
- `class=` em vez de `className=` → React avisa no console.
- Aspas em volta de chaves (`src="{x}"`) → passa string literal, não o valor.
- Esquecer `/` em self-closing (`<img>`) → erro de parse.
- `style="color: red"` (string CSS do HTML) não funciona — `style` recebe **objeto**.

## Verificação

- Um root único? Todas as tags fechadas? Atributos em camelCase?
- Interpolações com `{}` (sem aspas) nos dois lugares válidos?

Relacionados: [[react-components-basics]], [[react-conditional-rendering]], [[react-rendering-lists]].
