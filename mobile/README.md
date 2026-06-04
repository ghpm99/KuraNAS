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

<h3 align="center">KuraNAS Mobile</h3>

  <p align="center">
    App Android nativo do KuraNAS focado em compatibilidade com Android 4.1.2 (API 16).
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

Aplicativo Android nativo do KuraNAS, projetado para rodar em hardware legado.

### Restrições Obrigatórias

- Dispositivo alvo: Samsung Galaxy Tab 2 7.0 (GT-P3110), 1024x600.
- Stack obrigatória: **Java + XML Views + AppCompat**.
- Não usar Kotlin.
- Não usar Jetpack Compose.
- Toda decisão de API/lib deve ser compatível com `minSdk 16`.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Built With

* [![Java][Java.com]][Java-url]
* [![Android][Android.com]][Android-url]
* [![Gradle][Gradle.org]][Gradle-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

Como compilar e instalar o app Android.

### Prerequisites

* JDK 17
* Android SDK com `compileSdk 35` instalado
* `ANDROID_HOME`/`ANDROID_SDK_ROOT` configurado

### Installation

1. Clone o repositório
   ```sh
   git clone https://github.com/ghpm99/KuraNAS.git
   ```
2. Gere o APK de debug
   ```sh
   cd mobile && ./gradlew assembleDebug
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

### Estrutura

```text
mobile/
├── app/
│   ├── src/main/java/com/kuranas/mobile/
│   │   ├── app/                    # Application, Activity raiz, ServiceLocator
│   │   ├── data/                   # implementações de repository + mappers
│   │   ├── domain/                 # entidades, portas e contratos
│   │   ├── feature/                # ownership incremental por domínio (files/images/search/settings)
│   │   ├── i18n/                   # TranslationManager
│   │   ├── infra/                  # HTTP, cache, discovery, logging, preferences
│   │   └── presentation/           # Fragments/Activities legados e comuns
│   ├── src/main/res/               # layouts XML, drawables e resources
│   └── src/main/assets/translations/
├── gradle/
├── build.gradle
└── settings.gradle
```

### Comandos

| Comando | Descrição |
| --- | --- |
| `./gradlew assembleDebug` | Build debug |
| `./gradlew assembleRelease` | Build release |
| `./gradlew test` | Testes unitários |
| `./gradlew connectedAndroidTest` | Testes instrumentados |

### i18n

- Não hardcode texto visível ao usuário em Java/XML novo ou alterado.
- Traduções locais de fallback ficam em `app/src/main/assets/translations`.
- Traduções remotas são carregadas via `/api/v1/configuration/translation` por `TranslationManager`.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [ ] Ampliar ownership por feature (`feature/<domain>/{presentation,domain,data}`)
- [ ] Cobertura de testes unitários por domínio
- [ ] Validar layouts no form factor 7" (1024×600)

See the [open issues](https://github.com/ghpm99/KuraNAS/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

Antes de alterar o stack mobile, siga `docs/standards/mobile-standards.md`. Pontos obrigatórios:
preservar API 16; manter Java/XML/AppCompat; manter separação entre apresentação, domínio e dados
(incluindo `feature/<domain>/{presentation,domain,data}` onde já adotado); rodar `./gradlew test`
e `./gradlew assembleDebug` nas alterações.

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

* [Android Developers](https://developer.android.com/)
* [AppCompat](https://developer.android.com/jetpack/androidx/releases/appcompat)
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
[Java.com]: https://img.shields.io/badge/Java-ED8B00?style=for-the-badge&logo=openjdk&logoColor=white
[Java-url]: https://www.java.com/
[Android.com]: https://img.shields.io/badge/Android-3DDC84?style=for-the-badge&logo=android&logoColor=white
[Android-url]: https://developer.android.com/
[Gradle.org]: https://img.shields.io/badge/Gradle-02303A?style=for-the-badge&logo=gradle&logoColor=white
[Gradle-url]: https://gradle.org/
