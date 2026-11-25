export function normalizeApiBaseUrl(url) {
  if (!url || typeof url !== "string") return "";

  const trimmed = url.trim().replace(/\/+$/g, "");
  if (!trimmed) return "";

  if (/\/api\/v1$/i.test(trimmed)) {
    return trimmed;
  }

  return `${trimmed}/api/v1`;
}

export const CONNECTION_STATUS = Object.freeze({
  CONNECTED: "connected",
  NOT_FOUND: "not_found",
  FORBIDDEN: "forbidden",
});

export function createConnectionStatusChecker({ healthSuffix, timeoutMs, fetchImpl = fetch }) {
  async function fetchWithTimeout(url, options) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

    try {
      return await fetchImpl(url, { ...options, signal: controller.signal });
    } finally {
      clearTimeout(timeoutId);
    }
  }

  return async function checkConnectionStatus(apiBaseUrl) {
    try {
      const response = await fetchWithTimeout(`${apiBaseUrl}${healthSuffix}`, {
        method: "GET",
        cache: "no-store",
      });

      if (response.status === 403) {
        return CONNECTION_STATUS.FORBIDDEN;
      }

      if (!response.ok) {
        return CONNECTION_STATUS.NOT_FOUND;
      }

      const healthBody = (await response.text()).toLowerCase();
      return healthBody.includes("kuranas") ? CONNECTION_STATUS.CONNECTED : CONNECTION_STATUS.NOT_FOUND;
    } catch {
      return CONNECTION_STATUS.NOT_FOUND;
    }
  };
}

export function createReachabilityChecker(options) {
  const checkConnectionStatus = createConnectionStatusChecker(options);

  return async function isKuraNASApiReachable(apiBaseUrl) {
    return (await checkConnectionStatus(apiBaseUrl)) === CONNECTION_STATUS.CONNECTED;
  };
}

function isHttpUrl(address) {
  try {
    const { protocol } = new URL(address);
    return protocol === "http:" || protocol === "https:";
  } catch {
    return false;
  }
}

export function createServerAddressSaver({ isKuraNASApiReachable, saveApiBaseUrl }) {
  return async function saveVerifiedServerAddress(typedAddress) {
    const trimmedAddress = typeof typedAddress === "string" ? typedAddress.trim() : "";
    if (!trimmedAddress) {
      await saveApiBaseUrl("");
      return { ok: true, apiBaseUrl: "" };
    }

    const apiBaseUrl = normalizeApiBaseUrl(trimmedAddress);
    if (!isHttpUrl(apiBaseUrl)) {
      return { ok: false, reason: "invalid_address" };
    }

    if (!(await isKuraNASApiReachable(apiBaseUrl))) {
      return { ok: false, reason: "unreachable" };
    }

    await saveApiBaseUrl(apiBaseUrl);
    return { ok: true, apiBaseUrl };
  };
}
