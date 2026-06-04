import { useCallback, useEffect, useState } from 'react';
import {
    deleteConversation as deleteConversationRequest,
    getConversationMessages,
    listConversations,
    streamChatMessage,
    type ChatMessage,
    type Conversation,
} from '@/service/assistant';

export interface AssistantChatState {
    conversations: Conversation[];
    conversationId: number | null;
    messages: ChatMessage[];
    input: string;
    isLoading: boolean;
    hasError: boolean;
    setInput: (value: string) => void;
    send: () => Promise<void>;
    selectConversation: (id: number) => Promise<void>;
    newConversation: () => void;
    removeConversation: (id: number) => Promise<void>;
}

/**
 * Holds the assistant chat state: the list of stored conversations, the active
 * conversation and its messages. History is persisted on the backend; the
 * client sends only the new message plus the conversation id.
 */
export const useAssistantChat = (): AssistantChatState => {
    const [conversations, setConversations] = useState<Conversation[]>([]);
    const [conversationId, setConversationId] = useState<number | null>(null);
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [input, setInput] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [hasError, setHasError] = useState(false);

    const refreshConversations = useCallback(async () => {
        try {
            setConversations(await listConversations());
        } catch {
            // A failed list should not block chatting; leave the current list.
        }
    }, []);

    useEffect(() => {
        void refreshConversations();
    }, [refreshConversations]);

    const selectConversation = useCallback(async (id: number) => {
        setHasError(false);
        try {
            const stored = await getConversationMessages(id);
            setConversationId(id);
            setMessages(stored.map(({ role, content }) => ({ role, content })));
        } catch {
            setHasError(true);
        }
    }, []);

    const newConversation = useCallback(() => {
        setConversationId(null);
        setMessages([]);
        setHasError(false);
    }, []);

    const removeConversation = useCallback(
        async (id: number) => {
            try {
                await deleteConversationRequest(id);
            } catch {
                setHasError(true);
                return;
            }
            if (id === conversationId) {
                setConversationId(null);
                setMessages([]);
            }
            await refreshConversations();
        },
        [conversationId, refreshConversations]
    );

    const send = useCallback(async () => {
        const content = input.trim();
        if (content === '' || isLoading) {
            return;
        }

        const userMessage: ChatMessage = { role: 'user', content };
        setMessages((current) => [...current, userMessage, { role: 'assistant', content: '' }]);
        setInput('');
        setIsLoading(true);
        setHasError(false);

        let streamed = '';
        const replaceLast = (message: ChatMessage) => {
            setMessages((current) => {
                const next = [...current];
                next[next.length - 1] = message;
                return next;
            });
        };

        await streamChatMessage(conversationId, content, {
            onDelta: (delta) => {
                streamed += delta;
                replaceLast({ role: 'assistant', content: streamed });
            },
            onDone: (response) => {
                replaceLast(response.message);
                const isNew = conversationId === null;
                setConversationId(response.conversation_id);
                setIsLoading(false);
                if (isNew) {
                    void refreshConversations();
                }
            },
            onError: () => {
                setMessages((current) => {
                    const last = current[current.length - 1];
                    if (last && last.role === 'assistant' && last.content === '') {
                        return current.slice(0, -1);
                    }
                    return current;
                });
                setHasError(true);
                setIsLoading(false);
            },
        });
    }, [input, isLoading, conversationId, refreshConversations]);

    return {
        conversations,
        conversationId,
        messages,
        input,
        isLoading,
        hasError,
        setInput,
        send,
        selectConversation,
        newConversation,
        removeConversation,
    };
};

export default useAssistantChat;
