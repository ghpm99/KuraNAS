/* ========================================================================
 * KuraNAS Stream Grabber — Server-side fetch
 * Asks the KuraNAS server to pull a URL (yt-dlp) straight into the library,
 * instead of capturing it through the browser. Talks to the /ingest API.
 * ======================================================================== */

export function createFetcher({ getApiBaseUrl, fetchImpl = fetch }) {
  async function getJSON(path) {
    const apiUrl = await getApiBaseUrl();
    const resp = await fetchImpl(`${apiUrl}${path}`);
    if (!resp.ok) {
      throw new Error(`HTTP ${resp.status}`);
    }
    return resp.json();
  }

  async function submitFetch({ url, preset, targetRoot, subfolder }) {
    try {
      const apiUrl = await getApiBaseUrl();
      const resp = await fetchImpl(`${apiUrl}/ingest/fetch`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          url,
          preset,
          target_root: targetRoot,
          subfolder: subfolder || "",
        }),
      });
      const data = await resp.json().catch(() => ({}));
      if (!resp.ok) {
        return { ok: false, error: data.error || `HTTP ${resp.status}` };
      }
      return { ok: true, jobId: data.job_id };
    } catch (error) {
      return { ok: false, error: error.message };
    }
  }

  async function listTargets() {
    try {
      return { ok: true, targets: await getJSON("/ingest/targets") };
    } catch (error) {
      return { ok: false, error: error.message, targets: [] };
    }
  }

  async function listPresets() {
    try {
      return { ok: true, presets: await getJSON("/ingest/presets") };
    } catch (error) {
      return { ok: false, error: error.message, presets: [] };
    }
  }

  return { submitFetch, listTargets, listPresets };
}
