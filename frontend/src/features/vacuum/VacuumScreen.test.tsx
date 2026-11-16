import { screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import {
    getVacuumConsumables,
    getVacuumPreferences,
    getVacuumSettings,
    getVacuumStatus,
    getVacuumTotals,
} from '@/service/vacuum';
import { createServerFailure } from './testSupport/createServerFailure';
import { renderWithVacuumProviders } from './testSupport/renderWithVacuumProviders';
import VacuumScreen from './VacuumScreen';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);
jest.mock('@/service/vacuum', () => ({
    ...jest.requireActual('@/service/vacuum'),
    getVacuumSettings: jest.fn(),
    getVacuumStatus: jest.fn(),
    getVacuumPreferences: jest.fn(),
    getVacuumConsumables: jest.fn(),
    getVacuumTotals: jest.fn(),
}));

const mockedGetSettings = getVacuumSettings as jest.Mock;
const mockedGetStatus = getVacuumStatus as jest.Mock;
const mockedGetPreferences = getVacuumPreferences as jest.Mock;
const mockedGetConsumables = getVacuumConsumables as jest.Mock;
const mockedGetTotals = getVacuumTotals as jest.Mock;

const configuredSettings = { address: '10.0.0.5', device_id: 'abc', has_local_key: true };

const blockHeadingKeys = [
    'VACUUM_STATUS_TITLE',
    'VACUUM_COMMANDS_TITLE',
    'VACUUM_REMOTE_TITLE',
    'VACUUM_PREFERENCES_TITLE',
    'VACUUM_CONSUMABLES_TITLE',
    'VACUUM_TOTALS_TITLE',
];

describe('VacuumScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        const backendAbsent = new Error('backend absent');
        mockedGetSettings.mockRejectedValue(backendAbsent);
        mockedGetStatus.mockRejectedValue(backendAbsent);
        mockedGetPreferences.mockRejectedValue(backendAbsent);
        mockedGetConsumables.mockRejectedValue(backendAbsent);
        mockedGetTotals.mockRejectedValue(backendAbsent);
    });

    it('mounts without a backend', () => {
        expect(() => renderWithoutBackend(<VacuumScreen />)).not.toThrow();
        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('VACUUM_PAGE_TITLE');
    });

    it('shows a loading state, and no command, while it does not know if the robot is configured', () => {
        mockedGetSettings.mockReturnValue(new Promise(() => undefined));
        mockedGetStatus.mockReturnValue(new Promise(() => undefined));

        renderWithVacuumProviders(<VacuumScreen />);

        expect(screen.getByRole('status')).toHaveTextContent('LOADING');
        expect(screen.queryByText('VACUUM_COMMANDS_TITLE')).not.toBeInTheDocument();
    });

    it('stacks the blocks in the specified order when the robot is configured', async () => {
        mockedGetSettings.mockResolvedValue(configuredSettings);
        mockedGetStatus.mockResolvedValue({ state: 'idle' });
        mockedGetPreferences.mockResolvedValue({});
        mockedGetConsumables.mockResolvedValue({ items: [] });
        mockedGetTotals.mockResolvedValue({});

        renderWithVacuumProviders(<VacuumScreen />);

        await screen.findByText('VACUUM_COMMANDS_TITLE');
        const renderedHeadings = screen
            .getAllByRole('heading', { level: 2 })
            .map((heading) => heading.textContent);
        expect(renderedHeadings).toEqual(blockHeadingKeys);
    });

    it('keeps the commands when another block fails', async () => {
        mockedGetSettings.mockResolvedValue(configuredSettings);
        mockedGetStatus.mockResolvedValue({ state: 'idle' });
        mockedGetTotals.mockRejectedValue(createServerFailure(502, 'robot offline'));

        renderWithVacuumProviders(<VacuumScreen />);

        expect(await screen.findAllByText('robot offline')).not.toHaveLength(0);
        expect(screen.getByRole('button', { name: 'VACUUM_COMMAND_START' })).toBeInTheDocument();
    });

    it('renders the blocks when the settings cannot be read, leaving the errors to each block', async () => {
        renderWithVacuumProviders(<VacuumScreen />);

        await waitFor(() => expect(screen.getByText('VACUUM_COMMANDS_TITLE')).toBeInTheDocument());
    });

    it('explains what is missing and offers no command when the device id is missing', async () => {
        mockedGetSettings.mockResolvedValue({ address: '', device_id: '', has_local_key: true });

        renderWithVacuumProviders(<VacuumScreen />);

        expect(await screen.findByText('VACUUM_NOT_CONFIGURED_DEVICE_ID')).toBeInTheDocument();
        expect(screen.queryByText('VACUUM_NOT_CONFIGURED_LOCAL_KEY')).not.toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'VACUUM_NOT_CONFIGURED_ACTION' })).toBeInTheDocument();
        expect(screen.queryByText('VACUUM_COMMANDS_TITLE')).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'VACUUM_COMMAND_START' })).not.toBeInTheDocument();
        expect(mockedGetStatus).not.toHaveBeenCalled();
    });

    it('explains the missing local key and offers no command', async () => {
        mockedGetSettings.mockResolvedValue({ address: '', device_id: 'abc', has_local_key: false });

        renderWithVacuumProviders(<VacuumScreen />);

        expect(await screen.findByText('VACUUM_NOT_CONFIGURED_LOCAL_KEY')).toBeInTheDocument();
        expect(screen.queryByText('VACUUM_COMMANDS_TITLE')).not.toBeInTheDocument();
    });

    it('treats an empty settings payload as not configured', async () => {
        mockedGetSettings.mockResolvedValue({});

        renderWithVacuumProviders(<VacuumScreen />);

        expect(await screen.findByText('VACUUM_NOT_CONFIGURED_DEVICE_ID')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_NOT_CONFIGURED_LOCAL_KEY')).toBeInTheDocument();
    });

    it('shows the not configured explanation when the status answers 409, without any command', async () => {
        mockedGetSettings.mockResolvedValue(configuredSettings);
        mockedGetStatus.mockRejectedValue(createServerFailure(409, 'not configured on the server'));

        renderWithVacuumProviders(<VacuumScreen />);

        expect(await screen.findByText('not configured on the server')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_NOT_CONFIGURED_DESCRIPTION')).toBeInTheDocument();
        expect(screen.queryByText('VACUUM_COMMANDS_TITLE')).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'VACUUM_COMMAND_START' })).not.toBeInTheDocument();
    });
});
