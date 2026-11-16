# 13 — Takeout

**Prioridade:** P2 · **Plataformas:** web · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Trazer para o servidor, de uma vez, as fotos e vídeos que o dono tem guardados no Google Fotos, a partir do arquivo de exportação do Google Takeout, preservando as datas e sem perder nem duplicar nada.

## 2. Cenários de uso

- O dono exporta o Google Fotos, recebe vários arquivos `.zip` de dezenas de GB e envia cada um pela tela de Takeout; ao final, as fotos estão na galeria, nas datas em que foram tiradas.
- Envia o mesmo zip duas vezes por engano e nada é duplicado.

## 3. Escopo

Envio do zip em partes com retomada, importação em segundo plano (fotos e vídeos), aproveitamento dos metadados do Takeout, relatório do que foi importado, acompanhamento.

## 4. Fora de escopo

- Outros produtos do Takeout (Drive, Gmail, Contatos, YouTube): arquivos que não são foto nem vídeo são ignorados e contados no relatório.
- Importar álbuns do Google como álbuns do KuraNAS. *(ver D-01)*
- Exportações de outros serviços (iCloud, OneDrive).
- Baixar o Takeout direto da conta Google.
- Formatos de arquivo diferentes de `.zip` (`.tgz`).

## 5. Regras de negócio

- **RN-01** — Só `.zip` é aceito; outro formato é recusado antes do envio, com mensagem.
- **RN-02** — O zip é enviado em partes e o envio retoma após queda de conexão. O tamanho recebido precisa conferir com o declarado.
- **RN-03** — A importação é um job em segundo plano ([15](15-notificacoes-e-jobs.md)), com progresso por quantidade de arquivos.
- **RN-04** — São importados fotos e vídeos (os mesmos formatos de [04/RN-01](04-imagens.md) e [06/RN-01](06-videos.md)). O restante é ignorado e contado.
- **RN-05** — Destino: a biblioteca de imagens ou de vídeos, em `takeout/<nome do álbum ou pasta de origem>/`, preservando a estrutura do zip. *(ver D-02)*
- **RN-06** — **Nada é sobrescrito.** Arquivo idêntico (mesmo conteúdo) já existente no destino é pulado; arquivo diferente com o mesmo nome recebe sufixo.
- **RN-07** — Os metadados que acompanham cada foto no Takeout são aplicados: a data em que a foto foi tirada (quando o arquivo não a tem no EXIF), descrição e localização.
- **RN-08** — Entradas maliciosas do zip (caminho que tenta sair do destino, links simbólicos) são ignoradas e contadas.
- **RN-09** — A extração não depende do tamanho do zip nem dos arquivos dentro dele: memória e disco temporário são limitados.
- **RN-10** — Ao terminar (com sucesso ou falha), o zip enviado e os temporários são apagados.
- **RN-11** — Conclusão e falha geram notificação com o resumo: importados, já existentes, ignorados, com erro. A mensagem de falha é traduzida e não expõe erro interno.
- **RN-12** — Zip inválido ou corrompido falha com mensagem específica; o que foi importado antes da falha permanece.

## 6. Requisitos por tela

### Takeout (`/takeout`)

- **RF-01** — Explicação curta de como gerar a exportação no Google, área de arrastar e soltar e botão de selecionar; aceita vários zips em fila.
- **RF-02** — Por zip: nome, tamanho, barra de progresso do envio com velocidade e tempo restante, "Cancelar". Após o envio, a etapa de importação com progresso (arquivos processados / total).
- **RF-03** — Ao concluir: resumo de RN-11 e atalhos "Ver fotos importadas" e "Ver vídeos importados".
- **RF-04** — Falha mostra o motivo vindo do servidor e "Tentar de novo".
- **RF-05** — Histórico das importações anteriores: data, arquivo, resultado, resumo.
- **RF-06** — Sair da tela durante o envio pede confirmação; a importação (já no servidor) continua mesmo com a tela fechada.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| POST | `/takeout/upload/init` | abre sessão (`file_name`, `size`) |
| POST | `/takeout/upload/chunk` | envia parte |
| POST | `/takeout/upload/complete` | conclui e inicia a importação → `job_id` |
| GET | `/jobs/:id` · `/jobs?type=takeout_import` | andamento e histórico |

**A criar** (não existe hoje):

| Método | Rota | Uso |
|---|---|---|
| GET | `/takeout/upload/:upload_id` | quanto já foi recebido → `{"upload_id", "size", "received_size"}` |
| DELETE | `/takeout/upload/:upload_id` | cancela o envio e apaga o que foi recebido (RF-02) → `204` |
| GET | `/takeout/imports` | histórico de importações, paginado, mais recente primeiro (RF-05) |
| GET | `/takeout/imports/:id` | uma importação, com o resumo (RF-03) |

- **Retomada**: `POST /takeout/upload/init` com o mesmo `file_name` e `size` de uma sessão aberta devolve o mesmo `upload_id`, mais `received_size` e `resumed: true`. Sem `size` no pedido → `400`. Espaço livre menor que `size` → `507`.
- **Item de `/takeout/imports`**: `id`, `job_id`, `file_name`, `size`, `state` (`queued`, `running`, `completed`, `failed`), `total_entries`, `processed_entries`, `imported`, `already_existing`, `ignored`, `failed_entries`, `error` (traduzido, vazio fora de `failed`), `created_at`, `ended_at`. Vale `imported + already_existing + ignored + failed_entries = total_entries` ao terminar (CA-09).
- `POST /takeout/upload/complete` passa a devolver também `import_id`.
- **"Tentar de novo" (RF-04)**: como o zip é apagado ao terminar (RN-10), repetir é enviar o arquivo de novo; arquivos já importados são pulados (RN-06).
- Os destinos de "Ver fotos importadas" e "Ver vídeos importados" são as pastas `takeout/` das bibliotecas de imagens e de vídeos, abertas na galeria e no catálogo com o filtro de pasta.

## 8. Estados e erros

| Situação | O que o usuário vê |
|---|---|
| Arquivo não é `.zip` | recusa imediata |
| Zip corrompido | falha com mensagem específica |
| Sem espaço (temporário ou destino) | o `init` recusa quando o espaço livre é menor que o tamanho declarado; faltando espaço durante a importação, o job falha com motivo |
| Biblioteca de imagens/vídeos não configurada | falha com atalho para Configurações |
| Conexão caiu no envio | retoma |

## 9. Situação em 2026-11-10

- Os metadados do Takeout são lidos e descartados (RN-07 não é atendida).
- A estrutura é achatada em `takeout/` e arquivos de mesmo nome **se sobrescrevem** (contraria RN-05 e RN-06).
- Cada arquivo do zip é carregado inteiro em memória (contraria RN-09).
- A tela descarta a mensagem do servidor, não cancela, não retoma, não acompanha a importação nem tem histórico; a notificação de falha expõe o erro interno.

## 10. Critérios de aceite

- [ ] **CA-01** — Enviar um zip de Takeout de 10 GB conclui e as fotos aparecem na galeria.
- [ ] **CA-02** — Uma foto sem EXIF, cuja data está só no arquivo de metadados do Takeout, aparece na galeria na data correta.
- [ ] **CA-03** — Duas fotos diferentes com o mesmo nome, em pastas diferentes do zip, são as duas importadas.
- [ ] **CA-04** — Enviar o mesmo zip duas vezes não duplica nenhum arquivo; o resumo da segunda vez diz "já existentes".
- [ ] **CA-05** — Derrubar a conexão durante o envio e religar: o envio continua.
- [ ] **CA-06** — Fechar a tela durante a importação e voltar mostra o andamento atual.
- [ ] **CA-07** — Um zip com um vídeo de 3 GB dentro é importado sem o servidor ficar sem memória.
- [ ] **CA-08** — Um arquivo que não é zip é recusado; um zip corrompido falha com mensagem traduzida que diz que o arquivo está corrompido, sem erro interno.
- [ ] **CA-09** — O resumo final confere com o conteúdo do zip (importados + já existentes + ignorados + erros = total de entradas).
- [ ] **CA-10** — Depois de concluir, o zip e os temporários não ocupam espaço no servidor.
- [ ] **CA-11** — A notificação de conclusão leva às fotos importadas.

## 11. Decisões em aberto

- **D-01** — Criar álbuns do KuraNAS a partir dos álbuns do Google. *Provisório: não; os álbuns viram pastas (RN-05).*
- **D-02** — Organização do destino: por pasta de origem do zip (RN-05) ou por data (`AAAA/MM`), como o envio automático do celular ([03/RN-12](03-transferencia-entre-dispositivos.md)). *Provisório: pasta de origem.*

## 12. Rastreabilidade

`docs/revisao-produto` item 11 (pendente). Sem task em `docs/melhorias/`.
