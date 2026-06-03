export function createDownloader({
  fetchImpl = fetch,
  resolveUrl,
  uploadToKuraNAS,
}) {
  async function downloadHLS(manifestUrl, name) {
    try {
      const resp = await fetchImpl(manifestUrl);
      const text = await resp.text();

      if (text.includes("#EXT-X-STREAM-INF")) {
        return parseHLSMasterPlaylist(text, manifestUrl);
      }

      return await downloadHLSMediaPlaylist(text, manifestUrl, name);
    } catch (error) {
      return { error: error.message };
    }
  }

  function parseHLSMasterPlaylist(text, baseUrl) {
    const lines = text.split("\n");
    const variants = [];

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line.startsWith("#EXT-X-STREAM-INF:")) continue;

      const attrs = line.substring(18);
      const bandwidthMatch = attrs.match(/BANDWIDTH=(\d+)/);
      const resolutionMatch = attrs.match(/RESOLUTION=([^\s,]+)/);
      const codecsMatch = attrs.match(/CODECS="([^"]+)"/);

      const nextLine = (lines[i + 1] || "").trim();
      if (!nextLine || nextLine.startsWith("#")) continue;

      variants.push({
        url: resolveUrl(baseUrl, nextLine),
        bandwidth: bandwidthMatch ? parseInt(bandwidthMatch[1], 10) : 0,
        resolution: resolutionMatch ? resolutionMatch[1] : "",
        codecs: codecsMatch ? codecsMatch[1] : "",
      });
      i++;
    }

    return { type: "master", variants };
  }

  async function downloadHLSMediaPlaylist(text, baseUrl, name) {
    const lines = text.split("\n");
    const segmentUrls = [];

    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith("#")) continue;
      segmentUrls.push(resolveUrl(baseUrl, trimmed));
    }

    const chunks = [];
    let totalSize = 0;

    for (const url of segmentUrls) {
      const resp = await fetchImpl(url);
      const buf = await resp.arrayBuffer();
      chunks.push(new Uint8Array(buf));
      totalSize += buf.byteLength;
    }

    const merged = new Uint8Array(totalSize);
    let offset = 0;
    for (const chunk of chunks) {
      merged.set(chunk, offset);
      offset += chunk.byteLength;
    }

    const blob = new Blob([merged], { type: "video/mp2t" });
    const captureName = name || `stream_hls_${Date.now()}`;

    await uploadToKuraNAS(blob, captureName, "hls");
    return { ok: true, name: captureName };
  }

  async function downloadDASH(manifestUrl, name) {
    try {
      const resp = await fetchImpl(manifestUrl);
      const text = await resp.text();

      const parser = new DOMParser();
      const doc = parser.parseFromString(text, "application/xml");
      const representations = [];

      doc.querySelectorAll("Representation").forEach((rep) => {
        const adaptationSet = rep.closest("AdaptationSet");
        const mimeType = rep.getAttribute("mimeType") ||
          (adaptationSet ? adaptationSet.getAttribute("mimeType") : "") || "";

        representations.push({
          id: rep.getAttribute("id") || "",
          bandwidth: parseInt(rep.getAttribute("bandwidth") || "0", 10),
          width: parseInt(rep.getAttribute("width") || "0", 10),
          height: parseInt(rep.getAttribute("height") || "0", 10),
          codecs: rep.getAttribute("codecs") || "",
          mimeType,
          manifestUrl,
        });
      });

      if (representations.length > 1) {
        return { type: "dash_manifest", representations };
      }

      if (representations.length === 1) {
        return await downloadDASHRepresentation(
          manifestUrl,
          text,
          representations[0].id,
          name
        );
      }

      return { error: "No representations found" };
    } catch (error) {
      return { error: error.message };
    }
  }

  async function downloadDASHRepresentation(manifestUrl, manifestText, repId, name) {
    const parser = new DOMParser();
    const doc = parser.parseFromString(manifestText, "application/xml");
    const rep = repId
      ? doc.querySelector(`Representation[id="${repId}"]`)
      : doc.querySelector("Representation");

    if (!rep) return { error: "Representation not found" };

    const segmentUrls = collectDASHSegmentUrls(rep, manifestUrl);

    const chunks = [];
    let totalSize = 0;

    for (const url of segmentUrls) {
      const resp = await fetchImpl(url);
      const buf = await resp.arrayBuffer();
      chunks.push(new Uint8Array(buf));
      totalSize += buf.byteLength;
    }

    const merged = new Uint8Array(totalSize);
    let offset = 0;
    for (const chunk of chunks) {
      merged.set(chunk, offset);
      offset += chunk.byteLength;
    }

    const mimeType = rep.getAttribute("mimeType") || "video/mp4";
    const ext = mimeType.includes("audio") ? "m4a" : "mp4";
    const blob = new Blob([merged], { type: mimeType });
    const captureName = name || `stream_dash_${Date.now()}`;

    await uploadToKuraNAS(blob, captureName, "dash");
    return { ok: true, name: captureName, ext };
  }

  function collectDASHSegmentUrls(rep, manifestUrl) {
    const urls = [];
    const adaptationSet = rep.closest("AdaptationSet");
    const period = rep.closest("Period");

    const segTemplate =
      rep.querySelector("SegmentTemplate") ||
      (adaptationSet ? adaptationSet.querySelector("SegmentTemplate") : null);

    if (segTemplate) {
      const timeline = segTemplate.querySelector("SegmentTimeline");
      const initTemplate = segTemplate.getAttribute("initialization") || "";
      const mediaTemplate = segTemplate.getAttribute("media") || "";
      const startNumber = parseInt(
        segTemplate.getAttribute("startNumber") || "1",
        10
      );
      const timescale = parseInt(
        segTemplate.getAttribute("timescale") || "1",
        10
      );
      const repId = rep.getAttribute("id") || "";
      const bandwidth = rep.getAttribute("bandwidth") || "";

      if (initTemplate) {
        urls.push(
          resolveUrl(
            manifestUrl,
            expandDASHTemplate(initTemplate, repId, bandwidth, 0, 0)
          )
        );
      }

      if (timeline) {
        let number = startNumber;
        let time = 0;
        const entries = timeline.querySelectorAll("S");

        for (const segment of entries) {
          const t = segment.getAttribute("t");
          if (t !== null) time = parseInt(t, 10);
          const d = parseInt(segment.getAttribute("d") || "0", 10);
          const r = parseInt(segment.getAttribute("r") || "0", 10);

          for (let i = 0; i <= r; i++) {
            urls.push(
              resolveUrl(
                manifestUrl,
                expandDASHTemplate(mediaTemplate, repId, bandwidth, number, time)
              )
            );
            number++;
            time += d;
          }
        }
      } else {
        const duration = parseFloat(
          segTemplate.getAttribute("duration") || "0"
        );
        const periodDuration = parseDuration(
          (period ? period.getAttribute("duration") : null) || ""
        );
        if (duration > 0 && periodDuration > 0) {
          const segCount = Math.ceil(
            (periodDuration * timescale) / duration
          );
          for (let i = 0; i < segCount; i++) {
            urls.push(
              resolveUrl(
                manifestUrl,
                expandDASHTemplate(
                  mediaTemplate,
                  repId,
                  bandwidth,
                  startNumber + i,
                  i * duration
                )
              )
            );
          }
        }
      }
    } else {
      const segList =
        rep.querySelector("SegmentList") ||
        (adaptationSet ? adaptationSet.querySelector("SegmentList") : null);

      if (segList) {
        const init = segList.querySelector("Initialization");
        if (init) {
          urls.push(resolveUrl(manifestUrl, init.getAttribute("sourceURL")));
        }
        segList.querySelectorAll("SegmentURL").forEach((seg) => {
          urls.push(resolveUrl(manifestUrl, seg.getAttribute("media")));
        });
      } else {
        const baseUrl = rep.querySelector("BaseURL") ||
          (adaptationSet ? adaptationSet.querySelector("BaseURL") : null);
        if (baseUrl) {
          urls.push(resolveUrl(manifestUrl, baseUrl.textContent.trim()));
        }
      }
    }

    return urls;
  }

  function expandDASHTemplate(template, repId, bandwidth, number, time) {
    let result = template;
    result = result.replace(/\$RepresentationID\$/g, repId);
    result = result.replace(/\$Bandwidth\$/g, bandwidth);
    result = result.replace(/\$Time\$/g, String(time));

    result = result.replace(/\$Number(%(\d+)d)?\$/g, (_, _fmt, width) => {
      if (width) return String(number).padStart(parseInt(width, 10), "0");
      return String(number);
    });

    return result;
  }

  function parseDuration(iso) {
    if (!iso) return 0;
    const match = iso.match(
      /PT(?:(\d+(?:\.\d+)?)H)?(?:(\d+(?:\.\d+)?)M)?(?:(\d+(?:\.\d+)?)S)?/
    );
    if (!match) return 0;
    return (
      (parseFloat(match[1] || "0") * 3600) +
      (parseFloat(match[2] || "0") * 60) +
      parseFloat(match[3] || "0")
    );
  }

  async function downloadDirect(url, name) {
    try {
      const resp = await fetchImpl(url);
      const blob = await resp.blob();
      const captureName = name || `direct_${Date.now()}`;
      await uploadToKuraNAS(blob, captureName, "direct");
      return { ok: true, name: captureName };
    } catch (error) {
      return { error: error.message };
    }
  }

  return {
    downloadDASH,
    downloadDirect,
    downloadHLS,
  };
}
