import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import TieringSettingsSection from './TieringSettingsSection';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real section + useTieringSettings +
// service/tiering.ts run, so clicking Save asserts PUT /tiering/settings with
// the exact payload the backend tiering handler test decodes.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), put: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const serverSettings = {
	enabled: true,
	cold_dir_path: '/mnt/cold',
	min_age_days: 90,
	min_size_bytes: 1048576,
	interval_hours: 24,
};

const renderSection = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<TieringSettingsSection />
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('components/settings/TieringSettingsSection (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockImplementation((url: string) => {
			switch (url) {
				case '/tiering/settings':
					return Promise.resolve({ data: serverSettings });
				case '/tiering/status':
					return Promise.resolve({ data: { enabled: true, has_run: true, status: 'completed', last_error: '' } });
				case '/tiering/usage':
					return Promise.resolve({ data: { hot_files: 1, hot_bytes: 1, cold_files: 1, cold_bytes: 1 } });
				default:
					return Promise.reject(new Error(`unexpected GET ${url}`));
			}
		});
		mockedApi.put.mockResolvedValue({ data: serverSettings });
	});

	it('clicking Save issues PUT /tiering/settings with the edited payload', async () => {
		renderSection();

		const coldDir = await screen.findByRole('textbox');
		await waitFor(() => expect((coldDir as HTMLInputElement).value).toBe('/mnt/cold'));

		fireEvent.change(coldDir, { target: { value: '/mnt/frio' } });
		fireEvent.click(screen.getByText('SETTINGS_TIERING_SAVE'));

		await waitFor(() => expect(mockedApi.put).toHaveBeenCalledTimes(1));
		expect(mockedApi.put).toHaveBeenCalledWith('/tiering/settings', {
			enabled: true,
			cold_dir_path: '/mnt/frio',
			min_age_days: 90,
			min_size_bytes: 1048576,
			interval_hours: 24,
		});
	});
});
