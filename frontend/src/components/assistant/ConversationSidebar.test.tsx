import { fireEvent, render, screen } from '@testing-library/react';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => key,
    }),
}));

import ConversationSidebar from './ConversationSidebar';
import type { Conversation } from '@/service/assistant';

const conversations: Conversation[] = [
    { id: 1, title: 'Primeira', created_at: '', updated_at: '' },
    { id: 2, title: '', created_at: '', updated_at: '' },
];

const baseProps = {
    conversations,
    activeId: 1 as number | null,
    onSelect: jest.fn(),
    onNew: jest.fn(),
    onDelete: jest.fn(),
};

describe('components/assistant/ConversationSidebar', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('shows the empty state when there are no conversations', () => {
        render(<ConversationSidebar {...baseProps} conversations={[]} />);

        expect(screen.getByText('ASSISTANT_NO_CONVERSATIONS')).toBeInTheDocument();
    });

    it('renders conversations and falls back to a label for untitled ones', () => {
        render(<ConversationSidebar {...baseProps} />);

        expect(screen.getByText('Primeira')).toBeInTheDocument();
        expect(screen.getByText('ASSISTANT_UNTITLED')).toBeInTheDocument();
    });

    it('calls onNew when the new-conversation button is clicked', () => {
        const onNew = jest.fn();
        render(<ConversationSidebar {...baseProps} onNew={onNew} />);

        fireEvent.click(screen.getByText('ASSISTANT_NEW_CONVERSATION'));

        expect(onNew).toHaveBeenCalled();
    });

    it('calls onSelect when a conversation is clicked', () => {
        const onSelect = jest.fn();
        render(<ConversationSidebar {...baseProps} onSelect={onSelect} />);

        fireEvent.click(screen.getByText('Primeira'));

        expect(onSelect).toHaveBeenCalledWith(1);
    });

    it('calls onDelete without selecting when the trash icon is clicked', () => {
        const onSelect = jest.fn();
        const onDelete = jest.fn();
        render(<ConversationSidebar {...baseProps} onSelect={onSelect} onDelete={onDelete} />);

        const deleteButtons = screen.getAllByLabelText('ASSISTANT_DELETE_CONVERSATION');
        fireEvent.click(deleteButtons[0]!);

        expect(onDelete).toHaveBeenCalledWith(1);
        expect(onSelect).not.toHaveBeenCalled();
    });
});
