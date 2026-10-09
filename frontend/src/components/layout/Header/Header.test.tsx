import { fireEvent, render, screen } from '@testing-library/react';
import Header from './Header';

const mockOpenSearch = jest.fn();

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, options?: Record<string, string>) =>
            options?.shortcut ? `${key}:${options.shortcut}` : key,
    }),
}));

jest.mock('@/components/search/useGlobalSearch', () => ({
    __esModule: true,
    default: () => ({ openSearch: mockOpenSearch, shortcut: 'Ctrl+K' }),
}));

jest.mock('@/components/providers/notificationProvider/notificationContext', () => ({
    useNotifications: () => ({
        notifications: [],
        unreadCount: 0,
        markAllAsRead: jest.fn(),
        markAsRead: jest.fn(),
        refresh: jest.fn(),
    }),
}));

describe('layout/Header', () => {
    beforeEach(() => {
        mockOpenSearch.mockReset();
    });

    it('renders the search field and the notifications bell', () => {
        render(<Header />);
        expect(screen.getByText('SEARCH_PLACEHOLDER')).toBeInTheDocument();
        expect(screen.getByTitle('NOTIFICATIONS')).toBeInTheDocument();
    });

    it('opens the global search when the search field is clicked', () => {
        render(<Header />);
        fireEvent.click(screen.getByLabelText('GLOBAL_SEARCH_OPEN'));
        expect(mockOpenSearch).toHaveBeenCalled();
    });
});
