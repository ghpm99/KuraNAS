<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>
<!--
*** Thanks for checking out the Best-README-Template. If you have a suggestion
*** that would make this better, please fork the repo and create a pull request
*** or simply open an issue with the tag "enhancement".
*** Don't forget to give the project a star!
*** Thanks again! Now go create something AMAZING! :D
-->



<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![project_license][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/ghpm99/KuraNAS">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a>

<h3 align="center">KuraNAS Backend</h3>

  <p align="center">
    Serviço backend do KuraNAS em Go: API HTTP, persistência, processamento assíncrono e internacionalização.
    <br />
    <a href="https://github.com/ghpm99/KuraNAS"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/ghpm99/KuraNAS">View Demo</a>
    &middot;
    <a href="https://github.com/ghpm99/KuraNAS/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    &middot;
    <a href="https://github.com/ghpm99/KuraNAS/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

Serviço backend do KuraNAS, implementado em Go, responsável por API HTTP, persistência,
processamento assíncrono (workers) e internacionalização. Expõe a API REST em `/api/v1`
e também serve o frontend como SPA no build integrado.

Arquitetura em camadas por domínio: **Handler → Service → Repository**. Regras de negócio
ficam no service e o SQL no repository, sem bypass de camadas.

Domínios da API (`/api/v1`): `files`, `music`, `video`, `analytics`, `diary`,
`configuration`, `jobs`, `search`, `notifications`, `update`.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Built With

* [![Go][Go.dev]][Go-url]
* [![Gin][Gin.com]][Gin-url]
* [![PostgreSQL][Postgres.org]][Postgres-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

Como subir o backend localmente em modo de desenvolvimento.

### Prerequisites

* Go 1.24+
* PostgreSQL
* Make

### Installation

1. Clone o repositório
   ```sh
   git clone https://github.com/ghpm99/KuraNAS.git
   ```
2. Configure o `backend/.env` (veja a seção [Usage](#usage) para todas as variáveis)
   ```dotenv
   ENTRY_POINT=/mnt/storage
   LANGUAGE=pt-BR
   ENABLE_WORKERS=true
   ENV=dev
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=kuranas
   DB_PASSWORD=secret
   DB_NAME=kuranas
   ```
3. Rode em modo `dev` (tag `dev`, porta `8000`)
   ```sh
   make -C backend run
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

### Estrutura

```text
backend/
├── cmd/nas/                         # entrypoints (`main.go`, `main_windows.go`)
├── internal/
│   ├── api/v1/                      # handlers HTTP por domínio
│   ├── app/                         # bootstrap e rotas
│   ├── config/                      # carga de configuração e ambiente
│   └── worker/                      # orquestração de workers
├── pkg/
│   ├── database/
│   │   ├── migrations/              # migrations e registro
│   │   └── queries/                 # SQL por domínio
│   ├── i18n/                        # loader e resolução de traduções
│   ├── logger/                      # logging
│   └── utils/                       # utilitários compartilhados
├── tests/                           # suites de teste adicionais
├── translations/                    # arquivos JSON de tradução
└── Makefile
```

### Execução e Build

Modo desenvolvimento (tag `dev`, porta `8000`):

```bash
make -C backend run
```

Build backend:

```bash
make -C backend build
```

### Testes

Testes com tag `dev` (suite do backend):

```bash
make -C backend test
```

Cobertura via Makefile:

```bash
make -C backend coverage
```

Cobertura geral recomendada:

```bash
cd backend && go test ./... -cover
```

### Configuração de Ambiente

O backend tenta carregar variáveis de um arquivo `.env` e, se não encontrar, usa o ambiente do sistema.

Caminho esperado do `.env` por build:

- `dev`: `backend/.env`
- Linux (release): `/etc/kuranas/.env`
- Windows (release): `%ProgramFiles%/Kuranas/.env`

Atualmente o projeto não possui `backend/.env.example`. Use a tabela abaixo como referência oficial.

#### Variáveis da aplicação

| Variável | Obrigatória | Padrão | Observações |
| --- | --- | --- | --- |
| `ENTRY_POINT` | Sim | - | Diretório raiz monitorado pelo NAS. |
| `LANGUAGE` | Sim | - | Idioma base (ex.: `pt-BR`, `en-US`). |
| `ENABLE_WORKERS` | Não | `false` | Ativa workers em background quando `true`. |
| `ENV` | Não | vazio | Nome do ambiente (`dev`, `test`, `prod` etc.). |
| `DB_HOST` | Sim | - | Host do PostgreSQL. |
| `DB_PORT` | Sim | - | Porta do PostgreSQL (ex.: `5432`). |
| `DB_USER` | Sim | - | Usuário do banco. |
| `DB_PASSWORD` | Sim | - | Senha do banco. |
| `DB_NAME` | Sim | - | Nome do banco. |

#### Variáveis de workers (opcionais)

| Variável | Padrão | Observações |
| --- | --- | --- |
| `WORKER_CONCURRENCY_CHECKSUM` | `3` | Concorrência para jobs de checksum. |
| `WORKER_CONCURRENCY_METADATA` | `3` | Concorrência para extração de metadados. |
| `WORKER_CONCURRENCY_THUMBNAIL` | `2` | Concorrência para thumbnails. |
| `WORKER_RETRY_BACKOFF_MS` | `500` | Backoff de retry em milissegundos. |
| `WORKER_SCHEDULER_POLL_MS` | `2000` | Intervalo do scheduler em milissegundos. |
| `WORKER_MAX_CONCURRENT_JOBS` | `4` | Limite total de jobs concorrentes. |

#### Variáveis de IA (opcionais)

Se nenhuma chave de IA for definida, o serviço de IA é desativado automaticamente.

| Variável | Padrão | Observações |
| --- | --- | --- |
| `AI_OPENAI_API_KEY` | vazio | Chave da OpenAI. |
| `AI_OPENAI_MODEL` | `gpt-4o-mini` | Modelo padrão OpenAI. |
| `AI_OPENAI_BASE_URL` | `https://api.openai.com/v1` | URL base da OpenAI. |
| `AI_ANTHROPIC_API_KEY` | vazio | Chave da Anthropic. |
| `AI_ANTHROPIC_MODEL` | `claude-sonnet-4-20250514` | Modelo padrão Anthropic. |
| `AI_TIMEOUT_SECONDS` | `30` | Timeout das chamadas de IA. |
| `AI_MAX_RETRIES` | `2` | Número de tentativas por chamada. |
| `AI_RETRY_BACKOFF_MS` | `500` | Backoff entre retries. |

### Banco, SQL e Migrations

- Queries SQL: `pkg/database/queries/<feature>`
- Migrations: `pkg/database/migrations/queries`
- Registro de migrations: `pkg/database/migrations/migrations.go`

Não alterar migrations antigas de forma incompatível; criar nova migration para mudanças de schema.

### Internacionalização (Obrigatória)

- Não hardcode mensagens visíveis ao usuário.
- Toda nova chave de texto deve ser adicionada em `backend/translations`.
- O frontend consome as mesmas chaves via endpoint de configuração de tradução.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [ ] `backend/.env.example` versionado
- [ ] Documentação OpenAPI dos domínios `/api/v1`
- [ ] Ampliar cobertura de testes por domínio

See the [open issues](https://github.com/ghpm99/KuraNAS/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

Antes de alterar código backend, siga `docs/standards/backend-standards.md`. Pontos obrigatórios:
fluxo em camadas `Handler → Service → Repository`; regras de negócio no service e SQL no repository; sem bypass de camadas.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feat/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat: add some AmazingFeature'`)
4. Push to the Branch (`git push origin feat/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Top contributors:

<a href="https://github.com/ghpm99/KuraNAS/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=ghpm99/KuraNAS" alt="contrib.rocks image" />
</a>



<!-- LICENSE -->
## License

Nenhum arquivo de licença é distribuído com o projeto até o momento. Defina uma licença em `LICENSE.txt` na raiz do repositório.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTACT -->
## Contact

ghpm99 - ghpm99@gmail.com

Project Link: [https://github.com/ghpm99/KuraNAS](https://github.com/ghpm99/KuraNAS)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* [Gin Web Framework](https://gin-gonic.com/)
* [Best-README-Template](https://github.com/othneildrew/Best-README-Template)
* [Img Shields](https://shields.io)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/ghpm99/KuraNAS.svg?style=for-the-badge
[contributors-url]: https://github.com/ghpm99/KuraNAS/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/ghpm99/KuraNAS.svg?style=for-the-badge
[forks-url]: https://github.com/ghpm99/KuraNAS/network/members
[stars-shield]: https://img.shields.io/github/stars/ghpm99/KuraNAS.svg?style=for-the-badge
[stars-url]: https://github.com/ghpm99/KuraNAS/stargazers
[issues-shield]: https://img.shields.io/github/issues/ghpm99/KuraNAS.svg?style=for-the-badge
[issues-url]: https://github.com/ghpm99/KuraNAS/issues
[license-shield]: https://img.shields.io/github/license/ghpm99/KuraNAS.svg?style=for-the-badge
[license-url]: https://github.com/ghpm99/KuraNAS/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/linkedin_username
[product-screenshot]: images/screenshot.png
<!-- Shields.io badges. You can a comprehensive list with many more badges at: https://github.com/inttter/md-badges -->
[Go.dev]: https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Go-url]: https://go.dev/
[Gin.com]: https://img.shields.io/badge/Gin-008ECF?style=for-the-badge&logo=gin&logoColor=white
[Gin-url]: https://gin-gonic.com/
[Postgres.org]: https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white
[Postgres-url]: https://www.postgresql.org/
