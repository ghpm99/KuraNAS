import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import LibrarySettingsSection from './LibrarySettingsSection';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real section + useLibrarySettings +
// service/libraries.ts run, so saving a category row asserts PUT
// /libraries/:category with the typed path the backend libraries handler decodes.
jest.mock('@/service', () => ({
	apiBase: { get: jest.fn(), put: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock };

const libraries = [{ category: 'images', path: '/data/Imagens' }];

const renderSection = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<LibrarySettingsSection />
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('components/settings/LibrarySettingsSection (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockResolvedValue({ data: libraries });
		mockedApi.put.mockResolvedValue({ data: { category: 'images', path: '/data/Fotos' } });
	});

	it('saving a category issues PUT /libraries/:category with the typed path', async () => {
		renderSection();
		// The hook always renders the four categories in order; images is first.
		await waitFor(() => expect(screen.getByDisplayValue('/data/Imagens')).toBeInTheDocument());
		const path = screen.getAllByRole('textbox')[0]!;

		fireEvent.change(path, { target: { value: '/data/Fotos' } });
		fireEvent.click(screen.getAllByText('SETTINGS_LIBRARY_SAVE')[0]!);

		await waitFor(() =>
			expect(mockedApi.put).toHaveBeenCalledWith('/libraries/images', { path: '/data/Fotos' })
		);
	});
});
