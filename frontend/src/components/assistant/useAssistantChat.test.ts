import { act, renderHook, waitFor } from '@testing-library/react';

jest.mock('@/service/assistant', () => ({
    sendChatMessage: jest.fn(),
}));

import { sendChatMessage } from '@/service/assistant';
import { useAssistantChat } from './useAssistantChat';

const mockedSend = sendChatMessage as jest.Mock;

describe('components/assistant/useAssistantChat', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('starts empty', () => {
        const { result } = renderHook(() => useAssistantChat());

        expect(result.current.messages).toEqual([]);
        expect(result.current.input).toBe('');
        expect(result.current.isLoading).toBe(false);
        expect(result.current.hasError).toBe(false);
    });

    it('appends the user message and the assistant reply on send', async () => {
        mockedSend.mockResolvedValue({
            message: { role: 'assistant', content: 'Olá!' },
            model: 'llama3.1',
            provider: 'ollama',
        });

        const { result } = renderHook(() => useAssistantChat());

        act(() => result.current.setInput('oi'));
        await act(async () => {
            await result.current.send();
        });

        expect(mockedSend).toHaveBeenCalledWith([{ role: 'user', content: 'oi' }]);
        expect(result.current.messages).toEqual([
            { role: 'user', content: 'oi' },
            { role: 'assistant', content: 'Olá!' },
        ]);
        expect(result.current.input).toBe('');
        expect(result.current.isLoading).toBe(false);
        expect(result.current.hasError).toBe(false);
    });

    it('ignores empty or whitespace-only input', async () => {
        const { result } = renderHook(() => useAssistantChat());

        act(() => result.current.setInput('   '));
        await act(async () => {
            await result.current.send();
        });

        expect(mockedSend).not.toHaveBeenCalled();
        expect(result.current.messages).toEqual([]);
    });

    it('flags an error when the request fails but keeps the user message', async () => {
        mockedSend.mockRejectedValue(new Error('boom'));

        const { result } = renderHook(() => useAssistantChat());

        act(() => result.current.setInput('oi'));
        await act(async () => {
            await result.current.send();
        });

        await waitFor(() => expect(result.current.hasError).toBe(true));
        expect(result.current.messages).toEqual([{ role: 'user', content: 'oi' }]);
        expect(result.current.isLoading).toBe(false);
    });

    it('does not send while a previous request is in flight', async () => {
        let resolveSend: (value: unknown) => void = () => {};
        mockedSend.mockReturnValue(
            new Promise((resolve) => {
                resolveSend = resolve;
            })
        );

        const { result } = renderHook(() => useAssistantChat());

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

        expect(mockedSend).toHaveBeenCalledTimes(1);

        await act(async () => {
            resolveSend({ message: { role: 'assistant', content: 'ok' }, model: 'm', provider: 'p' });
            await firstSend;
        });
    });
});
