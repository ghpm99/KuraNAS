import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import YtDlpSettingsSection from './YtDlpSettingsSection';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real section + useYtDlpSettings +
// service/ingest.ts run, so clicking Update asserts POST /ingest/ytdlp/update.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const renderSection = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<YtDlpSettingsSection />
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('components/settings/YtDlpSettingsSection (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockResolvedValue({
			data: {
				installed: true,
				current_version: '2026.01.01',
				latest_version: '2026.06.01',
				update_available: true,
				release_url: 'https://x.test',
			},
		});
		mockedApi.post.mockResolvedValue({ data: undefined });
	});

	it('clicking Update issues POST /ingest/ytdlp/update', async () => {
		renderSection();
		const button = await screen.findByText('SETTINGS_YTDLP_UPDATE_BUTTON');

		fireEvent.click(button);

		await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/ingest/ytdlp/update'));
	});
});
