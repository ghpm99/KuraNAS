const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { pathToFileURL } = require('node:url');

const pluginRoot = path.resolve(__dirname, '..');
const hybridStateModuleUrl = pathToFileURL(
  path.join(pluginRoot, 'src/background/hybrid-state.js')
).href;

test('hybrid state machine arms and disarms monitor flow', async () => {
  const { createHybridStateMachine } = await import(hybridStateModuleUrl);
  const hybridStates = new Map();
  const tabMessages = [];
  const statusEvents = [];

  const machine = createHybridStateMachine({
    hybridStates,
    broadcastHybridStatus: (tabId) => statusEvents.push(tabId),
    initHybridUploadSession: async () => {},
    ensureOffscreen: async () => {},
    getMediaStreamId: async () => 'stream-id',
    sendRuntimeMessage: () => {},
    sendTabMessage: (tabId, message) => tabMessages.push([tabId, message.action]),
    stopOffscreenRecording: () => {},
    hybridStabilityMs: 200,
    hybridStopGraceMs: 5000,
    setTimeoutFn: (fn) => {
      fn();
      return 1;
    },
    clearTimeoutFn: () => {},
  });

  machine.armHybrid(10);
  let status = machine.getHybridStatus(10);
  assert.equal(status.armed, true);
  assert.equal(status.state, 'ARMED');
  assert.deepEqual(tabMessages[0], [10, 'hybrid_monitor_start']);

  machine.disarmHybrid(10);
  status = machine.getHybridStatus(10);
  assert.equal(status.armed, false);
  assert.equal(status.state, 'IDLE');
  assert.deepEqual(tabMessages[1], [10, 'hybrid_monitor_stop']);
  assert.ok(statusEvents.length >= 2);
});

test('hybrid state starts recording when stable playback conditions are met', async () => {
  const { createHybridStateMachine } = await import(hybridStateModuleUrl);
  const hybridStates = new Map();
  const runtimeMessages = [];

  const machine = createHybridStateMachine({
    hybridStates,
    broadcastHybridStatus: () => {},
    initHybridUploadSession: async () => {},
    ensureOffscreen: async () => {},
    getMediaStreamId: async () => 'stream-xyz',
    sendRuntimeMessage: (message) => runtimeMessages.push(message),
    sendTabMessage: () => {},
    stopOffscreenRecording: () => {},
    hybridStabilityMs: 200,
    hybridStopGraceMs: 5000,
    setTimeoutFn: (fn) => {
      fn();
      return 1;
    },
    clearTimeoutFn: () => {},
  });

  machine.armHybrid(20);
  machine.handleHybridVideoState(20, {
    hasVideo: true,
    isPlaying: true,
    isFullscreen: true,
    isEnded: false,
    url: 'https://example.com/video',
  });

  await new Promise((resolve) => setTimeout(resolve, 0));

  assert.equal(runtimeMessages.length, 1);
  assert.equal(runtimeMessages[0].action, 'offscreen_start_recording');
  assert.equal(runtimeMessages[0].tabId, 20);
  assert.equal(machine.getHybridStatus(20).state, 'RECORDING');
});
