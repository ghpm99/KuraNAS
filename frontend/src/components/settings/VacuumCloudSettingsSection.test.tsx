import { fireEvent, screen, waitFor } from '@testing-library/react';
import { createServerFailure } from '@/features/vacuum/testSupport/createServerFailure';
import { renderWithVacuumProviders } from '@/features/vacuum/testSupport/renderWithVacuumProviders';
import {
    getVacuumCloudSettings,
    testVacuumCloudConnection,
    updateVacuumCloudSettings,
} from '@/service/vacuum';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VacuumCloudSettingsSection from './VacuumCloudSettingsSection';

jest.mock(
    '@/components/i18n/provider/i18nContext',
    () => require('@/features/vacuum/testSupport/interpolatingI18nMock').interpolatingI18nMock
);
jest.mock('@/service/vacuum', () => ({
    ...jest.requireActual('@/service/vacuum'),
    getVacuumCloudSettings: jest.fn(),
    updateVacuumCloudSettings: jest.fn(),
    testVacuumCloudConnection: jest.fn(),
}));

const mockedGetSettings = getVacuumCloudSettings as jest.Mock;
const mockedUpdateSettings = updateVacuumCloudSettings as jest.Mock;
const mockedTestCloud = testVacuumCloudConnection as jest.Mock;

const storedSettings = { access_id: 'abc123', region: 'eu', has_access_secret: true };

const accessIdField = () => screen.getByLabelText('VACUUM_CLOUD_SETTINGS_ACCESS_ID_LABEL');
const accessSecretField = () => screen.getByLabelText('VACUUM_CLOUD_SETTINGS_ACCESS_SECRET_LABEL');
const saveButton = () => screen.getByRole('button', { name: 'VACUUM_CLOUD_SETTINGS_SAVE' });
const testButton = () => screen.getByRole('button', { name: 'VACUUM_CLOUD_SETTINGS_TEST' });

const renderLoadedSection = async (settings: unknown = storedSettings) => {
    mockedGetSettings.mockResolvedValue(settings);
    const rendered = renderWithVacuumProviders(<VacuumCloudSettingsSection />);
    await screen.findByLabelText('VACUUM_CLOUD_SETTINGS_ACCESS_ID_LABEL');
    return rendered;
};

describe('VacuumCloudSettingsSection', () => {
    beforeEach(() => {
        jest.resetAllMocks();
        mockedGetSettings.mockRejectedValue(new Error('backend absent'));
    });

    it('mounts without a backend', () => {
        expect(() => renderWithoutBackend(<VacuumCloudSettingsSection />)).not.toThrow();
    });

    it('renders empty fields for a partial payload', async () => {
        await renderLoadedSection({});

        expect(accessIdField()).toHaveValue('');
        expect(accessSecretField()).toHaveValue('');
        expect(
            screen.getByText('VACUUM_CLOUD_SETTINGS_ACCESS_SECRET_EMPTY_HELP')
        ).toBeInTheDocument();
        expect(
            screen.getByText('VACUUM_CLOUD_SETTINGS_TEST_NEEDS_SAVED_CREDENTIALS')
        ).toBeInTheDocument();
        expect(testButton()).toBeDisabled();
    });

    it('never shows the stored secret and explains that it is kept', async () => {
        await renderLoadedSection();

        expect(accessIdField()).toHaveValue('abc123');
        expect(accessSecretField()).toHaveValue('');
        expect(accessSecretField()).toHaveAttribute('type', 'password');
        expect(
            screen.getByText('VACUUM_CLOUD_SETTINGS_ACCESS_SECRET_STORED_HELP')
        ).toBeInTheDocument();
        expect(screen.getByText('VACUUM_CLOUD_SETTINGS_REGION_EU')).toBeInTheDocument();
    });

    it('shows the load error with the server message', async () => {
        mockedGetSettings.mockRejectedValue(createServerFailure(500, 'cloud unavailable'));

        renderWithVacuumProviders(<VacuumCloudSettingsSection />);

        expect(await screen.findByText('cloud unavailable')).toBeInTheDocument();
    });

    it('keeps save disabled until something changes and test disabled while unsaved', async () => {
        await renderLoadedSection();

        expect(saveButton()).toBeDisabled();
        expect(testButton()).toBeEnabled();

        fireEvent.change(accessIdField(), { target: { value: 'abc1234' } });

        expect(saveButton()).toBeEnabled();
        expect(testButton()).toBeDisabled();
        expect(screen.getByText('VACUUM_CLOUD_SETTINGS_TEST_SAVE_FIRST')).toBeInTheDocument();
    });

    it('saves once even on a double click and clears the secret field', async () => {
        mockedUpdateSettings.mockResolvedValue(storedSettings);
        await renderLoadedSection();

        fireEvent.change(accessSecretField(), { target: { value: 's'.repeat(32) } });
        fireEvent.click(saveButton());
        fireEvent.click(saveButton());

        await waitFor(() => expect(mockedUpdateSettings).toHaveBeenCalledTimes(1));
        expect(mockedUpdateSettings).toHaveBeenCalledWith({
            access_id: 'abc123',
            access_secret: 's'.repeat(32),
            region: 'eu',
        });
        await waitFor(() => expect(accessSecretField()).toHaveValue(''));
    });

    it('shows the server validation message when saving is refused', async () => {
        mockedUpdateSettings.mockRejectedValue(createServerFailure(400, 'segredo inválido'));
        await renderLoadedSection();

        fireEvent.change(accessSecretField(), { target: { value: 'short' } });
        fireEvent.click(saveButton());

        expect(await screen.findByText('segredo inválido')).toBeInTheDocument();
    });

    it('reports the cloud test success', async () => {
        mockedTestCloud.mockResolvedValue(undefined);
        await renderLoadedSection();

        fireEvent.click(testButton());

        expect(await screen.findByText('VACUUM_CLOUD_SETTINGS_TEST_SUCCEEDED')).toBeInTheDocument();
    });

    it('reports the cloud test failure with the server message', async () => {
        mockedTestCloud.mockRejectedValue(createServerFailure(502, 'nuvem recusou'));
        await renderLoadedSection();

        fireEvent.click(testButton());

        expect(await screen.findByText('nuvem recusou')).toBeInTheDocument();
    });
});
