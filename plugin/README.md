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

<h3 align="center">KuraNAS Stream Grabber</h3>

  <p align="center">
    Extensão Chrome (Manifest V3) para detectar e capturar mídia no navegador e enviar ao KuraNAS.
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

Extensão Chrome (Manifest V3) responsável por detectar e capturar mídia no navegador
para envio ao KuraNAS. A arquitetura preserva ownership por contexto
(`background`, `content`, `popup`, `offscreen`, `shared`).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Built With

* [![JavaScript][JavaScript.com]][JavaScript-url]
* [![Chrome][Chrome.com]][Chrome-url]
* [![Node.js][Node.js]][Node-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

Como preparar o ambiente de desenvolvimento do plugin.

### Prerequisites

* Node.js 20+
* npm 10+
* Google Chrome

### Installation

1. Clone o repositório
   ```sh
   git clone https://github.com/ghpm99/KuraNAS.git
   ```
2. Instale as dependências
   ```sh
   cd plugin && npm ci
   ```
3. Carregue no Chrome (desenvolvimento manual)
   1. Abrir `chrome://extensions`.
   2. Ativar `Developer mode`.
   3. Clicar em `Load unpacked`.
   4. Selecionar a pasta `plugin/`.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

### Estrutura

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

### Qualidade

Lint:

```bash
cd plugin && npm run lint
```

Testes:

```bash
cd plugin && npm test
```

### i18n

- Evitar hardcode de novos textos visíveis ao usuário em popup/fluxos de UI.
- Introduzir/usar camada de i18n do plugin quando novos textos forem adicionados.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [ ] Camada de i18n para textos do popup
- [ ] Cobertura de testes dos módulos de `background`
- [ ] Empacotamento para a Chrome Web Store

See the [open issues](https://github.com/ghpm99/KuraNAS/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

Antes de alterar o stack plugin, siga `docs/standards/plugin-standards.md`. Diretrizes de arquitetura:
não misturar nova feature com reorganização estrutural; manter comportamento funcional equivalente
durante refactors; preservar ownership por contexto; garantir consistência entre `manifest.json`
e os caminhos reais dos scripts.

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

* [Chrome Extensions (MV3)](https://developer.chrome.com/docs/extensions/mv3/intro/)
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
[JavaScript.com]: https://img.shields.io/badge/JavaScript-F7DF1E?style=for-the-badge&logo=javascript&logoColor=black
[JavaScript-url]: https://developer.mozilla.org/en-US/docs/Web/JavaScript
[Chrome.com]: https://img.shields.io/badge/Chrome-4285F4?style=for-the-badge&logo=googlechrome&logoColor=white
[Chrome-url]: https://developer.chrome.com/docs/extensions/
[Node.js]: https://img.shields.io/badge/Node.js-339933?style=for-the-badge&logo=nodedotjs&logoColor=white
[Node-url]: https://nodejs.org/
