import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import EmailSettingsSection from './EmailSettingsSection';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real section + useEmailSettings +
// service/email.ts run, so each command asserts the exact endpoint/payload the
// backend email handlers decode.
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

const accounts = [
	{ id: 1, provider: 'google', address: 'owner@gmail.com', status: 'linked', sync_enabled: true },
];

const renderSection = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<EmailSettingsSection />
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('components/settings/EmailSettingsSection (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockImplementation((url: string) => {
			if (url === '/email/accounts') return Promise.resolve({ data: accounts });
			if (url === '/email/settings/provider') return Promise.resolve({ data: { provider: 'ollama' } });
			return Promise.reject(new Error(`unexpected GET ${url}`));
		});
		mockedApi.post.mockResolvedValue({ data: { auth_url: 'https://accounts.google.com/o/oauth2' } });
		mockedApi.put.mockResolvedValue({ data: { provider: 'openai' } });
		mockedApi.delete.mockResolvedValue({ data: { message: 'removida' } });
		jest.spyOn(window, 'open').mockReturnValue(null);
	});

	it('linking Google issues POST /email/accounts/google/auth-url', async () => {
		renderSection();
		await screen.findByText('owner@gmail.com');

		fireEvent.click(screen.getByText('SETTINGS_EMAIL_ADD_GOOGLE'));

		await waitFor(() =>
			expect(mockedApi.post).toHaveBeenCalledWith('/email/accounts/google/auth-url')
		);
	});

	it('changing the AI provider issues PUT /email/settings/provider', async () => {
		renderSection();
		await screen.findByText('owner@gmail.com');

		fireEvent.mouseDown(screen.getByRole('combobox'));
		const listbox = await screen.findByRole('listbox');
		fireEvent.click(within(listbox).getByText('SETTINGS_EMAIL_AI_PROVIDER_OPENAI'));

		await waitFor(() =>
			expect(mockedApi.put).toHaveBeenCalledWith('/email/settings/provider', { provider: 'openai' })
		);
	});

	it('toggling account sync issues PUT /email/accounts/:id/sync-enabled', async () => {
		renderSection();
		await screen.findByText('owner@gmail.com');

		fireEvent.click(screen.getByLabelText('SETTINGS_EMAIL_SYNC_ENABLED'));

		await waitFor(() =>
			expect(mockedApi.put).toHaveBeenCalledWith('/email/accounts/1/sync-enabled', { sync_enabled: false })
		);
	});

	it('removing an account issues DELETE /email/accounts/:id', async () => {
		renderSection();
		await screen.findByText('owner@gmail.com');

		fireEvent.click(screen.getByText('SETTINGS_EMAIL_REMOVE'));

		await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/email/accounts/1'));
	});
});
