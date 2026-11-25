export function normalizeApiBaseUrl(url) {
  if (!url || typeof url !== "string") return "";

  const trimmed = url.trim().replace(/\/+$/g, "");
  if (!trimmed) return "";

  if (/\/api\/v1$/i.test(trimmed)) {
    return trimmed;
  }

  return `${trimmed}/api/v1`;
}

export function createReachabilityChecker({ healthSuffix, timeoutMs, fetchImpl = fetch }) {
  async function fetchWithTimeout(url, options) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

    try {
      return await fetchImpl(url, { ...options, signal: controller.signal });
    } finally {
      clearTimeout(timeoutId);
    }
  }

  return async function isKuraNASApiReachable(apiBaseUrl) {
    try {
      const response = await fetchWithTimeout(`${apiBaseUrl}${healthSuffix}`, {
        method: "GET",
        cache: "no-store",
      });

      if (!response.ok) {
        return false;
      }

      const healthBody = (await response.text()).toLowerCase();
      return healthBody.includes("kuranas");
    } catch {
      return false;
    }
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
