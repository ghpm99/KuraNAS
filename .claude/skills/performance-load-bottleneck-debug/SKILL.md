---
name: performance-load-bottleneck-debug
description: "Identifica e corrige gargalos de performance: queries N+1, queries lentas sem índice, endpoints lentos, tarefas assíncronas bloqueando a thread principal, paginação ausente, e serializers pesados. Use quando o usuário disser 'está lento', 'timeout na API', 'carregamento demora', 'query lenta', 'N+1', 'o endpoint trava com muitos dados', ou quando o response time estiver alto em produção."
---

# /performance-load-bottleneck-debug

Diagnostica e resolve os gargalos de performance mais comuns em aplicações Django/DRF + banco relacional. Foca em N+1 queries, índices ausentes, serializers pesados e falta de paginação.

## Quando invocar

- "Está lento / timeout"
- "Query N+1 detectada"
- "Endpoint trava com muito dado"
- "Response time alto em produção"
- Antes de adicionar cache — resolver a causa raiz primeiro

## Procedimento

### 1. Localizar o endpoint ou task afetada

Pergunte (se não informado):
- Qual endpoint/view/task está lenta?
- Em produção ou só em desenvolvimento?
- Com quantos registros o problema aparece?

### 2. Detectar N+1 com django-debug-toolbar ou logging

**Logging temporário de queries** (dev):
```python
# settings_local.py
LOGGING['loggers']['django.db.backends'] = {
    'handlers': ['console'], 'level': 'DEBUG', 'propagate': False
}
```

Procure no log por queries idênticas repetidas N vezes — sinal claro de N+1.

**No código**, busque relações não otimizadas:
```bash
grep -n "\.all()\|\.filter(" {view_file}
# verifique se há select_related/prefetch_related
grep -n "select_related\|prefetch_related" {view_file}
```

### 3. Fixes de N+1

```python
# Antes (N+1)
queryset = Transaction.objects.filter(user=user)
# serializer acessa transaction.payment_set.all() para cada item

# Depois
queryset = Transaction.objects.filter(user=user)\
    .select_related('user', 'category')\
    .prefetch_related('payments', 'installments')
```

Regra: `select_related` para FK (1-to-1, FK direta); `prefetch_related` para M2M e reverse FK.

### 4. Queries lentas — verificar índices

```sql
-- PostgreSQL: queries mais lentas
SELECT query, calls, total_time/calls AS avg_ms, rows
FROM pg_stat_statements
ORDER BY avg_ms DESC
LIMIT 10;

-- EXPLAIN ANALYZE na query suspeita
EXPLAIN ANALYZE
SELECT ... FROM transactions WHERE user_id = 42 AND status = 'open';
```

Se o plano mostrar `Seq Scan` numa tabela grande → adicionar índice:
```python
# models.py
class Meta:
    indexes = [
        models.Index(fields=['user', 'status']),
        models.Index(fields=['created_at']),
    ]
```

### 5. Serializers pesados

```python
# Não use SerializerMethodField com queries dentro — move para o queryset
class TransactionSerializer(serializers.ModelSerializer):
    payment_count = serializers.IntegerField(source='payments_count')  # annotate no queryset

# No queryset:
queryset = Transaction.objects.annotate(payments_count=Count('payments'))
```

### 6. Paginação ausente

Se o endpoint retorna lista sem paginação, adicionar imediatamente:
```python
class TransactionListView(generics.ListAPIView):
    pagination_class = PageNumberPagination
    # ou configurar DEFAULT_PAGINATION_CLASS nas settings
```

### 7. Tasks assíncronas bloqueando response

Se um endpoint faz IO pesado (envio de email, processamento CSV, cálculo complexo):
```python
# Mover para Celery task
from myapp.tasks import process_heavy_task
process_heavy_task.delay(obj.id)  # não chamar direto
```

## Armadilhas

- **`prefetch_related` invalida se você filtrar o queryset depois** do prefetch — o filter tem que vir antes ou usar `Prefetch()` com queryset próprio.
- **`select_related` em M2M causa produto cartesiano** — use `prefetch_related` para M2M.
- **Cache não resolve N+1** — cache mascara o problema mas não elimina as queries.
- **`annotate` muda a cardinalidade** quando há JOIN — use `distinct()` se necessário.
- Índices em colunas de baixa cardinalidade (ex.: `status` com 3 valores) são inúteis para tabelas grandes — prefira índice composto.

## Verificação

- Número de queries no endpoint cai para O(1) ou O(joins) — não O(n)
- `EXPLAIN ANALYZE` mostra `Index Scan` em vez de `Seq Scan` nas queries lentas
- Response time < 200ms para listas com paginação ativa
- Nenhum `SerializerMethodField` fazendo query por item
