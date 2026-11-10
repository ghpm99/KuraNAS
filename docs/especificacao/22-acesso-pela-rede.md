# 22 — Acesso pela rede

**Prioridade:** P1 · **Plataformas:** web (configuração); todos os clientes (efeito) · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Deixar que só os dispositivos da casa usem o servidor, sem login, e permitir montar o servidor como unidade de rede nos computadores.

## 2. Cenários de uso

- O dono libera o celular novo: abre Configurações em um PC já liberado e adiciona o IP.
- Uma visita conectada ao Wi-Fi tenta abrir o endereço do servidor e é bloqueada.
- No PC, o servidor aparece como uma unidade no Explorador de Arquivos; ele arrasta arquivos para lá.

## 3. Escopo

Whitelist de IPs (listar, adicionar, editar, habilitar/desabilitar, remover, liberar o dispositivo atual), bloqueio de quem não está na lista, acesso por WebDAV.

## 4. Fora de escopo

- **Login, senha, token, sessão, usuários, 2FA** — decisão registrada: não haverá autenticação.
- HTTPS/TLS, enquanto o uso for rede interna.
- Acesso de fora da rede local (VPN, túnel, redirecionamento de porta são do dono, fora do sistema).
- Permissões por pasta ou por dispositivo (somente leitura para um IP, por exemplo).
- SMB, NFS, FTP, DLNA.
- Limite de tentativas, detecção de ataque.

## 5. Regras de negócio

### Whitelist

- **RN-01** — Toda requisição — API, interface web, arquivos estáticos, WebDAV — só é atendida se o IP de origem é o próprio servidor (loopback) ou casa com uma entrada habilitada. As demais recebem `403` com mensagem traduzida.
- **RN-02** — O próprio servidor (localhost) sempre tem acesso, mesmo com a lista vazia: não é possível se trancar para fora.
- **RN-03** — Uma entrada é um IP (v4 ou v6) ou uma faixa CIDR, com um nome para identificar o dispositivo. Valor inválido é recusado com mensagem; entradas duplicadas são recusadas.
- **RN-04** — Mudanças na lista valem imediatamente, sem reiniciar.
- **RN-05** — O IP considerado é o da conexão. Cabeçalhos como `X-Forwarded-For` são ignorados.
- **RN-06** — Um cliente IPv4 que chega por socket IPv6 casa com a entrada IPv4 correspondente.
- **RN-07** — Remover ou desabilitar a entrada do dispositivo que está sendo usado pede confirmação avisando que o acesso deste dispositivo será perdido.
- **RN-08** — A descoberta do servidor na rede ([24](24-app-android.md)) continua respondendo a todos; a proteção é no acesso aos dados.

### WebDAV

- **RN-09** — Desligado por padrão. Ligado, expõe as raízes habilitadas em `/dav/`, para leitura e escrita, sujeito à mesma whitelist.
- **RN-10** — O que é feito por WebDAV vale como feito pelo aplicativo: aparece no índice sem ação manual, e excluir vai para a lixeira ([10/RN-01](10-lixeira.md)).
- **RN-11** — Pastas internas (lixeira, temporários) não aparecem nem são acessíveis por WebDAV; arquivos do tier frio aparecem no caminho lógico.
- **RN-12** — Arquivos grandes (mais de 1 GB) sobem e descem sem ocupar memória proporcional ao tamanho.

## 6. Requisitos por tela

### Configurações → Acesso

- **RF-01** — Mostra o IP do dispositivo atual e se ele está liberado por uma entrada (e qual).
- **RF-02** — Lista das entradas: nome, IP ou faixa, habilitada, data de cadastro. Ações: editar, habilitar/desabilitar, remover.
- **RF-03** — Adicionar entrada: nome e IP/faixa, com validação no campo. "Liberar este dispositivo" preenche o IP atual.
- **RF-04** — Texto curto explicando que não há senha e que quem está na lista tem acesso total.

### Configurações → WebDAV

- **RF-05** — Situação (ligado/desligado), o endereço para montar e instruções curtas para Windows, Linux e Android. Se ligar/desligar exige reinício ou variável de ambiente, a tela diz isso claramente.

### Tela de bloqueio

- **RF-06** — Dispositivo não liberado que abre o endereço vê uma página simples, traduzida, dizendo que o acesso não está liberado e mostrando o próprio IP para o dono cadastrar.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/access-control/client-ip` | IP do dispositivo atual |
| GET, POST | `/access-control/ips` | listar, adicionar |
| PUT, DELETE | `/access-control/ips/:id` | editar, remover |
| — | `/dav/*` | WebDAV |

## 8. Estados e erros

| Situação | Resposta |
|---|---|
| IP fora da lista | `403` em qualquer rota |
| IP ou CIDR inválido | `400` com motivo |
| Entrada duplicada | `409` |
| WebDAV desligado | `/dav/` não existe (`404`) |

## 9. Critérios de aceite

- [ ] **CA-01** — Um dispositivo fora da lista recebe a página de bloqueio ao abrir o endereço e `403` em qualquer chamada de API.
- [ ] **CA-02** — Adicionar o IP desse dispositivo o libera na hora, sem reiniciar o servidor.
- [ ] **CA-03** — No próprio servidor, o acesso funciona com a lista vazia.
- [ ] **CA-04** — Cadastrar `192.168.1.0/24` libera todos os dispositivos dessa faixa.
- [ ] **CA-05** — Enviar `X-Forwarded-For` com um IP liberado a partir de um dispositivo bloqueado não dá acesso.
- [ ] **CA-06** — Um valor que não é IP nem CIDR é recusado com mensagem no campo.
- [ ] **CA-07** — Remover a entrada do dispositivo em uso avisa da perda de acesso antes de confirmar.
- [ ] **CA-08** — O app Android e o plugin, em dispositivo liberado, funcionam; em dispositivo bloqueado, mostram a mensagem de acesso não liberado em vez de erro genérico.
- [ ] **CA-09** — Com WebDAV ligado, o servidor monta como unidade no Windows e no Linux; é possível abrir, criar, renomear e excluir.
- [ ] **CA-10** — Um arquivo copiado por WebDAV aparece em Arquivos e na biblioteca do seu tipo sem ação manual.
- [ ] **CA-11** — Um arquivo excluído por WebDAV aparece na Lixeira.
- [ ] **CA-12** — Copiar um arquivo de 2 GB por WebDAV conclui sem erro.
- [ ] **CA-13** — De dispositivo fora da lista, `/dav/` responde `403`.

## 10. Decisões em aberto

- **D-01** — Excluir por WebDAV vai para a lixeira (RN-10) ou é definitivo, como em uma unidade de rede comum. *Provisório: lixeira, coerente com "nada some".*

## 11. Rastreabilidade

`docs/melhorias/` tasks 04 (whitelist, concluída) e 11 (WebDAV, bloqueada aguardando validação manual do dono — CA-09 a CA-12).
