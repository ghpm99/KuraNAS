# KuraNAS

## Descrição

KuraNAS é um sistema NAS (Network Attached Storage) pessoal, projetado para ser simples, fácil de usar e acessível. Ele permite que você armazene, organize e compartilhe seus arquivos de forma segura em sua rede local.

## Funcionalidades

- **Armazenamento centralizado:** Armazene todos os seus arquivos em um único local, acessível de qualquer dispositivo na sua rede.
- **Interface web:** Gerencie seus arquivos através de uma interface web intuitiva e fácil de usar.
- **Compartilhamento de arquivos:** Compartilhe arquivos e pastas com outros usuários na sua rede.
- **Controle de acesso:** Defina permissões de acesso para diferentes usuários e pastas.
- **Upload e download:** Faça upload e download de arquivos facilmente através da interface web.
- **Organização de arquivos:** Organize seus arquivos em pastas e subpastas.
- **Pré-visualização de arquivos:** Visualize imagens, vídeos e documentos diretamente na interface web.
- **Compatibilidade:** Acesse seus arquivos de qualquer dispositivo com um navegador web, incluindo computadores, tablets e smartphones.

## Arquitetura

O KuraNAS é composto por duas partes principais:

- **Frontend:** A interface web, construída com React e TypeScript.
- **Backend:** O servidor, construído com Go.

O frontend se comunica com o backend através de chamadas de API REST. O backend gerencia o armazenamento de arquivos, o controle de acesso e o compartilhamento de arquivos.

## Pré-requisitos

Antes de começar, você precisará ter o seguinte instalado:

- **Go:** (Versão 1.20 ou superior)
- **Node.js:** (Versão 16 ou superior)
- **Yarn:** (Opcional, mas recomendado)

## Instalação

1.  **Clone o repositório:**

    ```bash
    git clone https://github.com/seu-usuario/KuraNAS.git
    cd KuraNAS
    ```

2.  **Construa a aplicação:**

    ```bash
    make
    ```

3.  **Execute o servidor:**

    ```bash
    ./main
    ```

    O servidor será executado na porta 8080 por padrão. Você pode alterar a porta configurando a variável de ambiente `PORT`.

## Configuração

O KuraNAS pode ser configurado através de variáveis de ambiente. As seguintes variáveis de ambiente estão disponíveis:

- `PORT`: A porta em que o servidor será executado (padrão: 8080).
- `DATA_DIR`: O diretório onde os arquivos serão armazenados (padrão: ./data).

## Uso

1.  Abra seu navegador web e acesse `http://localhost:8080` (ou o endereço e porta configurados).
2.  Crie uma conta de usuário.
3.  Comece a fazer upload e organizar seus arquivos.

## Contribuição

Contribuições são bem-vindas! Se você encontrar um bug ou tiver uma sugestão de melhoria, por favor, abra uma issue ou envie um pull request.

## Licença

Este projeto está licenciado sob a licença MIT. Veja o arquivo `LICENSE` para mais informações.

## Autores

- [Guilherme H.](https://github.com/ghpm99)

## Agradecimentos

- Agradecimentos à comunidade open source por fornecer as ferramentas e bibliotecas que tornaram este projeto possível.
