---
name: react-rendering-lists
description: >-
  Ensina renderizar listas em React: transformar arrays em JSX com map(),
  filtrar com filter(), e a prop key (regras: única entre irmãos, estável, não
  usar índice quando a lista reordena, não gerar com Math.random()). Cobre
  "Rendering Lists" de react.dev/learn. Use quando o usuário disser "renderizar
  uma lista", "map no JSX", "warning de key", "Each child should have a unique
  key prop", "qual key usar", ou "lista não atualiza direito ao reordenar".
when_to_use: >-
  Disparar ao mapear arrays para JSX, ao ver o warning de key faltando, ou ao
  decidir qual valor usar como key. NÃO usar para condicional (react-conditional-rendering).
---

# React: Renderizando listas

Fonte oficial: https://react.dev/learn/rendering-lists

## map() e filter()

```jsx
const listItems = people.map(person =>
  <li key={person.id}>{person.name}</li>
);
return <ul>{listItems}</ul>;

// Filtrar antes de mapear:
const chemists = people.filter(p => p.profession === 'chemist');
```

## Keys

Toda item de lista precisa de uma prop `key` única. A key diz ao React **qual
item do array** cada componente representa, permitindo casá-los corretamente
quando a lista é reordenada, filtrada ou alterada.

### Regras das keys

- **Única entre irmãos** (pode repetir entre arrays diferentes).
- **Estável** — não muda entre renders. **Não** gere com `Math.random()` durante
  o render (cria itens do zero a cada vez, destrói estado e performance).
- **Não use o índice do array como key** quando a lista pode reordenar, filtrar
  ou ter itens inseridos/removidos.

### De onde tirar keys

- Dados do banco: use os **IDs** (já únicos).
- Dados gerados localmente: contador incremental, `crypto.randomUUID()`, ou `uuid`.

### Keys com Fragment

Para múltiplos nós DOM por item, use `<Fragment key={id}>` (forma longa, pois `<>`
não aceita key):

```jsx
import { Fragment } from 'react';
people.map(person =>
  <Fragment key={person.id}>
    <h1>{person.name}</h1>
    <p>{person.bio}</p>
  </Fragment>
);
```

## Armadilhas

- **Warning "Each child in a list should have a unique key prop"** → falta `key`
  no elemento de topo retornado pelo `map`.
- Key no elemento errado: a `key` vai no elemento **mais externo** dentro do `map`,
  não num filho.
- Índice como key + lista reordenável → estado vaza para o item errado (inputs com
  valor trocado, checkbox marcado no item errado).
- Key não é uma prop legível: o componente **não recebe** `key` como prop. Se
  precisar do valor, passe também como outra prop (ex: `id={person.id}`).

## Verificação

- Cada item tem `key` única e estável (ID do dado, não índice/random)?
- Listas que reordenam usam ID real?

Relacionados: [[react-jsx]], [[react-conditional-rendering]], [[react-state-structure]].
