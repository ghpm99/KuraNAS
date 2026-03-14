import { fireEvent, render, screen } from '@testing-library/react';
import Header from './Header';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

jest.mock('@/components/layout/Sidebar/Sidebar', () => ({
	__esModule: true,
	default: ({ mobile }: { mobile?: boolean }) => <div>{mobile ? 'SidebarMobile' : 'SidebarDesktop'}</div>,
}));

describe('layout/Header', () => {
	beforeEach(() => {
		jest.useFakeTimers();
		jest.setSystemTime(new Date('2026-03-04T10:00:00.000Z'));
	});

	afterEach(() => {
		jest.useRealTimers();
	});

	it('renders search, clock and mobile drawer', () => {
		render(<Header showClock />);
		expect(screen.getByPlaceholderText('SEARCH_PLACEHOLDER')).toBeInTheDocument();
		expect(screen.getByTitle('NOTIFICATIONS')).toBeInTheDocument();
		expect(screen.getByText(/\d{1,2}:\d{2}:\d{2}/)).toBeInTheDocument();

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
