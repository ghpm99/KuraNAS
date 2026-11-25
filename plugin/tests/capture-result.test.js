const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');

const { describeCaptureResponse, describeCaptureFailure, SENDING_RESULT } = require(path.resolve(__dirname, '..', 'popup/capture-result.js'));

test('sending result is progress', () => {
  assert.equal(SENDING_RESULT.kind, 'progress');
});

test('ok response with name is success', () => {
  assert.deepEqual(describeCaptureResponse({ ok: true, name: 'video' }), { text: 'Concluído: video', kind: 'success' });
});

test('ok response without name is success', () => {
  assert.deepEqual(describeCaptureResponse({ ok: true }), { text: 'Concluído', kind: 'success' });
});

test('error response shows the message as error', () => {
  assert.deepEqual(describeCaptureResponse({ error: 'servidor offline' }), { text: 'servidor offline', kind: 'error' });
});

test('master and dash manifest ask for quality', () => {
  assert.equal(describeCaptureResponse({ type: 'master', variants: [] }).text, 'Escolha a qualidade.');
  assert.equal(describeCaptureResponse({ type: 'dash_manifest', representations: [] }).kind, 'progress');
});

test('missing response is error', () => {
  assert.equal(describeCaptureResponse(undefined).kind, 'error');
});

test('unrecognized response is error', () => {
  assert.equal(describeCaptureResponse({}).kind, 'error');
});

test('failure uses the error message', () => {
  assert.deepEqual(describeCaptureFailure(new Error('boom')), { text: 'boom', kind: 'error' });
  assert.equal(describeCaptureFailure(null).kind, 'error');
});
