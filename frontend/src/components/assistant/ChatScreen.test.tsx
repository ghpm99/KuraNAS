import { fireEvent, render, screen } from '@testing-library/react';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => key,
    }),
}));

const mockUseAssistantChat = jest.fn();
jest.mock('./useAssistantChat', () => ({
    __esModule: true,
    default: () => mockUseAssistantChat(),
}));

import ChatScreen from './ChatScreen';

const baseState = {
    messages: [],
    input: '',
    isLoading: false,
    hasError: false,
    setInput: jest.fn(),
    send: jest.fn(),
};

describe('components/assistant/ChatScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('shows the empty state when there are no messages', () => {
        mockUseAssistantChat.mockReturnValue({ ...baseState });

        render(<ChatScreen />);

        expect(screen.getByText('ASSISTANT_EMPTY')).toBeInTheDocument();
    });

    it('renders user and assistant bubbles', () => {
        mockUseAssistantChat.mockReturnValue({
            ...baseState,
            messages: [
                { role: 'user', content: 'oi' },
                { role: 'assistant', content: 'olá!' },
            ],
        });

        render(<ChatScreen />);

        expect(screen.getByText('oi')).toBeInTheDocument();
        expect(screen.getByText('olá!')).toBeInTheDocument();
        expect(screen.queryByText('ASSISTANT_EMPTY')).not.toBeInTheDocument();
    });

    it('shows the thinking indicator while loading', () => {
        mockUseAssistantChat.mockReturnValue({ ...baseState, isLoading: true });

        render(<ChatScreen />);

        expect(screen.getByText('ASSISTANT_THINKING')).toBeInTheDocument();
    });

    it('shows an error message when hasError is set', () => {
        mockUseAssistantChat.mockReturnValue({ ...baseState, hasError: true });

        render(<ChatScreen />);

        expect(screen.getByRole('alert')).toHaveTextContent('ASSISTANT_ERROR');
    });

    it('updates the input on change', () => {
        const setInput = jest.fn();
        mockUseAssistantChat.mockReturnValue({ ...baseState, setInput });

        render(<ChatScreen />);
        fireEvent.change(screen.getByLabelText('ASSISTANT_PLACEHOLDER'), {
            target: { value: 'oi' },
        });

        expect(setInput).toHaveBeenCalledWith('oi');
    });

    it('sends on send-button click', () => {
        const send = jest.fn();
        mockUseAssistantChat.mockReturnValue({ ...baseState, input: 'oi', send });

        render(<ChatScreen />);
        fireEvent.click(screen.getByLabelText('ASSISTANT_SEND'));

        expect(send).toHaveBeenCalled();
    });

    it('disables the send button when the input is blank', () => {
        mockUseAssistantChat.mockReturnValue({ ...baseState, input: '   ' });

        render(<ChatScreen />);

        expect(screen.getByLabelText('ASSISTANT_SEND')).toBeDisabled();
    });

    it('sends on Enter and not on Shift+Enter', () => {
        const send = jest.fn();
        mockUseAssistantChat.mockReturnValue({ ...baseState, input: 'oi', send });

        render(<ChatScreen />);
        const field = screen.getByLabelText('ASSISTANT_PLACEHOLDER');

        fireEvent.keyDown(field, { key: 'Enter', shiftKey: true });
        expect(send).not.toHaveBeenCalled();

        fireEvent.keyDown(field, { key: 'Enter' });
        expect(send).toHaveBeenCalledTimes(1);
    });
});
