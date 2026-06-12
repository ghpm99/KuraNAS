# 10 — Múltiplas raízes de armazenamento

**Tipo:** feature de maturidade · **Prioridade:** P3

## Contexto

Todo o sistema gira em torno de **uma única raiz**: o env `ENTRY_POINT` (`internal/config`). O scan de boot, o watcher, a árvore de arquivos (`GetFilesTreeHandler` usa `config.AppConfig.EntryPoint` como raiz default), a validação de paths das operações (`resolvePathInEntryPoint`) e a tradução relativo/absoluto (`ToRelativePath`/`ToAbsolutePath`) assumem essa raiz única.

Um NAS real quase nunca tem um disco só: HD interno + externo, partições separadas para mídia e documentos. Hoje o usuário teria que mover tudo para debaixo de uma pasta única ou rodar múltiplas instâncias (cada uma com seu Postgres e sua porta).

Observação: o domínio `libraries` e os `watchfolders` já apontam para pastas específicas, mas **dentro** do conceito de raiz única — não são raízes de indexação independentes.

### Visão de armazenamento que esta task fundamenta (registrada 2026-06-11)

Esta task deixou de ser "feature futura isolada" e virou **fundação** do desenho de armazenamento do dono: SSD 512 GB como tier quente, pool de HDs como tier frio + área de backup, HD externo 2 TB como segunda cópia. As tasks **12 (backup)** e **13 (tiering)** dependem das múltiplas raízes definidas aqui.

**Decisão de infraestrutura (fora do código):** redundância contra disco queimado **não será implementada no KuraNAS**. Os HDs avulsos serão agrupados em pool espelhado pelo **Windows Storage Spaces** (nativo, discos de tamanhos diferentes, reconstrução ao trocar disco), que apresenta o pool ao sistema como um volume comum — a redundância fica transparente para o KuraNAS, zero código. Recomendação registrada: discos do pool em SATA interno quando possível; USB fica para a segunda cópia de backup desconectável.

## Objetivo

O usuário cadastra N raízes de armazenamento (ex.: `D:\Arquivos`, `E:\Midia`); todas são indexadas, vigiadas e navegáveis, e os paths fora delas continuam inacessíveis pela API.

## O que fazer

1. Modelar raízes no banco (tabela `storage_root`), mantendo `ENTRY_POINT` como semente/compatibilidade.
2. Generalizar scan, watcher e validação de path de "a raiz" para "a lista de raízes".
3. Expor as raízes na API e na navegação do frontend (nível zero da árvore = lista de raízes).

## Como fazer

- **Modelo**: migração criando `storage_root` (`id`, `path`, `label`, `enabled`, `created_at`). No boot, se a tabela está vazia e `ENTRY_POINT` está definido, semear com ele — instalações existentes continuam funcionando sem ação do usuário.
- **Domínio**: pacote `internal/api/v1/storageroots/` no padrão do projeto, com CRUD pequeno (`GET/POST/PUT/DELETE /api/v1/storage-roots`). Validações: path existe, é diretório, não é ancestral/descendente de outra raiz cadastrada.
- **Generalizações** (procurar usos de `config.AppConfig.EntryPoint`):
  - `enqueueStartupScanJob` → um job `startup_scan` por raiz habilitada;
  - watcher de entry point → uma instância (ou um conjunto de watches) por raiz;
  - `GetFilesTreeHandler` com `file_parent=0` → retorna as raízes como nós de topo (linhas de `home_file` das raízes; garantir que existam — depende da task 01);
  - `resolvePathInEntryPoint` → `resolvePathInRoots` (path precisa estar sob **alguma** raiz habilitada); idem `resolveTargetFolder`;
  - `ToRelativePath`/`ToAbsolutePath` → resolver contra a raiz dona do path.
- **Mover/copiar entre raízes**: `os.Rename` falha entre volumes (`EXDEV`); detectar e cair para copiar+excluir, ou (mínimo aceitável) responder erro claro via i18n nesta primeira versão.
- **Analytics**: as queries agregadas continuam funcionando (são sobre `home_file` inteira); avaliar filtro por raiz como melhoria posterior.
- **Frontend**: nível zero da navegação lista as raízes; configuração das raízes na tela de Settings.
- **Testes**: validação de raízes (sobreposição, path inválido), resolução de path multi-raiz, árvore de topo, scan por raiz.

## Critérios de aceite

- [ ] Instalação existente migra sozinha: `ENTRY_POINT` vira a primeira raiz e tudo segue funcionando igual.
- [x] Cadastrar uma segunda raiz dispara indexação dela e ela aparece como nó de topo na árvore.
- [ ] Operações de arquivo (upload, move, rename, delete) funcionam em qualquer raiz e recusam paths fora de todas (400, mensagem i18n).
- [x] Raiz sobreposta a outra (ancestral/descendente) é recusada no cadastro.
- [x] Desabilitar uma raiz a tira da navegação sem apagar dados indexados.
- [ ] `make ci` verde (backend + frontend).

## Fora de escopo

- Gestão de discos/volumes/RAID/SMART — KuraNAS opera sobre paths, não sobre hardware.
- Quotas por raiz.
- Migração de dados entre raízes em massa.
- Mudar os apps Android/plugin (a API de árvore mantém o shape; raiz nova aparece como pasta de topo).
