import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import StorageRootsSettingsSection from './StorageRootsSettingsSection';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real section + useStorageRootsSettings
// + service/storageRoots.ts run, so each command button asserts the exact HTTP
// method/endpoint/payload the backend storageroots handler test decodes.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
	put: jest.Mock;
	delete: jest.Mock;
};

// roots[0] is treated as primary (toggle/delete disabled); operate on roots[1].
const roots = [
	{ id: 1, path: '/data', label: 'Principal', enabled: true, created_at: '2026-01-01T00:00:00Z' },
	{ id: 2, path: '/mnt/midia', label: 'Mídia', enabled: true, created_at: '2026-01-01T00:00:00Z' },
];

const renderSection = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<StorageRootsSettingsSection />
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('components/settings/StorageRootsSettingsSection (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockResolvedValue({ data: roots });
		mockedApi.post.mockResolvedValue({ data: roots[1] });
		mockedApi.put.mockResolvedValue({ data: roots[1] });
		mockedApi.delete.mockResolvedValue({ data: undefined });
	});

	it('adds a root via POST /storage-roots with the typed path and label', async () => {
		renderSection();
		await screen.findByText('/mnt/midia');

		const [path, label] = screen.getAllByRole('textbox');
		fireEvent.change(path!, { target: { value: '/mnt/novo' } });
		fireEvent.change(label!, { target: { value: 'Novo' } });
		fireEvent.click(screen.getByText('SETTINGS_STORAGE_ROOTS_ADD'));

		await waitFor(() =>
			expect(mockedApi.post).toHaveBeenCalledWith('/storage-roots', { path: '/mnt/novo', label: 'Novo' })
		);
	});

	it('toggles a non-primary root via PUT /storage-roots/:id', async () => {
		renderSection();
		await screen.findByText('/mnt/midia');

		const switches = screen.getAllByLabelText('SETTINGS_STORAGE_ROOTS_ENABLED');
		fireEvent.click(switches[1]!);

		await waitFor(() =>
			expect(mockedApi.put).toHaveBeenCalledWith('/storage-roots/2', { enabled: false })
		);
	});

	it('removes a non-primary root via DELETE /storage-roots/:id', async () => {
		renderSection();
		await screen.findByText('/mnt/midia');

		const removeButtons = screen.getAllByText('SETTINGS_STORAGE_ROOTS_REMOVE');
		fireEvent.click(removeButtons[1]!);

		await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/storage-roots/2'));
	});
});
