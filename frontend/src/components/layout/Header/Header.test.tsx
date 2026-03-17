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

jest.mock('@/components/layout/Sidebar/Sidebar', () => ({
    __esModule: true,
    default: ({ mobile }: { mobile?: boolean }) => (
        <div>{mobile ? 'SidebarMobile' : 'SidebarDesktop'}</div>
    ),
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
        jest.useFakeTimers();
        jest.setSystemTime(new Date('2026-03-04T10:00:00.000Z'));
        mockOpenSearch.mockReset();
    });

    afterEach(() => {
        jest.useRealTimers();
    });

    it('renders search, clock and mobile drawer', () => {
        render(<Header showClock />);
        expect(screen.getByText('SEARCH_PLACEHOLDER')).toBeInTheDocument();
        expect(screen.getByTitle('NOTIFICATIONS')).toBeInTheDocument();
        expect(screen.getByText(/\d{1,2}:\d{2}:\d{2}/)).toBeInTheDocument();

        fireEvent.click(screen.getByLabelText('GLOBAL_SEARCH_OPEN'));
        expect(mockOpenSearch).toHaveBeenCalled();

        fireEvent.click(screen.getByLabelText('OPEN_NAVIGATION_MENU'));
        expect(screen.getByText('SidebarMobile')).toBeInTheDocument();
    });

    it('renders without clock by default', () => {
        render(<Header />);
        expect(screen.queryByText(/\d{1,2}:\d{2}/)).not.toBeInTheDocument();

        const { rerender } = render(<Header />);
        rerender(<Header />);
        expect(screen.queryByText(/\d{1,2}:\d{2}/)).not.toBeInTheDocument();
    });
});
