---
name: react-conditional-rendering
description: >-
  Ensina renderização condicional em React: retornar JSX diferente com if,
  retornar null para não renderizar nada, o ternário {cond ? <A/> : <B/>} dentro
  do JSX, o operador lógico {cond && <A/>} (e a armadilha do 0 à esquerda), e
  salvar JSX em variável. Cobre "Conditional Rendering" de react.dev/learn. Use
  quando o usuário disser "renderizar condicionalmente", "mostrar componente só
  se", "ternário no JSX", "&& no JSX", "apareceu um 0 na tela", ou "como escondo
  um elemento".
when_to_use: >-
  Disparar ao mostrar/esconder JSX por condição, ou ao ver "0" renderizado
  inesperadamente. NÃO usar para listas (react-rendering-lists).
---

# React: Renderização condicional

Fonte oficial: https://react.dev/learn/conditional-rendering

Em React, você controla o branching com **JavaScript** (`if`, `? :`, `&&`).

## `if` retornando JSX diferente

```jsx
function Item({ name, isPacked }) {
  if (isPacked) {
    return <li className="item">{name} ✅</li>;
  }
  return <li className="item">{name}</li>;
}
```

## Retornar `null` para não renderizar nada

```jsx
function Item({ name, isPacked }) {
  if (isPacked) return null; // não renderiza nada
  return <li className="item">{name}</li>;
}
```

Na prática, retornar `null` é incomum — costuma ser mais claro incluir/excluir o
componente no JSX **do pai**.

## Ternário `? :` dentro do JSX

```jsx
<li className="item">
  {isPacked ? name + ' ✅' : name}
</li>

// Com JSX em cada ramo:
<li className="item">
  {isPacked ? <del>{name + ' ✅'}</del> : name}
</li>
```

## Lógico `&&`

Renderiza o JSX quando a condição é `true`, senão nada:

```jsx
<li className="item">{name} {isPacked && '✅'}</li>
```

⚠️ **Armadilha do número à esquerda:** se a esquerda do `&&` for `0`, a expressão
vira `0` e o React **renderiza `0`** na tela (não "nada"). Sempre booleanize:

```jsx
messageCount && <p>New messages</p>      // 🔴 renderiza 0 se count === 0
messageCount > 0 && <p>New messages</p>  // ✅
```

## Salvar JSX em variável (mais verboso, mais flexível)

```jsx
function Item({ name, isPacked }) {
  let itemContent = name;
  if (isPacked) {
    itemContent = <del>{name + ' ✅'}</del>;
  }
  return <li className="item">{itemContent}</li>;
}
```

## Resumo

- `{cond ? <A /> : <B />}` = "se cond, renderiza A, senão B".
- `{cond && <A />}` = "se cond, renderiza A, senão nada".
- Booleanize a esquerda do `&&` para não vazar `0`.
- Para lógica complexa, atribua JSX a uma variável com `if`/`let`.

Relacionados: [[react-jsx]], [[react-rendering-lists]].
