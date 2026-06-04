jest.mock('.', () => ({
    apiBase: {
        post: jest.fn(),
        get: jest.fn(),
        delete: jest.fn(),
    },
}));

import { apiBase } from '.';
import {
    deleteConversation,
    getConversationMessages,
    listConversations,
    parseSSEBuffer,
    sendChatMessage,
    streamChatMessage,
    type ChatStreamCallbacks,
} from './assistant';

const mockedPost = apiBase.post as jest.Mock;
const mockedGet = apiBase.get as jest.Mock;
const mockedDelete = apiBase.delete as jest.Mock;

const makeCallbacks = (): ChatStreamCallbacks & {
    deltas: string[];
    done: unknown[];
    errors: number;
} => {
    const deltas: string[] = [];
    const done: unknown[] = [];
    let errors = 0;
    return {
        deltas,
        done,
        get errors() {
            return errors;
        },
        onDelta: (delta) => deltas.push(delta),
        onDone: (response) => done.push(response),
        onError: () => {
            errors += 1;
        },
    };
};

const streamResponse = (chunks: string[], init: { ok?: boolean; hasBody?: boolean } = {}) => {
    const { ok = true, hasBody = true } = init;
    let index = 0;
    const reader = {
        read: async () => {
            if (index < chunks.length) {
                return { done: false, value: new TextEncoder().encode(chunks[index++]) };
            }
            return { done: true, value: undefined };
        },
    };
    return { ok, body: hasBody ? { getReader: () => reader } : null };
};

describe('service/assistant REST', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('posts a new message with the conversation id', async () => {
        const response = {
            conversation_id: 4,
            message: { role: 'assistant', content: 'Tudo ótimo!' },
            model: 'm',
            provider: 'p',
        };
        mockedPost.mockResolvedValue({ data: response });

        const result = await sendChatMessage(4, 'tudo bem?');

        expect(mockedPost).toHaveBeenCalledWith('/assistant/chat', {
            conversation_id: 4,
            message: 'tudo bem?',
        });
        expect(result).toEqual(response);
    });

    it('omits the conversation id for a new conversation', async () => {
        mockedPost.mockResolvedValue({ data: {} });

        await sendChatMessage(null, 'oi');

        expect(mockedPost).toHaveBeenCalledWith('/assistant/chat', {
            conversation_id: undefined,
            message: 'oi',
        });
    });

    it('lists conversations', async () => {
        const conversations = [{ id: 1, title: 'A', created_at: '', updated_at: '' }];
        mockedGet.mockResolvedValue({ data: conversations });

        const result = await listConversations();

        expect(mockedGet).toHaveBeenCalledWith('/assistant/conversations');
        expect(result).toEqual(conversations);
    });

    it('gets the messages of a conversation', async () => {
        mockedGet.mockResolvedValue({ data: [] });

        await getConversationMessages(7);

        expect(mockedGet).toHaveBeenCalledWith('/assistant/conversations/7/messages');
    });

    it('deletes a conversation', async () => {
        mockedDelete.mockResolvedValue({ data: { id: 7 } });

        await deleteConversation(7);

        expect(mockedDelete).toHaveBeenCalledWith('/assistant/conversations/7');
    });
});

describe('service/assistant parseSSEBuffer', () => {
    it('returns complete events and keeps the trailing partial chunk', () => {
        const buffer = 'event: delta\ndata: {"content":"a"}\n\nevent: delta\ndata: {"content":"b"';
        const { events, rest } = parseSSEBuffer(buffer);

        expect(events).toEqual([{ event: 'delta', data: '{"content":"a"}' }]);
        expect(rest).toBe('event: delta\ndata: {"content":"b"');
    });

    it('ignores blocks without a data line', () => {
        const { events } = parseSSEBuffer('event: ping\n\n');
        expect(events).toEqual([]);
    });
});

describe('service/assistant streamChatMessage', () => {
    const originalFetch = global.fetch;

    afterEach(() => {
        global.fetch = originalFetch;
    });

    it('sends the conversation id in the request body', async () => {
        const fetchMock = jest.fn().mockResolvedValue(
            streamResponse([
                'event: done\ndata: {"conversation_id":2,"message":{"role":"assistant","content":"hi"},"model":"m","provider":"p"}\n\n',
            ])
        );
        global.fetch = fetchMock;

        await streamChatMessage(2, 'oi', makeCallbacks());

        expect(fetchMock).toHaveBeenCalledWith(
            expect.stringContaining('/assistant/chat/stream'),
            expect.objectContaining({
                body: JSON.stringify({ conversation_id: 2, message: 'oi' }),
            })
        );
    });

    it('emits deltas and the final done payload', async () => {
        const sse =
            'event: delta\ndata: {"content":"Olá"}\n\n' +
            'event: delta\ndata: {"content":", mundo"}\n\n' +
            'event: done\ndata: {"conversation_id":1,"message":{"role":"assistant","content":"Olá, mundo"},"model":"m","provider":"p"}\n\n';
        global.fetch = jest.fn().mockResolvedValue(streamResponse([sse]));

        const callbacks = makeCallbacks();
        await streamChatMessage(null, 'oi', callbacks);

        expect(callbacks.deltas).toEqual(['Olá', ', mundo']);
        expect(callbacks.done).toHaveLength(1);
        expect(callbacks.errors).toBe(0);
    });

    it('reassembles events split across network chunks', async () => {
        global.fetch = jest
            .fn()
            .mockResolvedValue(
                streamResponse([
                    'event: delta\ndata: {"con',
                    'tent":"hi"}\n\nevent: done\ndata: {"conversation_id":1,"message":{"role":"assistant","content":"hi"},"model":"m","provider":"p"}\n\n',
                ])
            );

        const callbacks = makeCallbacks();
        await streamChatMessage(null, 'oi', callbacks);

        expect(callbacks.deltas).toEqual(['hi']);
        expect(callbacks.done).toHaveLength(1);
    });

    it('calls onError when fetch rejects', async () => {
        global.fetch = jest.fn().mockRejectedValue(new Error('network'));
        const callbacks = makeCallbacks();
        await streamChatMessage(null, 'oi', callbacks);
        expect(callbacks.errors).toBe(1);
    });

    it('calls onError on a non-ok response', async () => {
        global.fetch = jest.fn().mockResolvedValue(streamResponse([], { ok: false }));
        const callbacks = makeCallbacks();
        await streamChatMessage(null, 'oi', callbacks);
        expect(callbacks.errors).toBe(1);
    });

    it('calls onError when the body is missing', async () => {
        global.fetch = jest.fn().mockResolvedValue(streamResponse([], { hasBody: false }));
        const callbacks = makeCallbacks();
        await streamChatMessage(null, 'oi', callbacks);
        expect(callbacks.errors).toBe(1);
    });

    it('calls onError when an error event arrives', async () => {
        global.fetch = jest
            .fn()
            .mockResolvedValue(streamResponse(['event: error\ndata: {"error":"boom"}\n\n']));
        const callbacks = makeCallbacks();
        await streamChatMessage(null, 'oi', callbacks);
        expect(callbacks.errors).toBe(1);
    });

    it('calls onError when the stream ends without a done event', async () => {
        global.fetch = jest
            .fn()
            .mockResolvedValue(streamResponse(['event: delta\ndata: {"content":"x"}\n\n']));
        const callbacks = makeCallbacks();
        await streamChatMessage(null, 'oi', callbacks);
        expect(callbacks.deltas).toEqual(['x']);
        expect(callbacks.errors).toBe(1);
    });
});
