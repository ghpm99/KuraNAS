jest.mock('.', () => ({
    apiBase: {
        post: jest.fn(),
    },
}));

import { apiBase } from '.';
import { sendChatMessage, type ChatMessage } from './assistant';

const mockedApiPost = apiBase.post as jest.Mock;

describe('service/assistant', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('posts the conversation history to the chat endpoint', async () => {
        const messages: ChatMessage[] = [
            { role: 'user', content: 'oi' },
            { role: 'assistant', content: 'olá' },
            { role: 'user', content: 'tudo bem?' },
        ];
        const response = {
            message: { role: 'assistant', content: 'Tudo ótimo!' },
            model: 'llama3.1',
            provider: 'ollama',
        };
        mockedApiPost.mockResolvedValue({ data: response });

        const result = await sendChatMessage(messages);

        expect(mockedApiPost).toHaveBeenCalledWith('/assistant/chat', { messages });
        expect(result).toEqual(response);
    });
});
