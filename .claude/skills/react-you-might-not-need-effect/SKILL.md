---
name: react-you-might-not-need-effect
description: >-
  Ensina quando NÃO usar useEffect em React: calcular dados derivados no render (não
  num effect), cachear cálculo caro com useMemo, resetar state via key, ajustar
  state no render (raro), pôr lógica de evento em handlers (POST de form no handler,
  não no effect), evitar cadeias de effects, init de app no nível do módulo, e usar
  useSyncExternalStore para stores externas. A regra: effects são para sincronizar
  com sistemas externos, não para transformar dados nem tratar eventos. Cobre "You
  Might Not Need an Effect" de react.dev/learn. Use quando o usuário disser "effect
  sincronizando state com state", "setState dentro de useEffect", "effect roda
  demais/loop", "preciso mesmo desse useEffect", "cadeia de effects", ou "render
  extra".
when_to_use: >-
  Disparar ao ver useEffect que só transforma state/props, atualiza state de outro
  state, ou reage a um clique. NÃO usar quando o effect realmente sincroniza com
  sistema externo (react-effects).
---

# React: Você talvez não precise de um Effect

Fonte: react.dev/learn/you-might-not-need-an-effect

Effects são escape hatch para **sincronizar com sistemas externos**. Sem sistema
externo, normalmente **não precisa de Effect**. Pergunta-chave:

> Essa lógica roda porque o componente foi **exibido**, ou por uma **interação
> específica**? Exibido → Effect. Interação → event handler.

## Casos para remover o Effect

- **Dados derivados** → calcule **no render**, não com effect+state:
  ```jsx
  const fullName = firstName + ' ' + lastName; // não useEffect(setFullName)
  ```
- **Cálculo caro** → `useMemo`, não effect:
  ```jsx
  const visibleTodos = useMemo(() => getFilteredTodos(todos, filter), [todos, filter]);
  ```
- **Resetar state quando prop muda** → use `key`:
  ```jsx
  <Profile userId={userId} key={userId} /> // reseta tudo no change
  ```
- **Ajustar state no change de prop** (raro) → durante o render:
  ```jsx
  const [prevItems, setPrevItems] = useState(items);
  if (items !== prevItems) { setPrevItems(items); setSelection(null); }
  ```
  Melhor ainda: derive (`items.find(i => i.id === selectedId) ?? null`).
- **Lógica de evento** (POST de submit, notificação de "adicionado ao carrinho") →
  no **event handler**, não no effect. Só analytics de "visitou" fica no effect.
- **Cadeias de effects** (um effect setando state que dispara outro) → calcule tudo
  num único handler.
- **Notificar/passar dado ao pai** → faça o pai buscar e passar pra baixo.
- **Init da app** (uma vez) → código no nível do módulo, fora do componente.
- **Assinar store externa** → `useSyncExternalStore`, não effect.

## Armadilhas

- `useEffect(() => setX(deriveDe(y)), [y])` → render extra; derive no corpo.
- Mostrar notificação dentro de effect que depende de uma prop → dispara em renders
  que você não quis; mova para o handler.
- Cadeia de effects → cascata de re-renders e bugs difíceis.
- Fetch é caso válido de effect (com cleanup `ignore`), mas considere libs.

## Verificação

- Cada `useEffect` sincroniza com algo **externo**? Se não, dá pra calcular no
  render / mover pro handler / usar key / useMemo?

Relacionados: [[react-effects]], [[react-state-structure]], [[react-handling-events]], [[react-performance-hooks]].
