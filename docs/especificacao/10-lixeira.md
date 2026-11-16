# 10 — Lixeira

**Prioridade:** P1 · **Plataformas:** web, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Garantir que excluir por engano não seja perder: tudo o que é excluído pelo aplicativo fica recuperável por um período.

## 2. Cenários de uso

- O dono exclui uma pasta errada organizando documentos; abre a Lixeira e restaura.
- Quer liberar espaço: vê quanto a lixeira ocupa e esvazia.
- Não quer pensar nisso: itens antigos somem sozinhos depois do prazo.

## 3. Escopo

Exclusão recuperável, tela da Lixeira (listar, buscar, ordenar, restaurar, excluir de vez, esvaziar, lote), retenção automática, resumo de ocupação.

## 4. Fora de escopo

- Versionamento de arquivos e snapshots.
- Lixeira por dispositivo ou por pessoa.
- Recuperar o que foi excluído por fora do aplicativo (direto no disco, por outro programa).
- Navegar dentro de uma pasta que está na lixeira ou restaurar só parte dela.

## 5. Regras de negócio

- **RN-01** — Excluir pelo aplicativo — em qualquer tela, na web, no Android ou por WebDAV — move o item para a lixeira. Os bytes continuam no disco, na mesma raiz.
- **RN-02** — Exclusão definitiva direta só acontece por escolha explícita do dono, com confirmação que diz que não há volta.
- **RN-03** — Se a lixeira não puder ser usada (raiz sem lixeira, erro ao mover), a exclusão **falha com mensagem**; o sistema nunca cai para exclusão definitiva por conta própria.
- **RN-04** — Item na lixeira some de todas as telas, da busca, das playlists, dos álbuns, do backup e do WebDAV, e não é indexado.
- **RN-05** — Excluir uma pasta gera um único item de lixeira, com todo o conteúdo dentro.
- **RN-06** — Restaurar devolve o item ao caminho original com a mesma identidade: estrela, progresso, histórico e participação em playlists e álbuns voltam.
- **RN-07** — Se a pasta de origem não existe mais, restaurar a recria.
- **RN-08** — Se o caminho original está ocupado, restaurar responde `409` informando o caminho; o cliente oferece "restaurar com outro nome" (sufixo ` (restaurado)`). Nada é sobrescrito.
- **RN-09** — Retenção: itens com mais de N dias na lixeira são removidos de vez automaticamente. N é configurável (padrão 30, mínimo 1). A tela mostra em quantos dias cada item expira.
- **RN-10** — Remoção definitiva (manual ou automática) apaga os bytes — inclusive cópias no tier frio — e os registros do item.
- **RN-11** — O tamanho de um item de pasta é a soma do conteúdo.
- **RN-12** — Item cujos bytes sumiram da lixeira (apagados por fora) aparece marcado como irrecuperável e pode ser removido da lista; restaurá-lo dá erro específico.

## 6. Requisitos por tela

### Diálogo de exclusão (em qualquer tela)

- **RF-01** — Diz o que será excluído (nome ou quantidade), que vai para a Lixeira e por quantos dias fica. Oferece a opção "excluir definitivamente", desmarcada.
- **RF-02** — Depois de excluir, um aviso confirma e oferece "Desfazer" por 10 s, que restaura o item.

### Lixeira (`/trash`)

- **RF-03** — Cabeçalho com total de itens e espaço ocupado, busca por nome, ordenação, Recarregar, "Esvaziar lixeira" e o prazo de retenção com atalho para alterá-lo.
- **RF-04** — Lista paginada. Cada item: ícone de pasta/arquivo (miniatura quando houver), nome, caminho de origem, tamanho, excluído em, expira em.
- **RF-05** — Ordenação padrão: excluído mais recentemente primeiro. Opções: nome, tamanho, data de exclusão.
- **RF-06** — Ações por item: Restaurar, Excluir definitivamente. Depois de restaurar, o aviso oferece "Ver na pasta".
- **RF-07** — Seleção e lote: Restaurar selecionados, Excluir selecionados definitivamente; resultado por item.
- **RF-08** — "Esvaziar lixeira" pede confirmação com a quantidade e o espaço a liberar, e informa o que não pôde ser removido.
- **RF-09** — Um botão em uso trava só o item dele; a lista continua utilizável.
- **RF-10** — Clique em um item abre os detalhes dele (nome, origem, tamanho, datas, tipo); o conteúdo não é aberto enquanto estiver na lixeira.

### Retenção

- **RF-11** — Campo de dias com validação; valor inválido é apontado no campo; salvar confirma.

### Android

- **RF-12** — Excluir no explorer avisa que vai para a Lixeira. Tela de Lixeira paginada com restaurar, excluir definitivamente e esvaziar.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| DELETE | `/files/path` | excluir (para a lixeira; `permanent=true` definitivo) |
| GET | `/trash` | lista paginada (busca, ordenação) |
| GET | `/trash/summary` | total de itens e bytes |
| POST | `/trash/:id/restore` | restaurar |
| DELETE | `/trash/:id` | excluir definitivamente |
| POST | `/trash/restore` · `/trash/delete` | lote, com resultado por id |
| DELETE | `/trash` | esvaziar (`{purged, failed}`) |
| GET, PUT | `/trash/retention` | prazo de retenção |

## 8. Estados e erros

| Situação | Resposta | O que o usuário vê |
|---|---|---|
| Lixeira vazia | — | estado vazio: "nada na lixeira" e o prazo de retenção |
| Caminho original ocupado | `409` com o caminho | opção de restaurar com outro nome |
| Bytes ausentes | erro específico | item marcado como irrecuperável |
| Lixeira indisponível ao excluir | erro | mensagem; nada é apagado |
| Falha parcial ao esvaziar | `{purged, failed}` | quantos saíram e quantos ficaram |

## 9. Critérios de aceite

Além dos CA de [00](00-padroes-transversais.md) na lista da Lixeira:

- [ ] **CA-01** — Excluir um arquivo em Arquivos, uma foto em Imagens e uma faixa em Música: os três aparecem na Lixeira e somem das telas de origem e da busca.
- [ ] **CA-02** — Os bytes de um item excluído continuam no disco do servidor.
- [ ] **CA-03** — Restaurar devolve o item ao mesmo lugar, com estrela, e ele reaparece nas telas sem recarregar.
- [ ] **CA-04** — Excluir e restaurar uma pasta com subpastas devolve a árvore inteira.
- [ ] **CA-05** — Restaurar quando já existe outro arquivo no caminho mostra o conflito com o caminho e permite restaurar com outro nome; o arquivo existente fica intacto.
- [ ] **CA-06** — Restaurar um item cuja pasta de origem foi excluída recria a pasta.
- [ ] **CA-07** — "Desfazer" logo após excluir restaura o item.
- [ ] **CA-08** — O resumo mostra a quantidade e o espaço reais; o tamanho de uma pasta é o do conteúdo.
- [ ] **CA-09** — Excluir definitivamente remove os bytes do disco.
- [ ] **CA-10** — Um item mais antigo que a retenção é removido automaticamente; mudar a retenção para 1 dia remove os itens com mais de 1 dia na execução seguinte.
- [ ] **CA-11** — Restaurar 20 itens em lote informa o resultado de cada um.
- [ ] **CA-12** — Restaurar o último item de uma página não deixa a tela vazia se ainda há itens.
- [ ] **CA-13** — Um arquivo do tier frio excluído e restaurado abre normalmente.
- [ ] **CA-14** — Uma música excluída sai das playlists; restaurada, volta para elas.
- [ ] **CA-15** — No Android: excluir, ver na Lixeira e restaurar funcionam e refletem na web.

## 10. Decisões em aberto

Correspondem às lacunas 31–36 do board.

- **D-01** — Retenção "nunca expirar" e/ou teto de espaço da lixeira. *Provisório: só prazo em dias (RN-09).*
- **D-02** — Conflito ao restaurar. *Provisório: RN-08 (avisa e oferece outro nome).*
- **D-03** — Pai ausente ou também na lixeira. *Provisório: RN-07 (recria as pastas).*
- **D-04** — Espaço da lixeira no Analytics. *Provisório: mostrar só na própria Lixeira (RF-03) e remover as chaves órfãs de Analytics.*
- **D-05** — "Desfazer" após excluir. *Provisório: em escopo (RF-02).*
- **D-06** — Notificação quando a limpeza automática remove itens. *Provisório: não notificar remoção normal; notificar falha.*

## 11. Rastreabilidade

`docs/melhorias/09-lixeira.md`; `docs/revisao-produto/08-lixeira.md` (30 lacunas + 6 decisões).
