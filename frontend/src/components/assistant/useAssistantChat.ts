import { useCallback, useState } from 'react';
import { sendChatMessage, type ChatMessage } from '@/service/assistant';

export interface AssistantChatState {
    messages: ChatMessage[];
    input: string;
    isLoading: boolean;
    hasError: boolean;
    setInput: (value: string) => void;
    send: () => Promise<void>;
}

/**
 * Holds the conversation state for the assistant chat. The history lives only in
 * memory (this first iteration has no persistence) and is sent in full on every
 * turn so the backend can keep the dialogue coherent.
 */
export const useAssistantChat = (): AssistantChatState => {
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [input, setInput] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [hasError, setHasError] = useState(false);

    const send = useCallback(async () => {
        const content = input.trim();
        if (content === '' || isLoading) {
            return;
        }

        const nextMessages: ChatMessage[] = [...messages, { role: 'user', content }];
        setMessages(nextMessages);
        setInput('');
        setIsLoading(true);
        setHasError(false);

        try {
            const response = await sendChatMessage(nextMessages);
            setMessages((current) => [...current, response.message]);
        } catch {
            setHasError(true);
        } finally {
            setIsLoading(false);
        }
    }, [input, isLoading, messages]);

    return { messages, input, isLoading, hasError, setInput, send };
};

export default useAssistantChat;
