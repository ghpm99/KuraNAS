const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { pathToFileURL } = require('node:url');

const pluginRoot = path.resolve(__dirname, '..');
const moduleUrl = pathToFileURL(
  path.join(pluginRoot, 'src/background/capture-session.js')
).href;

function immediateTimers() {
  // Run grace timers synchronously so the transitions are deterministic.
  return {
    setTimeoutFn: (fn) => {
      fn();
      return 1;
    },
    clearTimeoutFn: () => {},
  };
}

function manualTimers() {
  let pending = null;
  return {
    fire: () => {
      const fn = pending;
      pending = null;
      if (fn) fn();
    },
    setTimeoutFn: (fn) => {
      pending = fn;
      return 1;
    },
    clearTimeoutFn: () => {
      pending = null;
    },
  };
}

function playing(overrides) {
  return {
    service: 'crunchyroll',
    episodeId: 'GEP01',
    title: 'Show - E1',
    isPlaying: true,
    isFullscreen: true,
    isEnded: false,
    currentTime: 30,
    duration: 1400,
    ...overrides,
  };
}

async function makeMachine(timers, overrides = {}) {
  const { createCaptureSessionMachine } = await import(moduleUrl);
  const captureSessions = new Map();
  const events = [];
  const machine = createCaptureSessionMachine({
    captureSessions,
    startCapture: (tabId, info) => {
      events.push(['start', tabId, info.episodeKey]);
      return { recording: true };
    },
    stopCapture: (tabId, info) => {
      events.push(['stop', tabId, info.episodeKey]);
    },
    graceMs: 8000,
    endEpsilonSeconds: 2,
    ...timers,
    ...overrides,
  });
  return { machine, captureSessions, events };
}

test('builds episode key from service and id', async () => {
  const { machine } = await makeMachine(immediateTimers());
  assert.equal(
    machine.buildEpisodeKey({ service: 'crunchyroll', episodeId: 'GEP01' }),
    'crunchyroll:GEP01'
  );
  assert.equal(machine.buildEpisodeKey({ service: 'crunchyroll' }), null);
});

test('starts recording on fullscreen playback and idles otherwise', async () => {
  const { machine, events } = await makeMachine(immediateTimers());

  // Not fullscreen -> no recording (clean frames only).
  machine.handleEpisodeState(1, playing({ isFullscreen: false }));
  assert.equal(machine.getSessionStatus(1).state, 'IDLE');
  assert.equal(events.length, 0);

  machine.handleEpisodeState(1, playing());
  assert.equal(machine.getSessionStatus(1).state, 'RECORDING');
  assert.deepEqual(events, [['start', 1, 'crunchyroll:GEP01']]);
});

test('finalizes when the episode reaches its end (owner asleep)', async () => {
  const { machine, events } = await makeMachine(immediateTimers());

  machine.handleEpisodeState(1, playing());
  machine.handleEpisodeState(1, playing({ currentTime: 1399, duration: 1400 }));

  assert.equal(machine.getSessionStatus(1).state, 'IDLE');
  assert.deepEqual(events, [
    ['start', 1, 'crunchyroll:GEP01'],
    ['stop', 1, 'crunchyroll:GEP01'],
  ]);
});

test('autoplay of the next episode starts a separate file', async () => {
  const { machine, events } = await makeMachine(immediateTimers());

  machine.handleEpisodeState(1, playing());
  machine.handleEpisodeState(1, playing({ episodeId: 'GEP02', title: 'Show - E2', currentTime: 5 }));

  assert.equal(machine.getSessionStatus(1).episodeKey, 'crunchyroll:GEP02');
  assert.deepEqual(events, [
    ['start', 1, 'crunchyroll:GEP01'],
    ['stop', 1, 'crunchyroll:GEP01'],
    ['start', 1, 'crunchyroll:GEP02'],
  ]);
});

test('short pause does not finalize; resume keeps the same capture', async () => {
  const timers = manualTimers();
  const { machine, events } = await makeMachine({
    setTimeoutFn: timers.setTimeoutFn,
    clearTimeoutFn: timers.clearTimeoutFn,
  });

  machine.handleEpisodeState(1, playing());
  // Pause (grace window opens, but does not fire yet).
  machine.handleEpisodeState(1, playing({ isPlaying: false, currentTime: 40 }));
  assert.equal(machine.getSessionStatus(1).state, 'RECORDING');
  // Resume within the window -> still the same single capture.
  machine.handleEpisodeState(1, playing({ currentTime: 41 }));
  assert.equal(machine.getSessionStatus(1).state, 'RECORDING');

  assert.deepEqual(events, [['start', 1, 'crunchyroll:GEP01']]);
});

test('pause beyond the grace window finalizes the capture', async () => {
  const timers = manualTimers();
  const { machine, events } = await makeMachine({
    setTimeoutFn: timers.setTimeoutFn,
    clearTimeoutFn: timers.clearTimeoutFn,
  });

  machine.handleEpisodeState(1, playing());
  machine.handleEpisodeState(1, playing({ isPlaying: false, currentTime: 40 }));
  timers.fire(); // grace window elapses, still paused

  assert.equal(machine.getSessionStatus(1).state, 'IDLE');
  assert.deepEqual(events, [
    ['start', 1, 'crunchyroll:GEP01'],
    ['stop', 1, 'crunchyroll:GEP01'],
  ]);
});

test('an already-archived episode is not re-recorded', async () => {
  const { machine, events } = await makeMachine(immediateTimers(), {
    startCapture: (tabId, info) => {
      events.push(['start', tabId, info.episodeKey]);
      return { recording: false }; // backend: already archived
    },
  });

  machine.handleEpisodeState(1, playing());
  // The "already archived" verdict comes from the (async) backend init, so let
  // the revert microtask settle before asserting.
  await Promise.resolve();
  assert.equal(machine.getSessionStatus(1).state, 'IDLE');
  // Replays of the same episode must not arm a second time.
  machine.handleEpisodeState(1, playing({ currentTime: 60 }));

  assert.deepEqual(events, [['start', 1, 'crunchyroll:GEP01']]);
});

test('cleanupTab stops an in-flight recording', async () => {
  const { machine, events } = await makeMachine(immediateTimers());

  machine.handleEpisodeState(1, playing());
  machine.cleanupTab(1);

  assert.equal(machine.getSessionStatus(1).state, 'IDLE');
  assert.deepEqual(events, [
    ['start', 1, 'crunchyroll:GEP01'],
    ['stop', 1, 'crunchyroll:GEP01'],
  ]);
});
