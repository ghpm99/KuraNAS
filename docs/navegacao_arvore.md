# Documentação de Produto

# Navegação em Árvore de Arquivos

## 1. Objetivo da funcionalidade

A funcionalidade de **navegação em árvore de arquivos** permite ao usuário explorar o sistema de arquivos de forma similar ao **Windows Explorer, Finder ou Dropbox**, possibilitando:

* navegar entre diretórios
* abrir arquivos
* acessar caminhos diretamente via URL
* compartilhar links para arquivos ou pastas
* utilizar navegação do navegador (voltar/avançar)

Essa funcionalidade deve ser **determinística, previsível e compartilhável via URL**.

---

# 2. Conceitos fundamentais

## 2.1 Path

O **path** representa a localização de um arquivo ou pasta dentro da árvore.

Exemplos:

```
/
```

```
/photos
```

```
/photos/trip_2024
```

```
/photos/trip_2024/image01.jpg
```

Regras:

* `/` representa a raiz
* caminhos são **case sensitive**
* não deve existir `//`
* não deve existir `/./` ou `/../`

---

## 2.2 Tipos de itens

Existem dois tipos de itens:

### Diretório (Folder)

* contém outros arquivos ou pastas
* ao acessar, lista seu conteúdo

### Arquivo (File)

* representa um conteúdo
* ao acessar, abre o visualizador apropriado

---

# 3. Estrutura de URL

A URL deve refletir **exatamente o path atual**.

## Estrutura base

```
/files/<path>
```

Exemplo:

```
/files/
```

```
/files/photos
```

```
/files/photos/trip_2024
```

```
/files/photos/trip_2024/image01.jpg
```

---

## Comportamento esperado

A URL deve ser **a única fonte de verdade da navegação**.

Isso significa:

* abrir uma URL deve restaurar o estado da navegação
* compartilhar a URL deve abrir o mesmo conteúdo
* atualizar a URL deve mudar a navegação

---

# 4. Estados da interface

Existem dois estados principais.

---

# 4.1 Visualização de diretório

Quando o path aponta para uma pasta.

Exemplo:

```
/files/photos/trip_2024
```

A interface deve mostrar:

* lista de arquivos
* lista de subpastas
* breadcrumbs
* árvore lateral

---

# 4.2 Visualização de arquivo

Quando o path aponta para um arquivo.

Exemplo:

```
/files/photos/trip_2024/image01.jpg
```

A interface deve:

* abrir o visualizador do arquivo
* permitir voltar para o diretório pai

---

# 5. Navegação

## 5.1 Abrir pasta

Quando o usuário clica em uma pasta:

1. a URL deve ser atualizada
2. o conteúdo da pasta deve ser carregado
3. o histórico do navegador deve registrar a navegação

Exemplo:

```
/files/photos
→ usuário clica trip_2024
→ /files/photos/trip_2024
```

---

### Critérios de aceite

* clicar em uma pasta atualiza a URL
* clicar em uma pasta adiciona entrada no histórico
* atualizar a página mantém o diretório aberto
* o conteúdo exibido corresponde ao path

---

# 5.2 Abrir arquivo

Quando o usuário clica em um arquivo.

Fluxo:

1. a URL deve mudar
2. o visualizador deve abrir
3. o histórico deve registrar

Exemplo:

```
/files/photos/trip_2024
→ usuário clica image01.jpg
→ /files/photos/trip_2024/image01.jpg
```

---

### Critérios de aceite

* clicar em um arquivo atualiza a URL
* o arquivo correto é exibido
* refresh mantém o arquivo aberto
* botão voltar retorna para a pasta

---

# 5.3 Botão voltar do navegador

O navegador deve funcionar corretamente.

Exemplo:

```
/files/photos
→ /files/photos/trip_2024
→ /files/photos/trip_2024/image01.jpg
```

Pressionando voltar:

```
/files/photos/trip_2024
```

Pressionando novamente:

```
/files/photos
```

---

### Critérios de aceite

* botão voltar retorna ao path anterior
* botão avançar reaplica a navegação
* estado da interface corresponde à URL

---

# 5.4 Breadcrumbs

Os breadcrumbs representam o path atual.

Exemplo:

```
Home / photos / trip_2024 / image01.jpg
```

Cada segmento deve ser clicável.

---

### Critérios de aceite

* cada breadcrumb navega para seu nível
* clicar em breadcrumb atualiza a URL
* breadcrumb reflete exatamente o path

---

# 5.5 Acesso direto via URL

Usuário pode abrir qualquer path diretamente.

Exemplo:

```
/files/photos/trip_2024/image01.jpg
```

---

### Critérios de aceite

* abrir URL direta carrega o item correto
* pasta abre listagem
* arquivo abre visualizador
* erros retornam estado apropriado

---

# 6. Estados de erro

---

## 6.1 Arquivo inexistente

Exemplo:

```
/files/photos/nao_existe.jpg
```

---

### Comportamento

Exibir erro:

```
File not found
```

ou

```
This file does not exist
```

---

### Critérios de aceite

* erro claro ao usuário
* URL permanece intacta
* usuário pode voltar

---

## 6.2 Pasta inexistente

Mesmo comportamento.

---

# 7. Navegação relativa

Sistema deve permitir navegar para diretório pai.

Exemplo:

```
/files/photos/trip_2024
```

botão voltar diretório:

```
/files/photos
```

---

### Critérios de aceite

* botão “voltar pasta” funciona
* não quebra na raiz

---

# 8. Atualização de estado

A aplicação deve ser **state driven pela URL**.

Regra fundamental:

```
URL → estado da aplicação
```

Não:

```
estado → URL opcional
```

---

### Critérios de aceite

* alterar URL manualmente atualiza a UI
* refresh não perde estado
* copiar URL mantém contexto

---

# 9. Performance

Mudanças de pasta devem ser rápidas.

---

### Critérios de aceite

* navegação não recarrega página inteira
* somente dados necessários são carregados
* UI mostra loading quando necessário

---

# 10. Comportamento de refresh

Atualizar a página deve manter estado.

Exemplo:

```
/files/photos/trip_2024/image01.jpg
```

Após refresh:

* arquivo continua aberto

---

### Critérios de aceite

* refresh preserva estado
* não retorna para raiz

---

# 11. Deep linking

Usuário deve poder compartilhar URLs.

Exemplo:

```
/files/photos/trip_2024/image01.jpg
```

Outro usuário acessa e vê o mesmo arquivo.

---

### Critérios de aceite

* URL reproduz estado completo
* nenhum estado depende apenas do frontend

---

# 12. Sincronização com árvore lateral

Ao navegar:

* árvore deve expandir diretórios relevantes
* item atual deve ser destacado

---

### Critérios de aceite

* árvore reflete path atual
* diretórios pais ficam expandidos
* item atual fica selecionado

---

# 13. Estados de carregamento

Durante navegação.

---

### Critérios de aceite

* loading ao mudar de pasta
* loading ao abrir arquivo grande
* não bloquear UI completamente

---

# 14. Regras de UX obrigatórias

A navegação deve ser:

* previsível
* compartilhável
* reversível
* consistente com navegador

---

# 15. Anti-patterns proibidos

Não pode:

* mudar pasta sem atualizar URL
* abrir arquivo sem atualizar URL
* resetar estado ao refresh
* usar estado interno como fonte de verdade

---

# 16. Critério global de aceite da feature

A funcionalidade é considerada completa quando:

* qualquer arquivo pode ser acessado via URL
* qualquer pasta pode ser acessada via URL
* navegação funciona com histórico do navegador
* refresh mantém estado
* breadcrumbs refletem path
* árvore lateral reflete path
* links podem ser compartilhados

---

# 17. Testes obrigatórios

Deve existir testes para:

### navegação

* abrir pasta
* abrir arquivo

### histórico

* voltar
* avançar

### URL direta

* abrir pasta via URL
* abrir arquivo via URL

### erro

* arquivo inexistente
* pasta inexistente