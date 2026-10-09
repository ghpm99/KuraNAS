---
name: react-effect-events
description: >-
  Ensina dependências reativas e Effect Events em React: valores reativos (props,
  state, e tudo derivado deles) devem entrar nas deps do useEffect; separar lógica
  NÃO-reativa com useEffectEvent (ler o último valor de props/state sem re-disparar
  o effect); e como remover dependências corretamente (mover funções/objetos para
  dentro do effect, usar updater, não "enganar" o linter). Cobre "Lifecycle of
  Reactive Effects", "Separating Events from Effects" e "Removing Effect
  Dependencies" de react.dev/learn. Use quando o usuário disser "useEffectEvent",
  "effect reconecta toda hora", "como tiro essa dependência do useEffect", "valor
  reativo vs não reativo", "o linter pede uma dependência que eu não quero", ou
  "effect dispara quando não deveria".
when_to_use: >-
  Disparar ao lutar com o array de deps do useEffect, querer ler o último valor sem
  re-disparar, ou remover deps com segurança. NÃO usar para o básico de useEffect
  (react-effects) nem para effects desnecessários (react-you-might-not-need-effect).
---

# React: Effect Events e dependências reativas

Fontes: react.dev/learn/lifecycle-of-reactive-effects,
react.dev/learn/separating-events-from-effects,
react.dev/learn/removing-effect-dependencies

## Valores reativos

**Valores reativos** = props, state e qualquer coisa calculada a partir deles
durante o render. **Toda** valor reativo que o Effect lê **deve** estar nas deps —
o linter `react-hooks/exhaustive-deps` exige. O Effect re-sincroniza quando
qualquer dep muda.

```jsx
function ChatRoom({ roomId }) {            // roomId é reativo
  const [serverUrl, setServerUrl] = useState('...'); // reativo
  useEffect(() => {
    const c = createConnection(serverUrl, roomId);
    c.connect();
    return () => c.disconnect();
  }, [serverUrl, roomId]); // ambos nas deps
}
```

## Effect Events (useEffectEvent) — lógica NÃO reativa

Às vezes parte da lógica do Effect deve usar o **último** valor de uma prop/state
**sem** re-disparar o Effect. Extraia num **Effect Event** com `useEffectEvent`:

```jsx
import { useEffect, useEffectEvent } from 'react';

function ChatRoom({ roomId, theme }) {
  const onConnected = useEffectEvent(() => {
    showNotification('Connected!', theme); // lê o último theme, NÃO reativo
  });
  useEffect(() => {
    const c = createConnection(roomId);
    c.on('connected', () => onConnected());
    c.connect();
    return () => c.disconnect();
  }, [roomId]); // theme NÃO entra; só roomId re-conecta
}
```

Effect Events: leem sempre o valor mais recente, **não** são reativos, e **não**
entram nas deps. Só chame-os de dentro de Effects (não passe pra outros componentes).

## Removendo dependências (do jeito certo)

Você só remove uma dep **provando** que ela não é reativa — não "enganando" o linter:
- **Funções/objetos criados no render** mudam toda vez → mova **para dentro** do
  Effect (ou para fora do componente se não usam props/state).
- Para atualizar state baseado no anterior, use **updater** `setX(x => ...)` e tire
  `x` das deps.
- Lógica que não deve reagir a um valor → `useEffectEvent`.
- Objeto/função reativa inevitável → memoize com `useMemo`/`useCallback`.

## Armadilhas

- **Desativar o linter** (`// eslint-disable-next-line`) ou omitir deps → bugs com
  valores stale; é a causa nº1 de Effect bugado.
- Pôr uma função definida no componente nas deps → reconecta a cada render (ela é
  recriada). Mova pra dentro do Effect ou memoize.
- Usar `useEffectEvent` para coisas que **deveriam** ser reativas → para de
  re-sincronizar quando precisava.
- `useEffectEvent` ainda é recente/experimental em algumas versões — confirme suporte.

## Verificação

- Todo valor reativo lido está nas deps (sem suprimir o linter)?
- O que não deve re-disparar virou Effect Event? Funções/objetos foram movidos/memoizados?

Relacionados: [[react-effects]], [[react-you-might-not-need-effect]], [[react-performance-hooks]].
