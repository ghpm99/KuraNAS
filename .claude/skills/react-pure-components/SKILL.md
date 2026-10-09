---
name: react-pure-components
description: >-
  Ensina pureza de componentes em React: mesmo input → mesmo output, sem efeitos
  colaterais durante o render, não mutar variáveis/objetos/props criados fora da
  função (mutação local é ok), onde os side effects pertencem (event handlers,
  useEffect), e o double-render do StrictMode para detectar impureza. Cobre
  "Keeping Components Pure" de react.dev/learn. Use quando o usuário disser
  "componente renderiza diferente a cada vez", "StrictMode chama duas vezes",
  "posso mutar isso no render", "side effect no render", ou comportamento
  imprevisível na renderização.
when_to_use: >-
  Disparar ao ver mutação de variável externa durante render, comportamento
  inconsistente entre renders, ou dúvida sobre StrictMode renderizar 2x. NÃO usar
  para efeitos em si (react-effects).
---

# React: Mantendo componentes puros

Fonte oficial: https://react.dev/learn/keeping-components-pure

React **assume que todo componente é uma função pura**. Uma função pura:
1. **Cuida só do que é dela** — não muda objetos/variáveis que existiam antes dela.
2. **Mesmo input → mesmo output** — dado o mesmo input, sempre retorna o mesmo JSX.

## O que quebra a pureza

```jsx
let guest = 0;
function Cup() {
  guest = guest + 1;        // 🔴 muta variável externa
  return <h2>Tea cup for guest #{guest}</h2>;
}
```

Impuro: modifica `guest` (criado fora) e produz JSX diferente a cada chamada.
**Conserto:** receba via prop.

```jsx
function Cup({ guest }) {
  return <h2>Tea cup for guest #{guest}</h2>;
}
```

## Mutação local é OK

Você **pode** mutar variáveis/objetos que criou **durante** aquele render:

```jsx
function TeaGathering() {
  const cups = [];                       // criado neste render
  for (let i = 1; i <= 12; i++) cups.push(<Cup key={i} guest={i} />); // ✅
  return cups;
}
```

Seguro porque nada de fora conhece `cups` ("local mutation").

## Onde os side effects pertencem

Render deve ser cálculo puro. Efeitos (mudar a tela, animações, requests) vão:
1. **Event handlers** (preferido) — `onClick={() => setData(x)}`.
2. **useEffect** (último recurso) — roda **depois** do render.

## StrictMode detecta impureza

Em desenvolvimento, o **Strict Mode** chama as funções dos componentes **duas
vezes**. Componente puro dá o mesmo resultado nas duas; impuro diverge (guest
#1,#2,#3 vira #2,#4,#6). Sem efeito em produção.

## Por que importa

Pureza habilita: server rendering, pular re-render quando inputs não mudam,
interromper/reiniciar render com segurança, e debugging previsível.

## Armadilhas

- Mutar **props, state ou context** durante o render → use atualização de state.
- Mutar variável/objeto **externo** (módulo, prop recebida, objeto de state) no render.
- Achar que o "render 2x" do StrictMode é bug — é detecção de impureza (só dev).
- Fazer fetch/subscrição/log direto no corpo do componente em vez de handler/effect.

## Verificação

- O componente retorna o mesmo JSX para os mesmos props/state?
- Nenhuma mutação de algo criado fora do render? Efeitos em handler/useEffect?

Relacionados: [[react-props]], [[react-effects]], [[react-handling-events]].
