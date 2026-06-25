import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import AccessControlSettingsSection from './AccessControlSettingsSection';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real section + useAccessControlSettings
// + service/accessControl.ts run, so each command button asserts the exact HTTP
// method/endpoint/payload the backend accesscontrol handler decodes.
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

const entries = [
	{ id: 1, cidr: '192.168.1.10/32', label: 'Notebook', enabled: true, created_at: '2026-01-01T00:00:00Z' },
];

const renderSection = () => {
	const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={client}>
			<SnackbarProvider>
				<AccessControlSettingsSection />
			</SnackbarProvider>
		</QueryClientProvider>
	);
};

describe('components/settings/AccessControlSettingsSection (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.get.mockImplementation((url: string) => {
			if (url === '/access-control/ips') return Promise.resolve({ data: entries });
			if (url === '/access-control/client-ip') return Promise.resolve({ data: { ip: '10.0.0.5' } });
			return Promise.reject(new Error(`unexpected GET ${url}`));
		});
		mockedApi.post.mockResolvedValue({ data: entries[0] });
		mockedApi.put.mockResolvedValue({ data: entries[0] });
		mockedApi.delete.mockResolvedValue({ data: undefined });
	});

	it('adds an entry via POST /access-control/ips with the typed cidr and label', async () => {
		renderSection();
		await screen.findByText('192.168.1.10/32');

		const [cidr, label] = screen.getAllByRole('textbox');
		fireEvent.change(cidr!, { target: { value: '10.0.0.0/24' } });
		fireEvent.change(label!, { target: { value: 'Rede' } });
		fireEvent.click(screen.getByText('SETTINGS_ACCESS_CONTROL_ADD'));

		await waitFor(() =>
			expect(mockedApi.post).toHaveBeenCalledWith('/access-control/ips', { cidr: '10.0.0.0/24', label: 'Rede' })
		);
	});

	it('adds the current device via POST /access-control/ips with the fetched IP', async () => {
		renderSection();
		await screen.findByText('192.168.1.10/32');

		fireEvent.click(screen.getByText('SETTINGS_ACCESS_CONTROL_ADD_CURRENT'));

		await waitFor(() =>
			expect(mockedApi.post).toHaveBeenCalledWith('/access-control/ips', { cidr: '10.0.0.5', label: '' })
		);
	});

	it('toggles an entry via PUT /access-control/ips/:id', async () => {
		renderSection();
		await screen.findByText('192.168.1.10/32');

		fireEvent.click(screen.getByLabelText('SETTINGS_ACCESS_CONTROL_ENABLED'));

		await waitFor(() =>
			expect(mockedApi.put).toHaveBeenCalledWith('/access-control/ips/1', { enabled: false })
		);
	});

	it('removes an entry via DELETE /access-control/ips/:id', async () => {
		renderSection();
		await screen.findByText('192.168.1.10/32');

		fireEvent.click(screen.getByText('SETTINGS_ACCESS_CONTROL_REMOVE'));

		await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/access-control/ips/1'));
	});
});
