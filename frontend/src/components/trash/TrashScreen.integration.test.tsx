import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { MemoryRouter } from 'react-router-dom';
import TrashScreen from './TrashScreen';
import { apiBase } from '@/service';

// Seam test: only the axios instance is mocked, so the real TrashScreen, its
// hook and the real service/trash.ts run together. Each command button asserts
// the exact HTTP method + endpoint it hits — the "click → which endpoint" half
// of the flow the backend trash handler test verifies on the receiving end.
jest.mock('@/service', () => ({
	apiBase: {
		get: jest.fn(),
		post: jest.fn(),
		delete: jest.fn(),
		put: jest.fn(),
	},
}));

const mockedApi = apiBase as unknown as {
	get: jest.Mock;
	post: jest.Mock;
	delete: jest.Mock;
	put: jest.Mock;
};

const onePage = {
	items: [
		{ id: 1, original_path: '/data/docs/relatorio.pdf', size: 2048, deleted_at: '2026-06-11T10:00:00Z' },
	],
	pagination: { page: 1, page_size: 15, has_next: false, has_prev: false },
};

const renderScreen = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<MemoryRouter>
					<TrashScreen />
				</MemoryRouter>
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('TrashScreen (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockImplementation((url: string) => {
			if (url === '/trash') return Promise.resolve({ data: onePage });
			if (url === '/trash/retention') return Promise.resolve({ data: { days: 30 } });
			return Promise.reject(new Error(`unexpected GET ${url}`));
		});
		mockedApi.post.mockResolvedValue({ data: undefined });
		mockedApi.delete.mockResolvedValue({ data: undefined });
		mockedApi.put.mockResolvedValue({ data: { days: 7 } });
	});

	it('restores an item via POST /trash/:id/restore', async () => {
		renderScreen();
		await screen.findByText('/data/docs/relatorio.pdf');

		fireEvent.click(screen.getByRole('button', { name: /TRASH_RESTORE_BUTTON/ }));

		await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/trash/1/restore'));
	});

	it('permanently deletes an item via DELETE /trash/:id', async () => {
		const confirmSpy = jest.spyOn(window, 'confirm').mockReturnValue(true);
		renderScreen();
		await screen.findByText('/data/docs/relatorio.pdf');

		fireEvent.click(screen.getByRole('button', { name: /TRASH_DELETE_FOREVER_BUTTON/ }));

		await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/trash/1'));
		confirmSpy.mockRestore();
	});

	it('empties the trash via DELETE /trash', async () => {
		const confirmSpy = jest.spyOn(window, 'confirm').mockReturnValue(true);
		renderScreen();
		await screen.findByText('/data/docs/relatorio.pdf');

		fireEvent.click(screen.getByRole('button', { name: /TRASH_EMPTY_BUTTON/ }));

		await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/trash'));
		confirmSpy.mockRestore();
	});

	it('saves the retention policy on blur via PUT /trash/retention', async () => {
		renderScreen();
		const input = await screen.findByLabelText('TRASH_RETENTION_LABEL');

		fireEvent.change(input, { target: { value: '7' } });
		fireEvent.blur(input);

		await waitFor(() =>
			expect(mockedApi.put).toHaveBeenCalledWith('/trash/retention', { days: 7 })
		);
	});
});
