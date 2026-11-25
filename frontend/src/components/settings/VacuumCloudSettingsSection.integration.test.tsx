import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import { apiBase } from '@/service';
import VacuumCloudSettingsSection from './VacuumCloudSettingsSection';

jest.mock('@/service', () => ({
    apiBase: {
        get: jest.fn(),
        put: jest.fn(),
        post: jest.fn(),
    },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { get: jest.Mock; put: jest.Mock; post: jest.Mock };

const serverSettings = { access_id: 'abc123', region: 'eu', has_access_secret: true };

const renderSection = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <SnackbarProvider>
                <VacuumCloudSettingsSection />
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

describe('VacuumCloudSettingsSection (seam)', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({ data: serverSettings });
        mockedApi.put.mockResolvedValue({ data: serverSettings });
        mockedApi.post.mockResolvedValue({ data: undefined });
    });

    it('clicking Save issues PUT /vacuum/cloud-settings with the edited payload', async () => {
        renderSection();
        const accessIdField = await screen.findByLabelText('VACUUM_CLOUD_SETTINGS_ACCESS_ID_LABEL');
        await waitFor(() => expect(accessIdField).toHaveValue('abc123'));

        fireEvent.change(accessIdField, { target: { value: 'xyz789' } });
        fireEvent.change(screen.getByLabelText('VACUUM_CLOUD_SETTINGS_ACCESS_SECRET_LABEL'), {
            target: { value: 'k'.repeat(32) },
        });
        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_CLOUD_SETTINGS_SAVE' }));

        await waitFor(() => expect(mockedApi.put).toHaveBeenCalledTimes(1));
        expect(mockedApi.get).toHaveBeenCalledWith('/vacuum/cloud-settings');
        expect(mockedApi.put).toHaveBeenCalledWith('/vacuum/cloud-settings', {
            access_id: 'xyz789',
            access_secret: 'k'.repeat(32),
            region: 'eu',
        });
    });

    it('clicking Test cloud issues POST /vacuum/cloud-connection-test without a body', async () => {
        renderSection();
        const testButton = await screen.findByRole('button', {
            name: 'VACUUM_CLOUD_SETTINGS_TEST',
        });
        await waitFor(() => expect(testButton).toBeEnabled());

        fireEvent.click(testButton);

        await waitFor(() => expect(mockedApi.post).toHaveBeenCalledTimes(1));
        expect(mockedApi.post).toHaveBeenCalledWith('/vacuum/cloud-connection-test');
    });
});
