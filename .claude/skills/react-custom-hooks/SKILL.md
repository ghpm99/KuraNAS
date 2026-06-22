---
name: react-custom-hooks
description: >-
  Ensina custom Hooks em React: extrair e reutilizar lógica com estado entre
  componentes, a convenção de nome use* (obrigatória para o React reconhecer e o
  linter aplicar as regras), que custom Hooks compartilham a LÓGICA e não o state
  (cada chamada tem state isolado), passar valores reativos/handlers (com
  useEffectEvent), e quando extrair. Cobre "Reusing Logic with Custom Hooks" de
  react.dev/learn. Use quando o usuário disser "custom hook", "useOnlineStatus",
  "reusar lógica entre componentes", "extrair um hook", "useFormInput", "criar meu
  próprio hook", ou "lógica de effect repetida".
when_to_use: >-
  Disparar ao extrair lógica com hooks repetida em vários componentes, ou ao criar
  um use* próprio. NÃO usar para compartilhar o state em si (react-sharing-state).
---

# React: Custom Hooks

Fonte: react.dev/learn/reusing-logic-with-custom-hooks

Um **custom Hook** é uma função que chama outros Hooks, para extrair e reutilizar
**lógica com estado** entre componentes.

```jsx
function useOnlineStatus() {
  const [isOnline, setIsOnline] = useState(true);
  useEffect(() => {
    const on = () => setIsOnline(true);
    const off = () => setIsOnline(false);
    window.addEventListener('online', on);
    window.addEventListener('offline', off);
    return () => {
      window.removeEventListener('online', on);
      window.removeEventListener('offline', off);
    };
  }, []);
  return isOnline;
}

function StatusBar() {
  const isOnline = useOnlineStatus(); // reusa em qualquer componente
  return <h1>{isOnline ? '✅ Online' : '❌ Disconnected'}</h1>;
}
```

## Convenção de nome: `use` + Maiúscula

Nomes de Hook **devem** começar com `use` (ex: `useFormInput`). Isso (1) sinaliza
que a função contém Hooks e segue as regras, e (2) deixa o linter aplicar as Rules
of Hooks. Função que **não** chama Hooks **não** deve usar `use`.

Evite Hooks "de ciclo de vida" como `useMount`/`useEffectOnce` — não combinam com
o paradigma. Extraia Hooks de **alto nível** com propósito claro.

## Compartilha LÓGICA, não state

Cada chamada de um custom Hook tem **state isolado**:

```jsx
const firstNameProps = useFormInput('Mary');   // state próprio
const lastNameProps  = useFormInput('Poppins'); // state separado
```

Equivale a dois `useState` independentes. Para compartilhar o **mesmo** state entre
componentes, faça lift up e passe via props ([[react-sharing-state]]).

## Valores reativos e handlers

Custom Hooks re-rodam a cada render do componente, então recebem sempre os últimos
props/state. Passe valores reativos como argumentos (entram nas deps internas). Para
event handlers passados ao Hook, envolva com `useEffectEvent` para não re-sincronizar
à toa ([[react-effect-events]]).

## Quando extrair

- Lógica usada em **vários lugares**.
- Ao escrever um Effect — às vezes envolvê-lo num Hook clareia a intenção.
- Lógica complexa ou ligada a sistema externo → esconde detalhes, componente fica legível.

Não extraia para duplicação trivial ou utilitários simples que não chamam Hooks.

## Armadilhas

- Nome sem `use` → o React/linter não trata como Hook (não pode chamar useState dentro).
- Esperar que duas chamadas compartilhem state — não compartilham (state é isolado).
- Chamar o custom Hook em condição/loop → quebra Rules of Hooks ([[react-state-usestate]]).
- Criar `useMount`/`useEffectOnce` genéricos em vez de um Hook com propósito.

## Verificação

- Nome começa com `use`? Só chamado no topo, incondicionalmente?
- Está compartilhando lógica (ok) e não esperando compartilhar state?

Relacionados: [[react-effects]], [[react-effect-events]], [[react-sharing-state]], [[react-utility-hooks]].
