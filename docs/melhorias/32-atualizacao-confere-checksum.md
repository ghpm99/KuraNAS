# 32 — Atualização do sistema confere a integridade do pacote

**Tipo:** segurança · **Prioridade:** P1 · **Spec:** [23](../especificacao/23-configuracoes-e-sistema.md) `RN-15`, `RN-17`, §7, `CA-11`, `D-03`

## Contexto

`DownloadAndApply` (`backend/internal/api/v1/updater/service.go:92-148`) baixa o pacote do release e valida **só o tamanho** (`info.Size() != expectedSize`, `:130-133`). Não há checksum, hash nem assinatura no pacote `updater`. Um pacote corrompido ou trocado, do mesmo tamanho, é aplicado como binário do servidor. A atualização do yt-dlp já faz a conferência por `SHA2-256SUMS` (`ingest/ytdlp_service.go`) e serve de referência.

## Objetivo

O servidor só aplica um pacote de atualização cujo SHA-256 confere com o publicado junto do release.

## O que fazer

1. **Publicação**: o processo de release passa a gerar e anexar `SHA256SUMS` (uma linha `<hash>  <nome do pacote>` por artefato). Documentar o passo no `Makefile`/README de release.
2. **Conferência**: `DownloadAndApply` baixa `SHA256SUMS` (limite de 1 MB), calcula o SHA-256 do pacote baixado e compara antes de aplicar. A conferência de tamanho continua.
3. **Recusa**: release sem `SHA256SUMS`, sem linha para o pacote ou com hash diferente → o pacote é apagado, nada é aplicado e `POST /update/apply` responde `409` com mensagem própria para cada caso (três chaves de tradução, pt-BR e en-US).
4. **Limite do download** do pacote (512 MB), para um release adulterado não encher o disco.
5. **Web**: a seção de atualização mostra a mensagem do servidor na recusa (`00/RN-41`).

## Critérios de aceite

- [ ] Teste de serviço: hash que confere aplica; hash diferente não aplica e apaga o pacote; release sem `SHA256SUMS` não aplica.
- [ ] Teste de serviço: pacote acima do limite é recusado.
- [ ] Teste de handler: os casos de recusa respondem `409` com a mensagem certa.
- [ ] `23/CA-11` — no sistema rodando: um pacote adulterado é recusado e a versão atual continua funcionando.
- [ ] O release seguinte a esta task é publicado com `SHA256SUMS` e a atualização a partir dele funciona.
- [ ] `make ci` verde.

## Fora de escopo

- Assinatura criptográfica do release.
- Atualização automática sem ação do dono (`23/RN-14`).
- Reversão automática depois de aplicado.
