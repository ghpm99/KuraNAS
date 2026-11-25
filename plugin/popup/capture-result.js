(function (root) {
  "use strict";

  const SENDING_RESULT = { text: "Enviando...", kind: "progress" };

  function describeCaptureResponse(response) {
    if (!response) {
      return { text: "Sem resposta da extensão.", kind: "error" };
    }
    if (response.error) {
      return { text: String(response.error), kind: "error" };
    }
    if (response.type === "master" || response.type === "dash_manifest") {
      return { text: "Escolha a qualidade.", kind: "progress" };
    }
    if (response.ok) {
      const suffix = response.name ? `: ${response.name}` : "";
      return { text: `Concluído${suffix}`, kind: "success" };
    }
    return { text: "Resposta inesperada da extensão.", kind: "error" };
  }

  function describeCaptureFailure(error) {
    const message = error && error.message ? error.message : "falha desconhecida";
    return { text: message, kind: "error" };
  }

  const api = { describeCaptureResponse, describeCaptureFailure, SENDING_RESULT };
  if (typeof module !== "undefined" && module.exports) {
    module.exports = api;
  }
  root.KuraNasCaptureResult = api;
})(typeof globalThis !== "undefined" ? globalThis : this);
