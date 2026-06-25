import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import AutoShutdownSettingsSection from './AutoShutdownSettingsSection';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real section + useAutoShutdownSettings
// + service/autoShutdown.ts run. Save asserts PUT /auto-shutdown/settings with
// the exact payload the backend autoshutdown handler test decodes; the suggest
// button asserts the GET it triggers.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), put: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const serverSettings = { enabled: true, time: '03:00', grace_period_seconds: 60 };

const renderSection = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<AutoShutdownSettingsSection />
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('components/settings/AutoShutdownSettingsSection (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockImplementation((url: string) => {
			if (url === '/auto-shutdown/settings') return Promise.resolve({ data: serverSettings });
			if (url === '/auto-shutdown/suggested-time')
				return Promise.resolve({ data: { available: true, time: '02:30', sample_size: 7 } });
			return Promise.reject(new Error(`unexpected GET ${url}`));
		});
		mockedApi.put.mockResolvedValue({ data: serverSettings });
	});

	it('clicking Save issues PUT /auto-shutdown/settings with the edited payload', async () => {
		renderSection();
		const toggle = await screen.findByRole('switch');

		fireEvent.click(toggle);
		fireEvent.click(screen.getByText('SETTINGS_AUTO_SHUTDOWN_SAVE'));

		await waitFor(() => expect(mockedApi.put).toHaveBeenCalledTimes(1));
		expect(mockedApi.put).toHaveBeenCalledWith('/auto-shutdown/settings', {
			enabled: false,
			time: '03:00',
			grace_period_seconds: 60,
		});
	});

	it('the suggest button fetches GET /auto-shutdown/suggested-time', async () => {
		renderSection();
		await screen.findByRole('switch');

		fireEvent.click(screen.getByText('SETTINGS_AUTO_SHUTDOWN_SUGGEST_BUTTON'));

		await waitFor(() =>
			expect(mockedApi.get).toHaveBeenCalledWith('/auto-shutdown/suggested-time')
		);
	});
});
