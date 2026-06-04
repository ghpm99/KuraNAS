import { act, renderHook, waitFor } from '@testing-library/react';

jest.mock('@/service/assistant', () => ({
    streamChatMessage: jest.fn(),
    listConversations: jest.fn(),
    getConversationMessages: jest.fn(),
    deleteConversation: jest.fn(),
}));

import {
    deleteConversation,
    getConversationMessages,
    listConversations,
    streamChatMessage,
    type ChatStreamCallbacks,
} from '@/service/assistant';
import { useAssistantChat } from './useAssistantChat';

const mockedStream = streamChatMessage as jest.Mock;
const mockedList = listConversations as jest.Mock;
const mockedGetMessages = getConversationMessages as jest.Mock;
const mockedDelete = deleteConversation as jest.Mock;

const conversation = (id: number) => ({ id, title: `Conversa ${id}`, created_at: '', updated_at: '' });

describe('components/assistant/useAssistantChat', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedList.mockResolvedValue([]);
        mockedGetMessages.mockResolvedValue([]);
        mockedDelete.mockResolvedValue(undefined);
    });

    it('loads conversations on mount', async () => {
        mockedList.mockResolvedValue([conversation(1)]);
        const { result } = renderHook(() => useAssistantChat());

        await waitFor(() => expect(result.current.conversations).toHaveLength(1));
        expect(result.current.messages).toEqual([]);
        expect(result.current.conversationId).toBeNull();
    });

    it('tolerates a failed conversation list', async () => {
        mockedList.mockRejectedValue(new Error('down'));
        const { result } = renderHook(() => useAssistantChat());

        await waitFor(() => expect(mockedList).toHaveBeenCalled());
        expect(result.current.conversations).toEqual([]);
    });

    it('streams deltas, finalizes and adopts the conversation id', async () => {
        mockedStream.mockImplementation(async (_id, _msg, callbacks: ChatStreamCallbacks) => {
            callbacks.onDelta('Olá');
            callbacks.onDone({
                conversation_id: 42,
                message: { role: 'assistant', content: 'Olá!' },
                model: 'm',
                provider: 'p',
            });
        });
        const { result } = renderHook(() => useAssistantChat());
        await waitFor(() => expect(mockedList).toHaveBeenCalled());

        act(() => result.current.setInput('oi'));
        await act(async () => {
            await result.current.send();
        });

        expect(mockedStream).toHaveBeenCalledWith(null, 'oi', expect.any(Object));
        expect(result.current.messages).toEqual([
            { role: 'user', content: 'oi' },
            { role: 'assistant', content: 'Olá!' },
        ]);
        expect(result.current.conversationId).toBe(42);
        // refreshed after creating a new conversation
        expect(mockedList).toHaveBeenCalledTimes(2);
    });

    it('ignores empty input', async () => {
        const { result } = renderHook(() => useAssistantChat());
        await waitFor(() => expect(mockedList).toHaveBeenCalled());

        act(() => result.current.setInput('   '));
        await act(async () => {
            await result.current.send();
        });

        expect(mockedStream).not.toHaveBeenCalled();
    });

    it('drops the placeholder and flags error when the stream fails', async () => {
        mockedStream.mockImplementation(async (_id, _msg, callbacks: ChatStreamCallbacks) => {
            callbacks.onError();
        });
        const { result } = renderHook(() => useAssistantChat());
        await waitFor(() => expect(mockedList).toHaveBeenCalled());

        act(() => result.current.setInput('oi'));
        await act(async () => {
            await result.current.send();
        });

        expect(result.current.hasError).toBe(true);
        expect(result.current.messages).toEqual([{ role: 'user', content: 'oi' }]);
    });

    it('keeps partial content when an error arrives mid-stream', async () => {
        mockedStream.mockImplementation(async (_id, _msg, callbacks: ChatStreamCallbacks) => {
            callbacks.onDelta('parcial');
            callbacks.onError();
        });
        const { result } = renderHook(() => useAssistantChat());
        await waitFor(() => expect(mockedList).toHaveBeenCalled());

        act(() => result.current.setInput('oi'));
        await act(async () => {
            await result.current.send();
        });

        expect(result.current.messages).toEqual([
            { role: 'user', content: 'oi' },
            { role: 'assistant', content: 'parcial' },
        ]);
    });

    it('selects a conversation and loads its messages', async () => {
        mockedGetMessages.mockResolvedValue([
            { id: 1, role: 'user', content: 'oi', created_at: '' },
            { id: 2, role: 'assistant', content: 'olá', created_at: '' },
        ]);
        const { result } = renderHook(() => useAssistantChat());
        await waitFor(() => expect(mockedList).toHaveBeenCalled());

        await act(async () => {
            await result.current.selectConversation(5);
        });

        expect(result.current.conversationId).toBe(5);
        expect(result.current.messages).toEqual([
            { role: 'user', content: 'oi' },
            { role: 'assistant', content: 'olá' },
        ]);
    });

    it('flags error when loading a conversation fails', async () => {
        mockedGetMessages.mockRejectedValue(new Error('boom'));
        const { result } = renderHook(() => useAssistantChat());
        await waitFor(() => expect(mockedList).toHaveBeenCalled());

        await act(async () => {
            await result.current.selectConversation(5);
        });

        expect(result.current.hasError).toBe(true);
    });

    it('starts a new conversation', async () => {
        const { result } = renderHook(() => useAssistantChat());
        await waitFor(() => expect(mockedList).toHaveBeenCalled());

        await act(async () => {
            await result.current.selectConversation(5);
        });
        act(() => result.current.newConversation());

        expect(result.current.conversationId).toBeNull();
        expect(result.current.messages).toEqual([]);
    });

    it('removes the active conversation and clears it', async () => {
        const { result } = renderHook(() => useAssistantChat());
        await waitFor(() => expect(mockedList).toHaveBeenCalled());
        await act(async () => {
            await result.current.selectConversation(5);
        });

        await act(async () => {
            await result.current.removeConversation(5);
        });

        expect(mockedDelete).toHaveBeenCalledWith(5);
        expect(result.current.conversationId).toBeNull();
        expect(result.current.messages).toEqual([]);
    });

    it('flags error when deletion fails', async () => {
        mockedDelete.mockRejectedValue(new Error('boom'));
        const { result } = renderHook(() => useAssistantChat());
        await waitFor(() => expect(mockedList).toHaveBeenCalled());

        await act(async () => {
            await result.current.removeConversation(9);
        });

        expect(result.current.hasError).toBe(true);
    });

    it('does not send while a previous request is in flight', async () => {
        let release: () => void = () => {};
        mockedStream.mockImplementation(
            () =>
                new Promise<void>((resolve) => {
                    release = resolve;
                })
        );
        const { result } = renderHook(() => useAssistantChat());
        await waitFor(() => expect(mockedList).toHaveBeenCalled());

        act(() => result.current.setInput('primeira'));
        let firstSend: Promise<void> = Promise.resolve();
        act(() => {
            firstSend = result.current.send();
        });
        await waitFor(() => expect(result.current.isLoading).toBe(true));

        act(() => result.current.setInput('segunda'));
        await act(async () => {
            await result.current.send();
        });

        expect(mockedStream).toHaveBeenCalledTimes(1);
        await act(async () => {
            release();
            await firstSend;
        });
    });
});
