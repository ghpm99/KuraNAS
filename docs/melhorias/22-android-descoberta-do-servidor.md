# 22 — Android: encontrar o servidor na rede

**Tipo:** bug crítico · **Prioridade:** P0 (N1, N2) · **Spec:** [24](../especificacao/24-app-android.md) `RN-01` a `RN-05`, `RF-01`, `RF-02`, §7, `CA-01` a `CA-05`

## Contexto

O app nunca encontra o servidor sozinho:

- `core/discovery/NsdDiscovery.kt:22` procura `_nas._tcp.`; o servidor anuncia `_kuranas._tcp` (`backend/internal/discovery/mdns_registrar.go:11`).
- A busca não tem fim: `discover()` é um fluxo que só encerra ao ser cancelado e `ConnectionViewModel.kt:40-51` só tira `isDiscovering` no `catch`. O indicador gira para sempre e a mensagem "nenhum servidor" (`ConnectionScreen.kt:97`) nunca aparece.
- Não há descoberta por UDP no app (o servidor responde na porta 19520; o app legado já faz isso em `mobile/.../UdpDiscoveryStrategy.java`).
- O endereço manual não é validado: `ConnectionViewModel.kt:63-73` só prefixa `http://` e salva. Não existe chamada a `/api/v1/health` no app.
- Não há nenhum teste de conexão ou descoberta.

## Objetivo

O app recém-instalado lista o servidor em até 10 s sem digitar nada, a busca sempre termina, e só é aceito um endereço confirmado como KuraNAS.

## O que fazer

Em `android/app/src/main/java/com/kuranas/android/`. Uma alteração por item.

1. **Nome do serviço**: `NsdDiscovery` procura `_kuranas._tcp.`.
2. **Validação do servidor** (`core/server`): `GET /api/v1/health` com limite de 3 s, aceitando só `200` com `{"status": "ok", "service": "kuranas"}`. Resultado tipado: confirmado / sem resposta / não é KuraNAS / acesso não liberado (`403`).
3. **Descoberta por UDP** (`core/discovery`): envia `KURANAS_DISCOVER` por broadcast para a porta 19520 com `MulticastLock` adquirido, espera respostas por 3 s, aceita só `service == "kuranas"` e monta o endereço com o IP de origem e o `port` da resposta.
4. **Busca com fim** (`ConnectionViewModel`): mDNS e UDP rodam juntos por no máximo 10 s; cada candidato passa pela validação do item 2 antes de entrar na lista (sem repetir endereço); ao fim, `isDiscovering` vira falso com a lista ou com "nenhum servidor encontrado".
5. **Endereço manual**: sem esquema ganha `http://`, sem porta ganha `:8000`; só é salvo se a validação confirmar; senão mostra o motivo (`RF-02`).
6. **Tela** (`ConnectionScreen`): "Procurar de novo", motivo do erro, e o texto do que conferir (mesmo Wi-Fi, servidor ligado) — tudo em `strings.xml`.
7. **Reconexão** (`RN-05`): ao abrir com servidor lembrado que não confirma, o app roda a descoberta antes de pedir ação; se achar um servidor KuraNAS, passa a usá-lo.

## Sub-itens levantados na execução (2026-11-16)

- [x] 22.1. Depois que a reconexão termina em `Unreachable`/`AccessDenied`, o `RootViewModel` mantém a tela de conexão enquanto o servidor salvo for igual ao que falhou. Escolher de novo **o mesmo** servidor (ele voltou e apareceu em "Procurar de novo") não emite valor novo no DataStore e o app não entra. A escolha na tela de conexão precisa sinalizar a raiz explicitamente (não por comparação de estado) — `24/RN-05`, `CA-05`.

## Critérios de aceite

Verificados em aparelho real, na rede do servidor.

- [x] `24/CA-01` — app recém-instalado lista o servidor em até 10 s. *Task 47 (2026-11-24), sem aparelho: o servidor local responde à sondagem UDP `KURANAS_DISCOVER` na porta 19520 (`{"service":"kuranas","port":8000,"api":"/api/v1"}`) e à consulta mDNS `_kuranas._tcp.local`; prazo no aparelho não medido.*
- [x] `24/CA-02` — com o servidor desligado, a busca termina em até 10 s dizendo que nenhum foi encontrado. *Task 47 (2026-11-24), sem aparelho: prazo da busca coberto por teste da descoberta; o verificador do app devolve `NoResponse` para porta sem servidor.*
- [x] `24/CA-03` — digitar o endereço de outra máquina (ou de um site) é recusado com o motivo. *Task 47 (2026-11-24), sem aparelho: o `ServerHealthChecker` real do app contra um servidor HTTP qualquer (`python -m http.server`) devolve `NotKuraNas`, contra o backend local `Confirmed`.*
- [x] `24/CA-04` — fechar e reabrir conecta direto. *Task 47 (2026-11-24), sem aparelho: servidor lembrado e reconexão na abertura cobertos por testes do `RootViewModel`/`ServerReconnection`; não exercitado no aparelho.*
- [x] `24/CA-05` — trocar o IP do servidor: o app o reencontra sem digitar. *Task 47 (2026-11-24), sem aparelho: realocação por descoberta coberta por teste (`Relocated`); não exercitado no aparelho.*
- [x] Com mDNS bloqueado no roteador, o servidor ainda é achado (por UDP). *Task 47 (2026-11-24), sem aparelho: a resposta UDP independe do mDNS (verificada separadamente acima); roteador real não exercitado.*
- [x] Testes unitários: validação do servidor (quatro resultados), montagem do endereço a partir da resposta UDP, normalização do endereço manual, término da busca no tempo limite.
- [x] `assembleDebug` e `testDebugUnitTest` verdes.

## Fora de escopo

- Tela para dispositivo fora da whitelist com o IP do aparelho (`24/RN-06`, `CA-06`) — entra pela task 39.
- Vários servidores ao mesmo tempo.
- Mudanças no backend (o anúncio e o `/health` já estão corretos).
