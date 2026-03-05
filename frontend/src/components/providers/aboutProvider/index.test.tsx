import { act, render, screen } from '@testing-library/react';
import { AboutProvider } from './index';
import { useAbout } from './AboutContext';
import { useQuery } from '@tanstack/react-query';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
	apiBase: {
		get: jest.fn(),
	},
}));

jest.mock('@tanstack/react-query', () => ({
	useQuery: jest.fn(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => (key === 'LOADING' ? 'Carregando' : key) }),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedApiGet = apiBase.get as jest.Mock;

function Consumer() {
	const about = useAbout();
	return (
		<div>
			<span data-testid="version">{about.version}</span>
			<span data-testid="uptime">{about.uptime}</span>
		</div>
	);
}

describe('providers/aboutProvider/index', () => {
	beforeEach(() => {
		jest.useFakeTimers();
		jest.setSystemTime(new Date('2025-01-01T00:00:10Z'));
		jest.clearAllMocks();
		mockedApiGet.mockResolvedValue({ status: 200, data: {} });
	});

	afterEach(() => {
		jest.useRealTimers();
	});

	it('uses loading text when startup time is missing', () => {
		mockedUseQuery.mockReturnValue({
			data: {
				version: '1.0.0',
				statup_time: '',
			},
		});

		render(
			<AboutProvider>
				<Consumer />
			</AboutProvider>,
		);

		expect(screen.getByTestId('version')).toHaveTextContent('1.0.0');
		expect(screen.getByTestId('uptime')).toHaveTextContent('Carregando');
	});

	it('calculates uptime when startup time exists and updates over time', () => {
		mockedUseQuery.mockReturnValue({
			data: {
				version: '2.0.0',
				statup_time: '2025-01-01T00:00:00Z',
			},
		});

		render(
			<AboutProvider>
				<Consumer />
			</AboutProvider>,
		);

		expect(screen.getByTestId('version')).toHaveTextContent('2.0.0');
		expect(screen.getByTestId('uptime').textContent).toMatch(/\d+s|\d+m/);

		act(() => {
			jest.advanceTimersByTime(1200);
		});
		expect(screen.getByTestId('uptime').textContent).toMatch(/\d+s|\d+m/);
	});

	it('falls back to initial context values when query data is absent', () => {
		mockedUseQuery.mockReturnValue({ data: undefined });

		render(
			<AboutProvider>
				<Consumer />
			</AboutProvider>,
		);

		expect(screen.getByTestId('version')).toHaveTextContent('');
		expect(screen.getByTestId('uptime')).toHaveTextContent('Carregando');
	});

	it('executes queryFn and returns API payload on success status', async () => {
		mockedApiGet.mockResolvedValueOnce({ status: 200, data: { version: '3.0.0' } });
		mockedUseQuery.mockReturnValue({ data: undefined });

		render(
			<AboutProvider>
				<Consumer />
			</AboutProvider>,
		);

		const options = mockedUseQuery.mock.calls[0][0];
		await expect(options.queryFn()).resolves.toEqual({ version: '3.0.0' });
		expect(mockedApiGet).toHaveBeenCalledWith('configuration/about');
	});

	it('executes queryFn and throws for non-200 status', async () => {
		mockedApiGet.mockResolvedValueOnce({ status: 500, data: {} });
		mockedUseQuery.mockReturnValue({ data: undefined });

		render(
			<AboutProvider>
				<Consumer />
			</AboutProvider>,
		);

		const options = mockedUseQuery.mock.calls[0][0];
		await expect(options.queryFn()).rejects.toThrow('Network response was not ok');
	});
});
