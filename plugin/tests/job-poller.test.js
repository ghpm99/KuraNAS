const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');

const { createJobPoller, isTerminalJobStatus } = require(path.resolve(__dirname, '..', 'popup/job-poller.js'));

function createManualScheduler() {
  const scheduled = [];
  const cancelled = [];
  return {
    scheduled,
    cancelled,
    setTimeoutImpl: (callback, delay) => {
      scheduled.push({ callback, delay });
      return scheduled.length;
    },
    clearTimeoutImpl: (timerId) => cancelled.push(timerId),
    runNext: async () => scheduled.shift().callback(),
  };
}

function jobResponse(status, progress) {
  return { ok: true, status, progress, lastError: '' };
}

test('poller keeps polling every 2s until the job is completed', async () => {
  const scheduler = createManualScheduler();
  const responses = [jobResponse('queued', 0), jobResponse('running', 40), jobResponse('completed', 100)];
  const updates = [];
  const poller = createJobPoller({
    requestJob: async () => responses.shift(),
    onUpdate: (update) => updates.push(`${update.status}:${update.progress}`),
    setTimeoutImpl: scheduler.setTimeoutImpl,
    clearTimeoutImpl: scheduler.clearTimeoutImpl,
  });

  await poller.start(5);
  assert.equal(scheduler.scheduled[0].delay, 2000);
  await scheduler.runNext();
  await scheduler.runNext();

  assert.deepEqual(updates, ['queued:0', 'running:40', 'completed:100']);
  assert.equal(scheduler.scheduled.length, 0);
});

test('poller stops on failed and canceled jobs', async () => {
  for (const terminalStatus of ['failed', 'canceled']) {
    const scheduler = createManualScheduler();
    const poller = createJobPoller({
      requestJob: async () => jobResponse(terminalStatus, 10),
      onUpdate: () => {},
      setTimeoutImpl: scheduler.setTimeoutImpl,
      clearTimeoutImpl: scheduler.clearTimeoutImpl,
    });

    await poller.start(5);

    assert.equal(scheduler.scheduled.length, 0);
  }
});

test('stop cancels the pending poll and ignores a response in flight', async () => {
  const scheduler = createManualScheduler();
  const updates = [];
  let resolveInFlight;
  const poller = createJobPoller({
    requestJob: () => new Promise((resolve) => { resolveInFlight = resolve; }),
    onUpdate: (update) => updates.push(update),
    setTimeoutImpl: scheduler.setTimeoutImpl,
    clearTimeoutImpl: scheduler.clearTimeoutImpl,
  });

  const polling = poller.start(5);
  poller.stop();
  resolveInFlight(jobResponse('running', 50));
  await polling;

  assert.deepEqual(updates, []);
  assert.equal(scheduler.scheduled.length, 0);
});

test('stop clears the scheduled timer', async () => {
  const scheduler = createManualScheduler();
  const poller = createJobPoller({
    requestJob: async () => jobResponse('running', 1),
    onUpdate: () => {},
    setTimeoutImpl: scheduler.setTimeoutImpl,
    clearTimeoutImpl: scheduler.clearTimeoutImpl,
  });

  await poller.start(5);
  poller.stop();

  assert.deepEqual(scheduler.cancelled, [1]);
});

test('poller keeps polling after a transient error response', async () => {
  const scheduler = createManualScheduler();
  const poller = createJobPoller({
    requestJob: async () => ({ ok: false, error: 'HTTP 500' }),
    onUpdate: () => {},
    setTimeoutImpl: scheduler.setTimeoutImpl,
    clearTimeoutImpl: scheduler.clearTimeoutImpl,
  });

  await poller.start(5);

  assert.equal(scheduler.scheduled.length, 1);
});

test('terminal statuses match the backend job statuses', () => {
  assert.equal(isTerminalJobStatus('completed'), true);
  assert.equal(isTerminalJobStatus('failed'), true);
  assert.equal(isTerminalJobStatus('canceled'), true);
  assert.equal(isTerminalJobStatus('running'), false);
  assert.equal(isTerminalJobStatus('queued'), false);
});
