(function (root) {
  "use strict";

  const TERMINAL_JOB_STATUSES = ["completed", "failed", "canceled"];
  const POLL_INTERVAL_MS = 2000;

  function isTerminalJobStatus(status) {
    return TERMINAL_JOB_STATUSES.includes(status);
  }

  function createJobPoller({
    requestJob,
    onUpdate,
    intervalMs = POLL_INTERVAL_MS,
    setTimeoutImpl = setTimeout,
    clearTimeoutImpl = clearTimeout,
  }) {
    let timerId = null;
    let activeGeneration = 0;

    async function pollOnce(jobId, generation) {
      const response = await requestJob(jobId);
      if (generation !== activeGeneration) {
        return;
      }
      onUpdate(response);
      if (response && response.ok && isTerminalJobStatus(response.status)) {
        stop();
        return;
      }
      timerId = setTimeoutImpl(() => pollOnce(jobId, generation), intervalMs);
    }

    function start(jobId) {
      stop();
      return pollOnce(jobId, activeGeneration);
    }

    function stop() {
      activeGeneration += 1;
      if (timerId !== null) {
        clearTimeoutImpl(timerId);
        timerId = null;
      }
    }

    return { start, stop };
  }

  const api = { createJobPoller, isTerminalJobStatus, POLL_INTERVAL_MS };
  if (typeof module !== "undefined" && module.exports) {
    module.exports = api;
  }
  root.KuraNasJobPoller = api;
})(typeof globalThis !== "undefined" ? globalThis : this);
