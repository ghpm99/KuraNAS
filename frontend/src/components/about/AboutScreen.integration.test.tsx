import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter } from 'react-router-dom';
import AboutScreen from './AboutScreen';
import {
	AboutContextProvider,
	type AboutContextType,
} from '@/components/providers/aboutProvider/AboutContext';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real AboutScreen + useAboutScreen +
// service/update.ts run, so the apply button asserts POST /update/apply.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; post: jest.Mock };

const aboutValue: AboutContextType = {
	version: '2.4.0',
	commit_hash: 'abc123',
	platform: 'linux',
	path: '/srv/media',
	lang: 'pt-BR',
	enable_workers: true,
	uptime: '1h',
	statup_time: '2026-03-15T12:00:00.000Z',
	gin_mode: 'release',
	gin_version: '1.10.0',
	go_version: 'go1.24.0',
	node_version: 'v24.1.0',
};

const renderAbout = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<MemoryRouter>
					<AboutContextProvider value={aboutValue}>
						<AboutScreen />
					</AboutContextProvider>
				</MemoryRouter>
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('components/about/AboutScreen (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockResolvedValue({
			data: {
				current_version: '2.4.0',
				latest_version: '2.5.0',
				update_available: true,
				release_url: 'https://x.test',
				release_date: '2026-06-01',
				release_notes: 'notas',
				asset_name: 'kuranas.exe',
				asset_size: 1024,
			},
		});
		mockedApi.post.mockResolvedValue({ data: undefined });
	});

	it('clicking Apply issues POST /update/apply', async () => {
		renderAbout();
		const apply = await screen.findByText('ABOUT_UPDATE_APPLY');

		fireEvent.click(apply);

		await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/update/apply'));
	});

	it('shows the server message verbatim when the package is refused', async () => {
		mockedApi.post.mockRejectedValue({
			response: { status: 409, data: { error: 'O pacote não confere com o checksum publicado.' } },
		});
		renderAbout();

		fireEvent.click(await screen.findByText('ABOUT_UPDATE_APPLY'));

		expect(
			await screen.findByText('O pacote não confere com o checksum publicado.')
		).toBeInTheDocument();
		expect(screen.queryByText('ERROR_UPDATE_APPLY')).not.toBeInTheDocument();
	});

	it('falls back to the generic message when the server sends none', async () => {
		mockedApi.post.mockRejectedValue(new Error('network'));
		renderAbout();

		fireEvent.click(await screen.findByText('ABOUT_UPDATE_APPLY'));

		expect(await screen.findByText('ERROR_UPDATE_APPLY')).toBeInTheDocument();
	});
});
