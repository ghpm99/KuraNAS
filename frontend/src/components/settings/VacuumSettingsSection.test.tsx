import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { createServerFailure } from '@/features/vacuum/testSupport/createServerFailure';
import { renderWithVacuumProviders } from '@/features/vacuum/testSupport/renderWithVacuumProviders';
import { vacuumQueryKeys } from '@/features/vacuum/vacuumQueryKeys';
import {
    discoverVacuumDevice,
    getVacuumSettings,
    testVacuumConnection,
    updateVacuumSettings,
} from '@/service/vacuum';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VacuumSettingsSection from './VacuumSettingsSection';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('@/features/vacuum/testSupport/interpolatingI18nMock').interpolatingI18nMock
);
jest.mock('@/service/vacuum', () => ({
    ...jest.requireActual('@/service/vacuum'),
    getVacuumSettings: jest.fn(),
    updateVacuumSettings: jest.fn(),
    testVacuumConnection: jest.fn(),
    discoverVacuumDevice: jest.fn(),
}));

const mockedGetSettings = getVacuumSettings as jest.Mock;
const mockedUpdateSettings = updateVacuumSettings as jest.Mock;
const mockedTestConnection = testVacuumConnection as jest.Mock;
const mockedDiscover = discoverVacuumDevice as jest.Mock;

const storedSettings = { address: '192.168.18.95', device_id: 'abc', has_local_key: true };
const typedLocalKey = '0123456789abcdef';

const deviceIdField = () => screen.getByLabelText('VACUUM_SETTINGS_DEVICE_ID_LABEL');
const localKeyField = () => screen.getByLabelText('VACUUM_SETTINGS_LOCAL_KEY_LABEL');
const addressField = () => screen.getByLabelText('VACUUM_SETTINGS_ADDRESS_LABEL');
const saveButton = () => screen.getByRole('button', { name: 'VACUUM_SETTINGS_SAVE' });
const discoverButton = () => screen.getByRole('button', { name: 'VACUUM_SETTINGS_DISCOVER' });
const testButton = () => screen.getByRole('button', { name: 'VACUUM_SETTINGS_TEST' });

const resolveWithServerState = (serverSettings: unknown) => {
    mockedGetSettings.mockResolvedValue(serverSettings);
    return serverSettings;
};

const renderLoadedSection = async (settings: unknown = storedSettings) => {
    mockedGetSettings.mockResolvedValue(settings);
    const rendered = renderWithVacuumProviders(<VacuumSettingsSection />);
    await screen.findByLabelText('VACUUM_SETTINGS_DEVICE_ID_LABEL');
    return rendered;
};

describe('VacuumSettingsSection', () => {
    beforeEach(() => {
        jest.resetAllMocks();
        mockedGetSettings.mockRejectedValue(new Error('backend absent'));
    });

    it('mounts without a backend', () => {
        expect(() => renderWithoutBackend(<VacuumSettingsSection />)).not.toThrow();
    });

    it('shows the loading state on the first load', () => {
        mockedGetSettings.mockReturnValue(new Promise(() => undefined));

        renderWithVacuumProviders(<VacuumSettingsSection />);

        expect(screen.getByRole('progressbar')).toBeInTheDocument();
        expect(screen.queryByLabelText('VACUUM_SETTINGS_DEVICE_ID_LABEL')).not.toBeInTheDocument();
    });

    it('renders empty fields for a partial payload', async () => {
        await renderLoadedSection({});

        expect(deviceIdField()).toHaveValue('');
        expect(addressField()).toHaveValue('');
        expect(localKeyField()).toHaveValue('');
        expect(screen.getByText('VACUUM_SETTINGS_LOCAL_KEY_EMPTY_HELP')).toBeInTheDocument();
    });

    it('renders empty fields for a null payload', async () => {
        await renderLoadedSection(null);

        expect(deviceIdField()).toHaveValue('');
        expect(screen.getByText('VACUUM_SETTINGS_HELP_WHERE')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_SETTINGS_HELP_KEY_CHANGES')).toBeInTheDocument();
    });

    it('never pre-fills the local key and explains that a stored key is kept', async () => {
        await renderLoadedSection();

        expect(deviceIdField()).toHaveValue('abc');
        expect(addressField()).toHaveValue('192.168.18.95');
        expect(localKeyField()).toHaveValue('');
        expect(localKeyField()).toHaveAttribute('type', 'password');
        expect(screen.getByText('VACUUM_SETTINGS_LOCAL_KEY_STORED_HELP')).toBeInTheDocument();
    });

    it('shows the load error with a retry and never an empty form', async () => {
        mockedGetSettings.mockRejectedValue(createServerFailure(500, 'settings unavailable'));

        renderWithVacuumProviders(<VacuumSettingsSection />);

        expect(await screen.findByText('settings unavailable')).toBeInTheDocument();
        expect(screen.queryByLabelText('VACUUM_SETTINGS_DEVICE_ID_LABEL')).not.toBeInTheDocument();

        mockedGetSettings.mockResolvedValue(storedSettings);
        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

        expect(await screen.findByLabelText('VACUUM_SETTINGS_DEVICE_ID_LABEL')).toHaveValue('abc');
    });

    it('falls back to the section load message when the failure has no server message', async () => {
        mockedGetSettings.mockRejectedValue(new Error('network down'));

        renderWithVacuumProviders(<VacuumSettingsSection />);

        expect(await screen.findByText('VACUUM_SETTINGS_LOAD_ERROR')).toBeInTheDocument();
    });

    it('keeps save disabled until something changes', async () => {
        await renderLoadedSection();

        expect(saveButton()).toBeDisabled();

        fireEvent.change(deviceIdField(), { target: { value: 'abcd' } });
        expect(saveButton()).toBeEnabled();

        fireEvent.change(deviceIdField(), { target: { value: 'abc' } });
        expect(saveButton()).toBeDisabled();
    });

    it('sends an empty local key when the key field was left empty', async () => {
        mockedUpdateSettings.mockResolvedValue(storedSettings);
        await renderLoadedSection();

        fireEvent.change(addressField(), { target: { value: '192.168.18.96' } });
        fireEvent.click(saveButton());

        await waitFor(() => expect(mockedUpdateSettings).toHaveBeenCalledTimes(1));
        expect(mockedUpdateSettings).toHaveBeenCalledWith({
            address: '192.168.18.96',
            device_id: 'abc',
            local_key: '',
        });
    });

    it('sends the typed key and clears the field and the page afterwards', async () => {
        mockedUpdateSettings.mockResolvedValue(storedSettings);
        await renderLoadedSection();

        fireEvent.change(localKeyField(), { target: { value: typedLocalKey } });
        fireEvent.click(saveButton());

        await waitFor(() => expect(mockedUpdateSettings).toHaveBeenCalledTimes(1));
        expect(mockedUpdateSettings).toHaveBeenCalledWith({
            address: '192.168.18.95',
            device_id: 'abc',
            local_key: typedLocalKey,
        });
        await waitFor(() => expect(localKeyField()).toHaveValue(''));
        expect(document.body.innerHTML).not.toContain(typedLocalKey);
        expect(saveButton()).toBeDisabled();
    });

    it('shows the confirmed values and a success notice after saving', async () => {
        await renderLoadedSection({ address: '', device_id: '', has_local_key: false });
        mockedUpdateSettings.mockImplementation(async () =>
            resolveWithServerState({ address: '10.0.0.7', device_id: 'xyz', has_local_key: true })
        );

        fireEvent.change(deviceIdField(), { target: { value: 'xyz' } });
        fireEvent.change(localKeyField(), { target: { value: typedLocalKey } });
        fireEvent.click(saveButton());

        expect(await screen.findByText('VACUUM_SETTINGS_SAVED')).toBeInTheDocument();
        await waitFor(() => expect(addressField()).toHaveValue('10.0.0.7'));
        expect(deviceIdField()).toHaveValue('xyz');
        expect(screen.getByText('VACUUM_SETTINGS_LOCAL_KEY_STORED_HELP')).toBeInTheDocument();
    });

    it('sends a double click on save only once and disables the button meanwhile', async () => {
        let resolveSave: () => void = () => undefined;
        await renderLoadedSection();
        mockedUpdateSettings.mockReturnValue(
            new Promise((resolve) => {
                resolveSave = () =>
                    resolve(resolveWithServerState({ ...storedSettings, device_id: 'abcd' }));
            })
        );

        fireEvent.change(deviceIdField(), { target: { value: 'abcd' } });
        fireEvent.click(saveButton());
        fireEvent.click(saveButton());

        await waitFor(() => expect(screen.getByRole('button', { name: 'SAVING' })).toBeDisabled());
        expect(mockedUpdateSettings).toHaveBeenCalledTimes(1);

        await act(async () => resolveSave());
        await waitFor(() => expect(saveButton()).toBeDisabled());
    });

    it('invalidates every robot query after saving', async () => {
        mockedUpdateSettings.mockResolvedValue(storedSettings);
        const { queryClient } = await renderLoadedSection();
        const invalidateQueries = jest.spyOn(queryClient, 'invalidateQueries');

        fireEvent.change(deviceIdField(), { target: { value: 'abcd' } });
        fireEvent.click(saveButton());

        await waitFor(() => expect(invalidateQueries).toHaveBeenCalledTimes(5));
        Object.values(vacuumQueryKeys).forEach((queryKey) =>
            expect(invalidateQueries).toHaveBeenCalledWith({ queryKey })
        );
    });

    it('shows the server validation message next to the form and keeps the typed key out of the dom', async () => {
        mockedUpdateSettings.mockRejectedValue(
            createServerFailure(400, 'local key must have 16 characters')
        );
        await renderLoadedSection();

        fireEvent.change(localKeyField(), { target: { value: 'short' } });
        fireEvent.click(saveButton());

        expect(await screen.findByText('local key must have 16 characters')).toBeInTheDocument();
        expect(saveButton()).toBeEnabled();
    });

    it('falls back to the connection message when saving fails without a server message', async () => {
        mockedUpdateSettings.mockRejectedValue(new Error('network down'));
        await renderLoadedSection();

        fireEvent.change(deviceIdField(), { target: { value: 'abcd' } });
        fireEvent.click(saveButton());

        expect(await screen.findByText('CONNECTION_BANNER_UNAVAILABLE')).toBeInTheDocument();
    });

    describe('connection actions', () => {
        it('stay disabled with an explanation when no key is stored', async () => {
            await renderLoadedSection({ address: '', device_id: 'abc', has_local_key: false });

            expect(discoverButton()).toBeDisabled();
            expect(testButton()).toBeDisabled();
            expect(
                screen.getByText('VACUUM_SETTINGS_CONNECTION_NEEDS_SAVED_CREDENTIALS')
            ).toBeInTheDocument();
        });

        it('stay disabled with an explanation when no device id is stored', async () => {
            await renderLoadedSection({ address: '', device_id: '', has_local_key: true });

            expect(discoverButton()).toBeDisabled();
            expect(testButton()).toBeDisabled();
        });

        it('stay disabled while the form has unsaved changes', async () => {
            await renderLoadedSection();
            expect(discoverButton()).toBeEnabled();
            expect(testButton()).toBeEnabled();

            fireEvent.change(localKeyField(), { target: { value: typedLocalKey } });

            expect(discoverButton()).toBeDisabled();
            expect(testButton()).toBeDisabled();
            expect(screen.getByText('VACUUM_SETTINGS_CONNECTION_SAVE_FIRST')).toBeInTheDocument();
        });

        it('fills the address with the discovered one and announces it', async () => {
            await renderLoadedSection({ address: '', device_id: 'abc', has_local_key: true });
            mockedDiscover.mockImplementation(async () => {
                resolveWithServerState({ ...storedSettings, address: '192.168.18.200' });
                return { address: '192.168.18.200' };
            });

            fireEvent.click(discoverButton());

            expect(
                await screen.findByText('VACUUM_SETTINGS_DISCOVERY_FOUND:192.168.18.200')
            ).toBeInTheDocument();
            await waitFor(() => expect(addressField()).toHaveValue('192.168.18.200'));
            expect(discoverButton()).toBeEnabled();
        });

        it('shows a spinner on the discover button only while it waits', async () => {
            let resolveDiscovery: (discovery: unknown) => void = () => undefined;
            mockedDiscover.mockReturnValue(
                new Promise((resolve) => {
                    resolveDiscovery = resolve;
                })
            );
            await renderLoadedSection();

            fireEvent.click(discoverButton());

            await waitFor(() => expect(discoverButton()).toBeDisabled());
            expect(testButton()).toBeEnabled();
            expect(saveButton()).toBeDisabled();
            expect(screen.getByRole('progressbar')).toBeInTheDocument();

            await act(async () => resolveDiscovery({ address: '10.0.0.9' }));
            await waitFor(() => expect(discoverButton()).toBeEnabled());
            expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
        });

        it('handles a discovery answer without an address', async () => {
            mockedDiscover.mockResolvedValue({});
            await renderLoadedSection();

            fireEvent.click(discoverButton());

            await waitFor(() => expect(discoverButton()).toBeEnabled());
            expect(addressField()).toHaveValue('192.168.18.95');
        });

        it('shows the not-found message as the server sent it', async () => {
            mockedDiscover.mockRejectedValue(createServerFailure(404, 'no device accepted the key'));
            await renderLoadedSection();

            fireEvent.click(discoverButton());

            expect(await screen.findByText('no device accepted the key')).toBeInTheDocument();
            expect(addressField()).toHaveValue('192.168.18.95');
        });

        it('shows the not-configured message as the server sent it', async () => {
            mockedDiscover.mockRejectedValue(createServerFailure(409, 'robot not configured'));
            await renderLoadedSection();

            fireEvent.click(discoverButton());

            expect(await screen.findByText('robot not configured')).toBeInTheDocument();
        });

        it('shows a visible success after a passing connection test', async () => {
            mockedTestConnection.mockResolvedValue(undefined);
            await renderLoadedSection();

            fireEvent.click(testButton());

            expect(await screen.findByText('VACUUM_SETTINGS_TEST_SUCCEEDED')).toBeInTheDocument();
        });

        it('shows the unreachable and the rejected-key messages as the server sent them', async () => {
            await renderLoadedSection();

            mockedTestConnection.mockRejectedValueOnce(createServerFailure(502, 'robot unreachable'));
            fireEvent.click(testButton());
            expect(await screen.findByText('robot unreachable')).toBeInTheDocument();

            mockedTestConnection.mockRejectedValueOnce(createServerFailure(502, 'key rejected'));
            fireEvent.click(testButton());
            expect(await screen.findByText('key rejected')).toBeInTheDocument();
            expect(screen.queryByText('robot unreachable')).not.toBeInTheDocument();
        });

        it('disables only the test button while testing', async () => {
            mockedTestConnection.mockReturnValue(new Promise(() => undefined));
            await renderLoadedSection();

            fireEvent.click(testButton());

            await waitFor(() => expect(testButton()).toBeDisabled());
            expect(discoverButton()).toBeEnabled();
        });
    });
});
