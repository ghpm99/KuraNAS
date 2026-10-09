import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import BackupSettingsSection from './BackupSettingsSection';
import { apiBase } from '@/service';

// Seam test: only the axios instance is mocked. The real component, the real
// useBackupSettings hook and the real service/backup.ts all run, so this asserts
// the full frontend chain "click Save → which endpoint, which payload" against
// the actual route string and DTO shape the backend handler test decodes.
jest.mock('@/service', () => ({
	apiBase: {
		get: jest.fn(),
		put: jest.fn(),
	},
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const serverSettings = {
	enabled: true,
	destination_path: '/mnt/backup',
	retention_days: 30,
	interval_hours: 24,
};

const renderSection = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<BackupSettingsSection />
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('components/settings/BackupSettingsSection (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockImplementation((url: string) => {
			switch (url) {
				case '/backup/settings':
					return Promise.resolve({ data: serverSettings });
				case '/backup/status':
					return Promise.resolve({
						data: { enabled: true, has_run: true, status: 'completed', last_error: '' },
					});
				case '/backup/pending':
					return Promise.resolve({ data: { pending_files: 3 } });
				default:
					return Promise.reject(new Error(`unexpected GET ${url}`));
			}
		});
		mockedApi.put.mockResolvedValue({ data: serverSettings });
	});

	it('clicking Save issues PUT /backup/settings with the edited payload', async () => {
		renderSection();

		const destination = await screen.findByRole('textbox');
		await waitFor(() => expect((destination as HTMLInputElement).value).toBe('/mnt/backup'));

		fireEvent.change(destination, { target: { value: '/mnt/cold' } });
		fireEvent.click(screen.getByText('SETTINGS_BACKUP_SAVE'));

		await waitFor(() => expect(mockedApi.put).toHaveBeenCalledTimes(1));
		expect(mockedApi.put).toHaveBeenCalledWith('/backup/settings', {
			enabled: true,
			destination_path: '/mnt/cold',
			retention_days: 30,
			interval_hours: 24,
		});
	});
});
