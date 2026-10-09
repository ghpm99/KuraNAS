---
name: react-state-objects-arrays
description: >-
  Ensina a atualizar objetos e arrays no state do React tratando-os como
  imutáveis (read-only): copiar com spread {...obj} / [...arr] em vez de mutar,
  atualizar objetos aninhados copiando cada nível, métodos imutáveis para arrays
  (concat, filter, map, slice) no lugar de push/pop/splice/sort, e Immer para
  casos profundos. Cobre "Updating Objects in State" e "Updating Arrays in State"
  de react.dev/learn. Use quando o usuário disser "mudei o objeto e não
  re-renderiza", "push no state não funciona", "atualizar objeto aninhado",
  "imutabilidade no React", "spread no setState", ou "Immer".
when_to_use: >-
  Disparar ao atualizar objeto/array em state, ao ver mutação direta (obj.x = ..,
  arr.push) que não re-renderiza, ou aninhamento profundo. NÃO usar para batching
  (react-state-updates).
---

# React: Atualizando objetos e arrays no state

Fontes oficiais: react.dev/learn/updating-objects-in-state,
react.dev/learn/updating-arrays-in-state

## Trate o state como imutável

Embora JS permita mutar, o React **conta com imutabilidade** para detectar
mudanças (compara referência `prev === next`), otimizar e suportar snapshots.

```jsx
const [position, setPosition] = useState({ x: 0, y: 0 });
position.x = 5;                   // 🔴 React não detecta
setPosition({ ...position, x: 5 }); // ✅ novo objeto
```

## Objetos: spread para copiar

```jsx
setPerson({ ...person, firstName: e.target.value });
```

Objetos aninhados: copie **cada nível** que você altera (spread é raso):

```jsx
setPerson({
  ...person,
  artwork: { ...person.artwork, city: e.target.value },
});
```

## Arrays: métodos imutáveis

| Operação | 🔴 Muta (evite) | ✅ Imutável |
|---|---|---|
| Adicionar | `push`, `unshift` | `[...arr, novo]`, `arr.concat(novo)` |
| Remover | `pop`, `shift`, `splice` | `arr.filter(...)` |
| Transformar/Substituir | atribuir `arr[i] = ` | `arr.map(...)` |
| Inserir | `splice` | `slice` + spread |
| Ordenar/Inverter | `sort`, `reverse` | copiar primeiro: `[...arr].sort()` |

```jsx
setItems([...items, 'd']);                 // add
setItems(items.filter(i => i !== 'b'));    // remove
setItems(items.map(i => i.toUpperCase())); // transformar
setItems([...items].sort());               // ordenar (cópia antes!)
```

## Mutação local é OK

Mutar objeto/array que você **acabou de criar** é seguro:

```jsx
const next = {};
next.x = e.clientX; next.y = e.clientY;
setPosition(next); // ✅
```

## Immer para aninhamento profundo

`use-immer` deixa você "mutar" um `draft` (Proxy) e gera o objeto imutável:

```jsx
const [person, updatePerson] = useImmer({ artwork: { city: 'Hamburg' } });
updatePerson(draft => { draft.artwork.city = e.target.value; });
```

## Armadilhas

- `arr.push(x); setArr(arr)` → mesma referência, **sem re-render**.
- `obj.prop = x; setObj(obj)` → mesma referência, **sem re-render**.
- Spread copia só o nível de cima — objeto aninhado mutado ainda quebra.
- `sort`/`reverse` mutam o array original → copie com `[...arr]` antes.

## Verificação

- Toda atualização cria **novo** objeto/array (referência nova)?
- Cada nível aninhado alterado foi copiado? Nenhum push/splice/sort no state direto?

Relacionados: [[react-state-usestate]], [[react-state-updates]], [[react-state-structure]].
