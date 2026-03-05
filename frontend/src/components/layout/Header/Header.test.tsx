import { fireEvent, render, screen } from '@testing-library/react';
import Header from './Header';
import { MemoryRouter } from 'react-router-dom';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

describe('layout/Header', () => {
	it('renders search, clock and mobile drawer', () => {
		render(
			<MemoryRouter>
				<Header showClock currentTime={new Date('2026-03-04T10:00:00.000Z')} />
			</MemoryRouter>,
		);
		expect(screen.getByPlaceholderText('SEARCH_PLACEHOLDER')).toBeInTheDocument();
		expect(screen.getByTitle('NOTIFICATIONS')).toBeInTheDocument();
		expect(screen.getByText(/\d{1,2}:\d{2}:\d{2}/)).toBeInTheDocument();

		fireEvent.click(screen.getAllByRole('button')[0]!);
		expect(screen.getByText('KuraNAS')).toBeInTheDocument();
		expect(screen.getAllByText('ALL_FILES').length).toBeGreaterThan(0);
		expect(screen.getByText('NAV_IMAGES')).toBeInTheDocument();
		fireEvent.click(screen.getByText('NAV_IMAGES'));
	});

	it('renders without clock by default and with missing currentTime', () => {
		render(
			<MemoryRouter>
				<Header />
			</MemoryRouter>,
		);
		expect(screen.queryByText(/\d{1,2}:\d{2}/)).not.toBeInTheDocument();

		const { rerender } = render(
			<MemoryRouter>
				<Header showClock />
			</MemoryRouter>,
		);
		rerender(
			<MemoryRouter>
				<Header showClock />
			</MemoryRouter>,
		);
		expect(screen.queryByText(/\d{1,2}:\d{2}/)).not.toBeInTheDocument();
	});
});
