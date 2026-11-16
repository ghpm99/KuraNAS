# 28 — Takeout: extrair sem carregar cada arquivo na memória

**Tipo:** bug · **Prioridade:** P1 · **Spec:** [13](../especificacao/13-takeout.md) `RN-09`, `RN-10`, `RN-12`, `CA-07`, `CA-10`

## Contexto

O zip é lido do disco (`zip.OpenReader`, `takeout/extractor.go:93`), mas **cada entrada é carregada inteira na memória**: `readZipEntry` (`:189-205`) copia para um `bytes.Buffer` e `os.WriteFile` grava depois (`:163-167`). Um vídeo de 3 GB dentro do zip ocupa 3 GB de RAM e pode derrubar o servidor. O zip e a pasta da sessão só são apagados no sucesso (`backend/internal/worker/engine/step_takeout.go:50-51`): em falha, ficam ocupando disco.

## Objetivo

A importação usa memória constante, qualquer que seja o tamanho dos arquivos do zip, e sempre limpa o que enviou.

## O que fazer

1. **Cópia em fluxo**: cada mídia é copiada do zip para um arquivo provisório `.part` no destino com `io.Copy` e renomeada ao fim; o SHA-256 usado pela task 26 é calculado durante a cópia. Só os `.json` (limitados a 1 MB cada) são lidos em memória.
2. **Entrada com falha** não interrompe a importação: o `.part` é removido, a entrada conta como falha e o restante continua (`RN-12`).
3. **Limpeza sempre** (`RN-10`): zip e pasta da sessão são apagados ao terminar, com sucesso ou falha.
4. **Progresso** por quantidade de entradas processadas (`RN-03`), atualizado no job e no registro da importação.

## Critérios de aceite

- [ ] Teste do extrator com uma entrada de 200 MB gerada no teste: a memória alocada durante a extração fica abaixo de 32 MB.
- [ ] Teste: uma entrada corrompida no meio do zip não impede as seguintes; nenhum `.part` sobra no destino.
- [ ] Teste do step: em falha, zip e pasta da sessão são removidos.
- [ ] `13/CA-07` — no sistema rodando: um zip com um vídeo de 3 GB é importado sem o servidor ficar sem memória.
- [ ] `13/CA-10` — depois de concluir, o zip e os temporários não ocupam espaço.
- [ ] `make ci-backend` verde.

## Fora de escopo

- Suporte a `.tgz`.
- Paralelizar a extração.

## Depende de

Task 26.
