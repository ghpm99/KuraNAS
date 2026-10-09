---
name: table-filter-debounce-optimization
description: "Otimiza a performance de tabelas com filtros: adiciona debounce em inputs de busca, implementa paginação server-side, evita re-renders desnecessários, e elimina chamadas de API excessivas ao digitar. Use quando o usuário disser 'a tabela trava ao filtrar', 'muitas chamadas de API ao digitar', 'filtro lento', 'debounce na busca', 'tabela com muitos dados está lenta', ou 'o servidor sobrecarrega quando filtra'."
---

# /table-filter-debounce-optimization

Aplica otimizações de performance em componentes de tabela com filtros: debounce nos inputs, paginação server-side, e prevenção de re-renders desnecessários.

## Quando invocar

- "Tabela trava / está lenta ao filtrar"
- "Muitas chamadas de API ao digitar no filtro"
- "Adiciona debounce na busca"
- "Servidor sobrecarrega com filtro"

## Procedimento

### 1. Diagnosticar o problema

Pergunte (se não óbvio) ou inspecione:
- É React Web, React Native, ou Flutter?
- O fetch é feito a cada tecla ou só ao confirmar?
- Há paginação? Server-side ou client-side?
- Quantos registros a tabela exibe sem paginação?

### 2. Adicionar debounce no input de busca

**React / React Native (hook customizado):**
```typescript
// hooks/useDebounce.ts
import { useState, useEffect } from 'react';

export function useDebounce<T>(value: T, delay: number = 300): T {
  const [debouncedValue, setDebouncedValue] = useState(value);
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);
  return debouncedValue;
}

// Uso no componente:
const [search, setSearch] = useState('');
const debouncedSearch = useDebounce(search, 300);

useEffect(() => {
  fetchData({ search: debouncedSearch, page: 1 });
}, [debouncedSearch]);
```

**Flutter:**
```dart
Timer? _debounce;
void _onSearchChanged(String query) {
  if (_debounce?.isActive ?? false) _debounce!.cancel();
  _debounce = Timer(const Duration(milliseconds: 300), () {
    _fetchData(search: query, page: 1);
  });
}
```

### 3. Implementar paginação server-side

Se a tabela carrega tudo de uma vez:

```typescript
// Antes: busca tudo
const { data } = useQuery(['items'], () => api.getAll());

// Depois: paginação server-side
const [page, setPage] = useState(1);
const { data } = useQuery(
  ['items', page, debouncedSearch],
  () => api.getList({ page, search: debouncedSearch, pageSize: 20 }),
  { keepPreviousData: true }  // evita flash de loading ao trocar página
);
```

### 4. Prevenir re-renders desnecessários

```typescript
// Memoize colunas estáticas da tabela
const columns = useMemo(() => [
  { key: 'name', label: 'Nome' },
  { key: 'status', label: 'Status' },
], []);  // sem dependências se são estáticas

// Memoize callbacks de ação
const handleRowClick = useCallback((id: string) => {
  navigate(`/items/${id}`);
}, [navigate]);
```

### 5. Cancelar requests anteriores (React Query / SWR)

```typescript
// React Query cancela automaticamente com AbortController quando
// a key muda antes da resposta chegar — confirme que está usando
// ['items', page, search] como query key, não string fixa

// Se usar fetch manual:
useEffect(() => {
  const controller = new AbortController();
  fetch(url, { signal: controller.signal }).then(...);
  return () => controller.abort();
}, [debouncedSearch]);
```

### 6. Reset de página ao filtrar

```typescript
// Bug comum: mudar o filtro mas não resetar para página 1
const handleSearchChange = (value: string) => {
  setSearch(value);
  setPage(1);  // sempre resetar ao filtrar
};
```

## Armadilhas

- **Debounce de 0ms não funciona** — mínimo de 150ms; 300ms é o padrão confortável.
- **`keepPreviousData: true`** evita o flash de loading mas pode mostrar dados velhos durante o load — adicionar indicador visual sutil.
- **Não cancele requests em loop** — se o debounce está curto (< 150ms), requests se acumulam mesmo com cancel.
- **Filtros múltiplos (data + status + busca)** — todos devem usar o mesmo `useEffect` de fetch com todas as dependências, não múltiplos `useEffect` separados.
- **Resetar para página 1** ao qualquer mudança de filtro — esquecer isso faz o usuário ver "nenhum resultado" porque a página 5 não existe com o novo filtro.

## Verificação

- DevTools Network: ao digitar rapidamente no filtro, apenas 1 request é feito (após o debounce)
- Nenhum request pendente é exibido com dados da resposta anterior
- Ao trocar de página, a tabela mostra dados anteriores (blur) até o novo load chegar
- Trocar filtro reseta para página 1 automaticamente
