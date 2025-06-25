import { render, screen } from '@testing-library/react';
import StatusCard from './StatusCard';

// Mocks
jest.mock('@/components/ui/Card/Card', () => ({ title, children }: any) => (
	<div data-testid='card'>
		<div>{title}</div>
		{children}
	</div>
));

const mockT = (key: string) => key;
jest.mock('@/components/i18n/provider/i18nContext', () => () => ({
	t: mockT,
}));

const mockUseAbout = jest.fn();
jest.mock('@/components/hooks/AboutProvider/AboutContext', () => ({
	useAbout: () => mockUseAbout(),
}));

describe('StatusCard', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('renders enabled workers info', () => {
		mockUseAbout.mockReturnValue({
			enable_workers: true,
			path: '/mnt/data',
			uptime: '2 days',
		});

		render(<StatusCard />);

		expect(screen.getByText('STATUS_SYSTEM_TITLE')).toBeInTheDocument();
		expect(screen.getByText('WORKERS')).toBeInTheDocument();
		expect(screen.getByText('ENABLED_WORKERS')).toBeInTheDocument();
		expect(screen.getByText('ENABLED_WORKERS_DESCRIPTION')).toBeInTheDocument();
		expect(screen.getByText('/mnt/data')).toBeInTheDocument();
		expect(screen.getByText('WATCH_PATH')).toBeInTheDocument();
		expect(screen.getByText('WATCH_PATH_DESCRIPTION')).toBeInTheDocument();
		expect(screen.getByText('UPTIME')).toBeInTheDocument();
		expect(screen.getByText('2 days')).toBeInTheDocument();
		expect(screen.getByText('UPTIME_DESCRIPTION')).toBeInTheDocument();
	});

	it('renders disabled workers info', () => {
		mockUseAbout.mockReturnValue({
			enable_workers: false,
			path: '/mnt/backup',
			uptime: '5 hours',
		});

		render(<StatusCard />);

		expect(screen.getByText('DISABLED_WORKERS')).toBeInTheDocument();
		expect(screen.getByText('DISABLED_WORKERS_DESCRIPTION')).toBeInTheDocument();
		expect(screen.getByText('/mnt/backup')).toBeInTheDocument();
		expect(screen.getByText('5 hours')).toBeInTheDocument();
	});
});
