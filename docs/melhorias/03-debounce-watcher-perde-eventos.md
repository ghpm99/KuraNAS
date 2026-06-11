# 03 — Debounce do watcher descarta mudanças silenciosamente

**Tipo:** bug · **Prioridade:** P1

## Contexto

O watcher de entry point (`internal/worker/engine/watcher.go`, `startEntryPointWatcher`) tira um snapshot da árvore a cada 5s, calcula o diff contra o snapshot anterior e despacha jobs. Há um debounce de 2s para não despachar em rajada:

```go
currentSnapshot := collectEntryPointSnapshot(entryPoint)
changed := snapshotDiffPaths(lastSnapshot, currentSnapshot)
lastSnapshot = currentSnapshot            // ← snapshot avança AQUI

if len(changed) == 0 { continue }
if !lastDispatchAt.IsZero() && time.Since(lastDispatchAt) < debounceWindow {
    continue                              // ← ...e as mudanças são jogadas fora
}
```

O `lastSnapshot` é avançado **antes** do check de debounce. Quando o tick cai dentro da janela de 2s após um dispatch, as mudanças daquele tick são descartadas e **nunca mais serão vistas** — o snapshot seguinte já as considera estado normal. Cópias em lote (encher uma pasta de músicas, por exemplo) são o cenário típico: o primeiro tick despacha, o segundo cai no debounce e perde arquivos, que só serão indexados num eventual rescan completo.

## Objetivo

Nenhuma mudança detectada pelo watcher é perdida: o debounce pode **adiar** o despacho, nunca descartá-lo.

## O que fazer

Acumular as mudanças detectadas durante a janela de debounce e despachá-las no primeiro tick em que o despacho for permitido.

## Como fazer

- Manter um conjunto `pending map[string]struct{}` no loop do watcher. A cada tick, fazer merge de `changed` em `pending`. Quando o debounce permitir despachar, enviar o conteúdo de `pending` (com o snapshot corrente para resolver criado × deletado) e só então zerar o conjunto.
- Alternativa mais simples (também aceitável): só avançar `lastSnapshot` quando o despacho de fato acontecer. Tem o custo de recalcular o diff acumulado a cada tick, mas elimina o estado extra. Escolher uma das duas — a primeira é mais eficiente para janelas longas.
- Atenção à interação com `watcherMaxIndividualJobs` (50): o conjunto acumulado pode crescer durante a janela e estourar o limite — o fallback para `enqueueFilesystemEventJob` (scan completo) já cobre isso, manter.
- Atenção ao caso "mudou e voltou" dentro da janela (arquivo criado e apagado): ao despachar, resolver cada path contra o snapshot **corrente**, não contra o estado do tick em que a mudança foi vista — a lógica atual de `existsInCurrent` já faz isso, basta preservá-la.
- **Testes**: `watcher_test.go` já cobre o dispatch; adicionar caso simulando dois lotes de mudanças com menos de 2s entre eles e verificando que o segundo lote gera jobs.

## Critérios de aceite

- [ ] Teste de unidade: mudanças que chegam dentro da janela de debounce são despachadas no tick seguinte permitido (nenhum path some).
- [ ] Teste de unidade: arquivo criado e removido dentro da janela não gera job de persist órfão (resolve contra o snapshot corrente).
- [ ] Copiar um lote grande de arquivos em duas levas com <2s de intervalo resulta em todos os arquivos indexados sem rescan manual.
- [ ] `make ci-backend` verde (cobertura ≥ 80%).

## Fora de escopo

- Substituir o mecanismo de polling por eventos nativos (task 06) — esta task conserta o algoritmo atual, que continuará existindo como fallback/reconciliação.
- Indexação de diretórios (task 01).
- Ajustar o intervalo de 5s ou a janela de 2s — manter os valores atuais.
