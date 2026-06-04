import { apiBase } from '.';
import { getApiV1BaseUrl } from './apiUrl';

export type ChatRole = 'user' | 'assistant';

export interface ChatMessage {
    role: ChatRole;
    content: string;
}

export interface ChatResponse {
    conversation_id: number;
    message: ChatMessage;
    model: string;
    provider: string;
}

export interface Conversation {
    id: number;
    title: string;
    created_at: string;
    updated_at: string;
}

export interface StoredMessage {
    id: number;
    role: ChatRole;
    content: string;
    created_at: string;
}

/**
 * Sends a new user message (optionally in an existing conversation) and returns
 * the assistant reply in one shot. Kept as a non-streaming fallback; the UI
 * prefers streamChatMessage.
 */
export const sendChatMessage = async (
    conversationId: number | null,
    message: string
): Promise<ChatResponse> => {
    const response = await apiBase.post<ChatResponse>('/assistant/chat', {
        conversation_id: conversationId ?? undefined,
        message,
    });
    return response.data;
};

export const listConversations = async (): Promise<Conversation[]> => {
    const response = await apiBase.get<Conversation[]>('/assistant/conversations');
    return response.data;
};

export const getConversationMessages = async (conversationId: number): Promise<StoredMessage[]> => {
    const response = await apiBase.get<StoredMessage[]>(
        `/assistant/conversations/${conversationId}/messages`
    );
    return response.data;
};

export const deleteConversation = async (conversationId: number): Promise<void> => {
    await apiBase.delete(`/assistant/conversations/${conversationId}`);
};

export interface ChatStreamCallbacks {
    onDelta: (delta: string) => void;
    onDone: (response: ChatResponse) => void;
    onError: () => void;
}

interface SSEEvent {
    event: string;
    data: string;
}

/**
 * Splits a raw SSE buffer into complete events plus the trailing partial chunk
 * that has not been fully received yet. Pure and incremental so it can be fed
 * one network chunk at a time.
 */
export const parseSSEBuffer = (buffer: string): { events: SSEEvent[]; rest: string } => {
    const parts = buffer.split('\n\n');
    const rest = parts.pop() ?? '';
    const events: SSEEvent[] = [];
    for (const part of parts) {
        const parsed = parseSSEEvent(part);
        if (parsed) {
            events.push(parsed);
        }
    }
    return { events, rest };
};

const parseSSEEvent = (raw: string): SSEEvent | null => {
    let event = 'message';
    const dataLines: string[] = [];
    for (const line of raw.split('\n')) {
        if (line.startsWith('event:')) {
            event = line.slice('event:'.length).trim();
        } else if (line.startsWith('data:')) {
            dataLines.push(line.slice('data:'.length).trim());
        }
    }
    if (dataLines.length === 0) {
        return null;
    }
    return { event, data: dataLines.join('\n') };
};

/**
 * Streams the assistant reply over Server-Sent Events, invoking onDelta for each
 * content chunk and onDone with the final message. Any failure (network, bad
 * status, or a stream that ends without a `done` event) calls onError.
 */
export const streamChatMessage = async (
    conversationId: number | null,
    message: string,
    callbacks: ChatStreamCallbacks
): Promise<void> => {
    let response: Response;
    try {
        response = await fetch(`${getApiV1BaseUrl()}/assistant/chat/stream`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ conversation_id: conversationId ?? undefined, message }),
        });
    } catch {
        callbacks.onError();
        return;
    }

    if (!response.ok || !response.body) {
        callbacks.onError();
        return;
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    let finished = false;

    try {
        for (;;) {
            const { done, value } = await reader.read();
            if (done) {
                break;
            }
            buffer += decoder.decode(value, { stream: true });
            const { events, rest } = parseSSEBuffer(buffer);
            buffer = rest;
            for (const evt of events) {
                if (evt.event === 'delta') {
                    callbacks.onDelta((JSON.parse(evt.data) as { content: string }).content);
                } else if (evt.event === 'done') {
                    callbacks.onDone(JSON.parse(evt.data) as ChatResponse);
                    finished = true;
                } else if (evt.event === 'error') {
                    callbacks.onError();
                    finished = true;
                }
            }
        }
    } catch {
        if (!finished) {
            callbacks.onError();
        }
        return;
    }

    if (!finished) {
        callbacks.onError();
    }
};
