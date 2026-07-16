import assert from 'node:assert/strict';
import test from 'node:test';
import { decodeChatReadResponse, decodeChatSendResponse } from './chat-response.ts';

function response() {
  return {
    Status: 'Success',
    Messages: [{ id: '1', channel: 'all', sender: 0, body: 'Hello', timestamp: 100 }],
    ViewChannels: ['all', 'team/Red'],
    SendChannels: ['all'],
    UserIDMap: { ada: 0 },
    ChatConfig: { Enabled: true, PrebakedOnly: false, AllowedMessages: null as string[] | null },
  };
}

test('chat decoder copies exact bounded data and normalizes Go nil slices', () => {
  assert.deepEqual(decodeChatReadResponse(response()), {
    Messages: [{ id: '1', channel: 'all', sender: 0, body: 'Hello', timestamp: 100 }],
    ViewChannels: ['all', 'team/Red'],
    SendChannels: ['all'],
    UserIDMap: { ada: 0 },
    ChatConfig: { Enabled: true, PrebakedOnly: false, AllowedMessages: [] },
  });
  assert.deepEqual(decodeChatSendResponse({ Status: 'Success', MessageID: '2', secret: true }), {
    MessageID: '2',
  });
});

test('chat decoder rejects duplicate IDs and send-only channels', () => {
  const duplicate = response();
  duplicate.Messages.push({ ...duplicate.Messages[0] });
  assert.throws(() => decodeChatReadResponse(duplicate), /duplicate id/);

  const invisible = response();
  invisible.SendChannels = ['dm/a/b'];
  assert.throws(() => decodeChatReadResponse(invisible), /non-viewable channel/);
});

test('chat decoder rejects contradictory pre-baked configuration', () => {
  const invalid = response();
  invalid.ChatConfig.AllowedMessages = ['yes'];
  assert.throws(() => decodeChatReadResponse(invalid), /requires PrebakedOnly/);
});

test('chat user maps copy special user IDs without prototype mutation', () => {
  const special = response();
  special.UserIDMap = JSON.parse('{"__proto__":0}');
  const decoded = decodeChatReadResponse(special);
  assert.equal(Object.getPrototypeOf(decoded.UserIDMap), Object.prototype);
  assert.equal(Object.hasOwn(decoded.UserIDMap, '__proto__'), true);
  assert.equal(decoded.UserIDMap['__proto__'], 0);
});
