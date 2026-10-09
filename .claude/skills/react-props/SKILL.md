---
name: react-props
description: >-
  Ensina props em React: passar dados de pai para filho, ler com destructuring
  {person, size}, valores default (size = 100), encaminhar com spread {...props},
  passar JSX via children, e que props são read-only / snapshots imutáveis. Cobre
  "Passing Props to a Component" de react.dev/learn. Use quando o usuário disser
  "como passo props", "props com valor default", "children no React", "spread de
  props", "posso mudar props", ou "componente não recebe os dados do pai".
when_to_use: >-
  Disparar ao passar/ler props, definir defaults, usar children como wrapper, ou
  quando alguém tenta mutar props. NÃO usar para estado interno mutável
  (react-state-usestate) nem para Context (react-context).
---

# React: Props

Fonte oficial: https://react.dev/learn/passing-props-to-a-component

Props são como componentes se comunicam: todo pai passa informação aos filhos via
props. Funcionam como atributos HTML, mas aceitam **qualquer valor JS** (objetos,
arrays, funções).

## Passar e ler

```jsx
// Pai passa props (como atributos):
export default function Profile() {
  return <Avatar person={{ name: 'Lin', imageId: '1bX5QH6' }} size={100} />;
}

// Filho lê via destructuring do único argumento (props):
function Avatar({ person, size }) {
  return <img src={getImageUrl(person)} alt={person.name} width={size} height={size} />;
}
```

`function Avatar({ person, size })` equivale a `function Avatar(props) { let person = props.person; ... }`.
React passa **um único argumento**: o objeto `props`.

## Valores default

```jsx
function Avatar({ person, size = 100 }) { /* ... */ }
```

O default é usado **só** se a prop estiver **ausente** ou for `undefined`. **Não**
é usado para `size={0}` nem `size={null}`.

## Encaminhar com spread `{...props}`

```jsx
function Profile(props) {
  return <div className="card"><Avatar {...props} /></div>;
}
```

Use com moderação. Se você espalha props em todo componente, provavelmente devia
**dividir os componentes** e passar `children` em vez disso.

## Passar JSX como `children`

Conteúdo aninhado dentro de uma tag chega na prop especial `children`:

```jsx
function Card({ children }) {
  return <div className="card">{children}</div>;
}

<Card>
  <Avatar size={100} person={{ name: 'Katsuko', imageId: 'YfeOqp2' }} />
</Card>
```

`Card` não precisa saber o que tem dentro — ótimo para wrappers visuais (painéis,
grids).

## Props mudam com o tempo, mas são imutáveis

Props são **snapshots read-only no tempo**: a cada render o componente recebe uma
**nova versão** das props. Você **não pode mudar props** — elas são imutáveis.

- Para receber props diferentes, o **pai precisa passar props diferentes**; as
  antigas são descartadas e coletadas pelo GC.
- Para responder a interação do usuário (mudar algo internamente), use **state**
  ([[react-state-usestate]]), não tente mutar props.

## Armadilhas

- Esquecer as chaves no destructuring: `function Avatar(person, size)` (errado) vs
  `function Avatar({ person, size })` (certo).
- Achar que `size={0}` ou `size={null}` aciona o default — não aciona.
- Mutar uma prop (`props.x = ...` ou empurrar num array recebido) → bug; props são read-only.
- Abusar de `{...props}` em vez de modelar `children`.

## Verificação

- Filho lê via `{ destructuring }`? Defaults só cobrem ausente/undefined?
- Nada muta props? Mudança interna usa state?

Relacionados: [[react-components-basics]], [[react-state-usestate]], [[react-jsx]].
