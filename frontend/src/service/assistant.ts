import { apiBase } from '.';

export type ChatRole = 'user' | 'assistant';

export interface ChatMessage {
    role: ChatRole;
    content: string;
}

export interface ChatResponse {
    message: ChatMessage;
    model: string;
    provider: string;
}

/**
 * Sends the full conversation history to the assistant and returns its reply.
 * This first iteration is conversation-only: the client holds the history and
 * the backend does not persist it.
 */
export const sendChatMessage = async (messages: ChatMessage[]): Promise<ChatResponse> => {
    const response = await apiBase.post<ChatResponse>('/assistant/chat', { messages });
    return response.data;
};
