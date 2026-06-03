# Plugin KuraNAS

Extensão Chrome (Manifest V3) responsável por detectar e capturar mídia no navegador para envio ao KuraNAS.

## Estrutura Atual

```text
plugin/
├── manifest.json
├── background.js               # service worker de composição
├── content/
│   ├── bridge.js
│   ├── blob-interceptor.js
│   └── title-detector.js
├── popup/
│   ├── popup.html
│   ├── popup.css
│   └── popup.js
├── offscreen/
│   ├── recorder.html
│   └── recorder.js
├── icons/
├── src/
│   ├── background/             # módulos de detecção, roteamento, upload/download e estado
│   └── shared/                 # constantes e utilitários compartilhados
└── tests/                      # testes unitários do stack plugin
```

## Setup de Desenvolvimento

Instalar dependências do plugin:

```bash
cd plugin && npm ci
```

## Qualidade

Lint:

```bash
cd plugin && npm run lint
```

Testes:

```bash
cd plugin && npm test
```

## Carregar no Chrome (desenvolvimento manual)

1. Abrir `chrome://extensions`.
2. Ativar `Developer mode`.
3. Clicar em `Load unpacked`.
4. Selecionar a pasta `plugin/`.

## Diretrizes de Arquitetura

- Não misturar nova feature com reorganização estrutural.
- Manter comportamento funcional equivalente durante refactors.
- Preservar ownership por contexto (`background`, `content`, `popup`, `offscreen`, `shared`).
- Garantir consistência entre `manifest.json` e caminhos reais dos scripts.

## i18n

- Evitar hardcode de novos textos visíveis ao usuário em popup/fluxos de UI.
- Introduzir/usar camada de i18n do plugin quando novos textos forem adicionados.

## Padrões de Implementação

Antes de alterar o stack plugin, seguir:

- `/docs/standards/plugin-standards.md`
