import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import ChatScreen from './ChatScreen';
import { apiBase } from '@/service';

// Seam test: apiBase is mocked for the conversations list + delete; the chat
// send goes through the real streamChatMessage (fetch). The real ChatScreen +
// useAssistantChat + service/assistant.ts run, so each command asserts the exact
// endpoint/payload the backend assistant handlers decode.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn(), delete: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock; delete: jest.Mock };

describe('components/assistant/ChatScreen (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockResolvedValue({ data: [{ id: 3, title: 'Conversa' }] });
		mockedApi.delete.mockResolvedValue({ data: undefined });
		global.fetch = jest.fn().mockResolvedValue({
			ok: true,
			body: { getReader: () => ({ read: () => Promise.resolve({ done: true }) }) },
		}) as unknown as typeof fetch;
	});

	it('sending a message POSTs the message body to /assistant/chat/stream', async () => {
		render(<ChatScreen />);

		fireEvent.change(screen.getByLabelText('ASSISTANT_PLACEHOLDER'), {
			target: { value: 'Olá assistente' },
		});
		fireEvent.click(screen.getByLabelText('ASSISTANT_SEND'));

		await waitFor(() => expect(global.fetch).toHaveBeenCalledTimes(1));
		const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
		expect(String(url)).toContain('/assistant/chat/stream');
		expect(init.method).toBe('POST');
		expect(JSON.parse(init.body)).toEqual({ message: 'Olá assistente' });
	});

	it('deleting a conversation issues DELETE /assistant/conversations/:id', async () => {
		render(<ChatScreen />);
		await screen.findByText('Conversa');

		fireEvent.click(screen.getByLabelText('ASSISTANT_DELETE_CONVERSATION'));

		await waitFor(() =>
			expect(mockedApi.delete).toHaveBeenCalledWith('/assistant/conversations/3')
		);
	});
});
