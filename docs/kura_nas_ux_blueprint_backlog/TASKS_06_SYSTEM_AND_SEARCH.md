# Tasks 06 - System And Search

## TASK-016 - Separar Analytics em visao geral e biblioteca/indexacao
- Status: `TODO`
- Tamanho: `M`
- Objetivo: alinhar analytics ao blueprint com duas camadas: saude geral e estado da biblioteca.
- Estado atual: ha apenas uma pagina `Analytics` com overview unico.
- Escopo:
- Criar estrutura com visao geral e biblioteca/indexacao.
- Reorganizar cards, tabelas e indicadores em blocos mais previsiveis.
- Expor melhor pendencias de thumbnail, metadata e erros de processamento.
- Criterios de aceite:
- O dominio tecnico fica mais claro para usuario tecnico.
- A pagina deixa de misturar tudo no mesmo nivel.
- Dependencias: `TASK-002`.

## TASK-017 - Criar Settings e consolidar configuracoes
- Status: `TODO`
- Tamanho: `M`
- Objetivo: introduzir o dominio `Configuracoes` no produto e no backend.
- Estado atual: backend expoe apenas `translation` e `about`; nao existe pagina de configuracao.
- Escopo:
- Criar rota e pagina `Settings`.
- Definir contratos iniciais para biblioteca, indexacao, players, aparencia e idioma.
- Garantir compatibilidade com i18n e valores persistidos.
- Criterios de aceite:
- O produto passa a ter uma area real de configuracao.
- O dominio deixa de vazar para About ou para configuracoes hardcoded.
- Dependencias: `TASK-001`, `TASK-002`.

## TASK-018 - Limpar About e retirar Activity Diary da camada principal
- Status: `TODO`
- Tamanho: `S`
- Objetivo: posicionar `About` como pagina tecnica compacta e remover `Activity Diary` da experiencia principal.
- Estado atual: `About` ja existe; `activity-diary` ainda e pagina de primeiro nivel e parte do shell.
- Escopo:
- Simplificar `About` dentro do novo shell.
- Tirar `activity-diary` da navegacao principal.
- Decidir se o diario vira historico tecnico interno ou rota isolada de manutencao.
- Criterios de aceite:
- Navegacao principal nao exibe mais `activity-diary`.
- `About` fica enxuto e coerente com o dominio Sistema.
- Dependencias: `TASK-001`, `TASK-017`.

## TASK-019 - Implementar busca global e remover legado restante
- Status: `TODO`
- Tamanho: `M`
- Objetivo: fechar o blueprint com uma busca/global command palette e eliminar o legado visual e estrutural remanescente.
- Estado atual: `Header.tsx` tem um input visual sem comportamento global.
- Escopo:
- Implementar busca global por arquivos, pastas, artistas, albuns, playlists, videos e imagens.
- Adicionar acoes rapidas de navegacao.
- Remover componentes, rotas e estados legados que sobrarem apos as migracoes anteriores.
- Criterios de aceite:
- A busca global funciona como entrada primaria do produto.
- Nao restam elementos principais do mapa antigo contradizendo o blueprint.
- Dependencias: `TASK-003`, `TASK-006`, `TASK-009`, `TASK-012`, `TASK-015`, `TASK-018`.
