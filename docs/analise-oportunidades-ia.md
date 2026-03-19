# Análise de Oportunidades de IA no KuraNAS

## Contexto do escopo atual

Com base na estrutura e documentação do projeto, o KuraNAS já possui fundamentos fortes para adoção de IA:

- catálogo unificado de arquivos e metadados (imagem, áudio, vídeo);
- eventos comportamentais de uso (ex.: playback e interações);
- pipeline assíncrono com jobs/workers para processamento;
- domínio de busca global e analytics operacional já consolidados.

Referências usadas:

- `docs/system-audit.md`
- `docs/system-api-reference.md`
- `backend/internal/api/v1/files/image_classification.go`
- `backend/internal/api/v1/video/playlist/engine.go`
- `backend/internal/api/v1/search/service.go`

## Oportunidades recomendadas (priorizadas)

### 1) Busca semântica global

**Impacto:** Alto  
**Esforço:** Médio

Hoje a busca é principalmente lexical. Com embeddings, o sistema passa a responder por intenção e contexto, por exemplo:

- "foto da praia ao pôr do sol";
- "vídeos da viagem com família";
- "músicas calmas para foco".

Resultado esperado: maior taxa de descoberta e menor dependência de nome exato de arquivo.

### 2) Classificação inteligente de imagens

**Impacto:** Alto  
**Esforço:** Baixo a Médio

O projeto já classifica por heurística (`capture`, `photo`, `other`).  
IA visual pode elevar precisão e abrir novas categorias:

- pessoas;
- documentos;
- recibos/notas;
- screenshots de app;
- paisagem/retrato.

Resultado esperado: biblioteca de imagens mais útil e filtros mais inteligentes.

### 3) Recomendação personalizada para vídeo e música

**Impacto:** Alto  
**Esforço:** Médio

Há dados prontos de comportamento (`started`, `paused`, `completed`, `skipped`, etc.) e engine de playlist.  
IA pode evoluir o ranking automático por perfil de uso.

Resultado esperado: melhor "continue watching/listening" e playlists mais relevantes.

### 4) Deduplicação inteligente (além de checksum)

**Impacto:** Alto  
**Esforço:** Médio

Checksum detecta duplicado exato. IA (ex.: hash perceptual/embeddings) identifica quase duplicados:

- fotos com corte/resolução diferente;
- vídeos reencodados;
- imagens visualmente iguais com bytes diferentes.

Resultado esperado: maior recuperação de espaço com menos esforço manual.

### 5) Assistente em linguagem natural para operações do NAS

**Impacto:** Médio  
**Esforço:** Médio a Alto

Permitir comandos como:

- "mostre pastas que mais cresceram esta semana";
- "encontre vídeos acima de 2 GB";
- "mova arquivos antigos para pasta X" (com confirmação).

Resultado esperado: operação mais rápida para usuários não técnicos e tarefas administrativas recorrentes.

### 6) Normalização automática de metadados de mídia

**Impacto:** Médio  
**Esforço:** Baixo a Médio

IA para corrigir inconsistências em artista/álbum/gênero e agrupar variações de nomes.

Resultado esperado: catálogo musical e de vídeo mais limpo, com menos fragmentação.

### 7) Insights e alertas automáticos de analytics

**Impacto:** Médio  
**Esforço:** Baixo

Gerar explicações e alertas acionáveis:

- crescimento atípico de armazenamento;
- falhas recorrentes de metadata/thumbnail;
- oportunidades de limpeza.

Resultado esperado: observabilidade mais útil e decisão mais rápida.

### 8) Assistente técnico interno para jobs/workers

**Impacto:** Baixo a Médio  
**Esforço:** Baixo

Análise assistida de erros em pipeline para sugerir causa provável e próximos passos operacionais.

Resultado esperado: redução do tempo de diagnóstico.

## Roadmap sugerido (ordem de implementação)

1. Classificação inteligente de imagens.  
2. Busca semântica global.  
3. Recomendações de vídeo/música.  
4. Deduplicação inteligente.  
5. Assistente operacional em linguagem natural.

Essa sequência entrega valor cedo com risco técnico controlado e reaproveita o pipeline já existente.
