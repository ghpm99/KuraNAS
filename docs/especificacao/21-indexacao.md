# 21 — Indexação

**Prioridade:** P1 · **Plataformas:** servidor; web (configuração e acompanhamento) · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Fazer com que tudo o que está no disco do servidor apareça no sistema — e só o que está no disco —, com miniaturas e metadados, sem o dono precisar mandar.

## 2. Cenários de uso

- O dono copia uma pasta de músicas direto para o disco do servidor (pelo Explorador, por WebDAV): em segundos elas estão em Música.
- Apaga um arquivo direto no disco: ele some do sistema.
- Joga arquivos em uma "caixa de entrada" e o sistema os distribui para as pastas de fotos, músicas, vídeos e documentos.
- Depois de uma queda de energia, o sistema se acerta sozinho com o que está no disco.

## 3. Escopo

Detecção de mudanças no disco, varredura de reconciliação, reindexação manual, processamento por tipo (metadados, miniaturas, texto, marcação por IA), bibliotecas por tipo, pastas monitoradas (caixa de entrada), preferências de indexação.

## 4. Fora de escopo

- Indexar o que está fora das raízes.
- Monitorar pastas de outros computadores da rede.
- Regras de organização personalizadas (por nome, por data, por expressão).
- Deduplicação automática.
- Antivírus ou verificação de conteúdo.

## 5. Conceitos

- **Índice**: o registro que o sistema mantém de cada pasta e arquivo das raízes.
- **Varredura (reconciliação)**: comparação completa entre disco e índice.
- **Biblioteca**: a pasta padrão de um tipo — imagens, músicas, vídeos, documentos. É o destino padrão de envios, importações e capturas daquele tipo. Não restringe o que aparece nas telas: uma foto em qualquer raiz aparece em Imagens.
- **Pasta monitorada**: pasta fora das raízes, usada como caixa de entrada; o que cai nela é movido para a biblioteca do tipo.

## 6. Regras de negócio

### Disco e índice

- **RN-01** — O índice reflete o disco. Pastas e arquivos são ambos indexados. Um arquivo criado, alterado, movido ou apagado no disco — por qualquer meio — está refletido no sistema em até 10 s, sem ação do usuário.
- **RN-02** — Em repouso, o sistema não percorre o disco: a detecção é por eventos do sistema operacional.
- **RN-03** — Uma varredura completa roda ao iniciar o servidor, periodicamente (padrão 24 h) e sempre que a detecção por eventos falha ou transborda. Ela corrige qualquer divergência.
- **RN-04** — Operações feitas pelo aplicativo atualizam o índice na própria operação, sem depender de detecção ([02/RN-04](02-arquivos.md)).
- **RN-05** — Arquivo que some do disco deixa de aparecer no sistema. Raiz cujo volume está desconectado **não** é tratada como "tudo apagado" ([20/RN-07](20-armazenamento.md)).
- **RN-06** — Arquivo movido ou renomeado **pelo aplicativo** mantém identidade e tudo o que depende dela. Movido por fora, é tratado como um arquivo removido e outro novo. *(ver [09/D-01](09-favoritos.md))*
- **RN-07** — Não são indexados: lixeira, temporários de envio, área de backup, tier frio (seus arquivos aparecem pelo caminho lógico), arquivos ocultos de sistema.
- **RN-08** — A indexação é idempotente e retomável: interrompida, continua de onde parou; repetida, não duplica nada.
- **RN-09** — Preferências e marcações do dono — estrela, playlists, álbuns, progresso, metadados corrigidos — nunca são alteradas pela indexação.

### Processamento

- **RN-10** — Arquivo novo ou alterado passa, em segundo plano e por ordem de prioridade: registro no índice (imediato, o arquivo já aparece) → checksum → metadados do tipo → miniatura → texto de documento → marcação por IA.
- **RN-11** — A falha de uma etapa não impede o arquivo de aparecer nem as outras etapas de rodar. Falhas ficam registradas com o motivo e podem ser reprocessadas ([16/RF-10](16-analytics.md)).
- **RN-12** — O processamento não degrada o uso: reprodução, navegação e envio têm prioridade sobre trabalho de fundo.
- **RN-13** — Toda indexação e processamento é job visível, com progresso ([15](15-notificacoes-e-jobs.md)).
- **RN-14** — Toda preferência de indexação exposta na tela tem efeito real; opção sem efeito não existe na tela.

### Reindexar

- **RN-15** — "Reindexar" age sobre o que o rótulo diz: a pasta atual quando acionado em uma pasta; a raiz ou tudo quando acionado nas configurações. Mostra o progresso e avisa ao terminar; a listagem se atualiza quando o job conclui.

### Bibliotecas

- **RN-16** — Cada tipo tem uma biblioteca, que deve ser uma pasta dentro de uma raiz habilitada. Na instalação recebem valores padrão.
- **RN-17** — Mudar a biblioteca de um tipo muda só o destino de novos arquivos; nada é movido.

### Pastas monitoradas

- **RN-18** — Uma pasta monitorada deve existir, ficar fora de todas as raízes e não repetir nem conter outra.
- **RN-19** — Arquivo que aparece nela é **movido** para a biblioteca do seu tipo quando está completo (tamanho e data de modificação inalterados há 10 s). Tipo não reconhecido fica onde está.
- **RN-20** — Nada é sobrescrito: nome já existente no destino recebe sufixo.
- **RN-21** — A estrutura de subpastas da pasta monitorada é preservada no destino.
- **RN-22** — Importações geram notificação agrupada ([15/RN-05](15-notificacoes-e-jobs.md)).

## 7. Requisitos por tela

### Configurações → Indexação

- **RF-01** — Situação: em dia / indexando (com progresso) / com erro; última varredura completa; próxima; pendências de processamento por etapa.
- **RF-02** — "Reindexar tudo" e reindexar por raiz, com confirmação e progresso.
- **RF-03** — Preferências com efeito real (RN-14): intervalo da varredura periódica; gerar miniaturas; extrair metadados; indexar texto de documentos; marcação por IA.

### Configurações → Bibliotecas

- **RF-04** — Para cada tipo: a pasta atual, com seletor de pasta e validação.

### Configurações → Pastas monitoradas

- **RF-05** — Lista: caminho, situação (ativa, pausada, indisponível), última verificação, arquivos importados. Adicionar (com teste de caminho), pausar/retomar, remover.
- **RF-06** — Explicação de que os arquivos são **movidos** da pasta monitorada.

### Arquivos

- **RF-07** — "Reindexar esta pasta" no menu da pasta ([02](02-arquivos.md)), com efeito e retorno conforme RN-15.

## 8. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| POST | `/files/update` | reindexação manual |
| GET, PUT | `/configuration/settings` | preferências de indexação |
| GET | `/libraries` · PUT `/libraries/:category` | bibliotecas |
| GET, POST | `/watch-folders` | pastas monitoradas |
| PUT, DELETE | `/watch-folders/:id` | editar, remover |
| GET | `/analytics/health` · `/analytics/processing` | situação |
| GET | `/jobs` | andamento |

**A criar** (não existe hoje):

| Método | Rota | Uso |
|---|---|---|
| POST | `/indexing/reindex` | reindexa tudo, uma raiz ou uma pasta, devolvendo o job (RN-15) |

- **`POST /indexing/reindex`** — corpo JSON: `{"scope": "all"}`, `{"scope": "root", "root_id": 2}` ou `{"scope": "folder", "folder_id": 345}` (a pasta e suas subpastas). Resposta `202`: `{"job_id"}`. `400` para `scope` inválido ou identificador ausente; `404` para raiz ou pasta inexistente; `409` com o `job_id` em andamento quando já há reindexação do mesmo escopo.
- `POST /files/update` continua existindo e equivale a `scope: all` (compatibilidade); passa a responder `{"job_id"}` no corpo.
- **Preferências de indexação** em `/configuration/settings` (RN-14) — cada chave e o efeito que deve ter:

| Chave | Efeito |
|---|---|
| `extract_metadata` (existe, hoje sem efeito) | desligada, a etapa de metadados do tipo não roda para arquivos novos ou alterados |
| `generate_previews` (existe, hoje sem efeito) | desligada, a etapa de miniatura/prévia não roda para arquivos novos ou alterados |
| `scan_interval_hours` (nova; 1 a 168, padrão 24) | intervalo da varredura periódica de RN-03 |
| `index_document_text` (nova; padrão ligada) | desligada, a etapa de texto de documento não roda |
| `scan_on_startup` (existe, hoje sem efeito) | **sai da tela**: a varredura ao iniciar é obrigatória (RN-03). A chave continua aceita e devolvida, sem efeito, por compatibilidade |

  Religar uma etapa não reprocessa o passado na hora: o que ficou sem processar entra pela reindexação. Exceção: as duas filas de pendência que já existem — documentos sem texto indexado e arquivos sem metadados, reconciliados ao iniciar — são processadas quando a etapa volta a ficar ligada; enquanto ela está desligada, essas filas não rodam. A marcação por IA é ligada e desligada em Configurações → IA ([18/RF-11](18-assistente.md)).

## 9. Situação em 2026-11-10

- As opções "varrer ao iniciar", "extrair metadados" e "gerar prévias" **não têm efeito**: nenhum código as lê (contraria RN-14).
- "Reescanear pasta" reindexa todas as raízes, não devolve o job e a tela recarrega antes de o trabalho terminar (contraria RN-15).
- Pastas monitoradas não têm tela na web nem no Android; só a API.
- Bibliotecas só aceitam subpasta da raiz principal.
- A validação da detecção por eventos no servidor Windows de produção está pendente (task 06).

## 10. Critérios de aceite

- [ ] **CA-01** — Copiar um arquivo para uma raiz pelo sistema operacional o faz aparecer em Arquivos em até 10 s e na biblioteca do tipo, com miniatura, em até 1 minuto.
- [ ] **CA-02** — Apagar um arquivo pelo sistema operacional o tira de todas as telas em até 10 s.
- [ ] **CA-03** — Criar uma pasta nova com arquivos dentro, pelo sistema operacional: pasta e arquivos aparecem.
- [ ] **CA-04** — Copiar 5 000 arquivos de uma vez: todos são indexados; o total confere com o disco.
- [ ] **CA-05** — Com o servidor parado, adicionar e remover arquivos; ao iniciar, o índice é corrigido.
- [ ] **CA-06** — Em repouso, o servidor não lê o disco continuamente (o disco pode hibernar).
- [ ] **CA-07** — Desconectar e reconectar o volume de uma raiz não remove itens do índice nem os manda para a lixeira.
- [ ] **CA-08** — Reindexar não remove estrelas, playlists, álbuns, progresso de vídeo nem metadados corrigidos.
- [ ] **CA-09** — Um arquivo corrompido aparece em Arquivos; a falha de metadados fica registrada com motivo e não trava a fila.
- [ ] **CA-10** — Durante a indexação de uma biblioteca grande, tocar música e assistir vídeo continuam sem nenhuma pausa para carregar em 10 minutos de reprodução.
- [ ] **CA-11** — "Reindexar esta pasta" reindexa só aquela pasta, mostra progresso e a listagem se atualiza ao concluir.
- [ ] **CA-12** — Desligar "gerar miniaturas" faz arquivos novos não terem miniatura gerada; religar volta a gerar.
- [ ] **CA-13** — Um arquivo colocado em uma pasta monitorada é movido para a biblioteca do tipo e aparece no sistema; um arquivo de mesmo nome já existente não é sobrescrito.
- [ ] **CA-14** — Um arquivo grande ainda sendo copiado para a pasta monitorada só é movido depois de completo.
- [ ] **CA-15** — Lixeira, área de backup e temporários não aparecem em nenhuma tela de conteúdo.
- [ ] **CA-16** — A detecção por eventos funciona no servidor Windows de produção.

## 11. Decisões em aberto

- **D-01** — Bibliotecas em qualquer raiz (RN-16) ou só na principal (como hoje). *Provisório: qualquer raiz habilitada.*
- **D-02** — Pasta monitorada move (como hoje) ou copia. *Provisório: move, com a explicação de RF-06.*

## 12. Rastreabilidade

`docs/melhorias/` tasks 01, 02, 03, 05, 07, 08 (concluídas) e 06 (bloqueada, aguardando validação no Windows); `docs/revisao-produto` item 16 (pendente).
