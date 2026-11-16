import { fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { getVacuumStatus, sendVacuumCommand } from '@/service/vacuum';
import { createServerFailure } from './testSupport/createServerFailure';
import { renderWithVacuumProviders } from './testSupport/renderWithVacuumProviders';
import VacuumCommandsBlock from './VacuumCommandsBlock';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);
jest.mock('@/service/vacuum', () => ({
    ...jest.requireActual('@/service/vacuum'),
    getVacuumStatus: jest.fn(),
    sendVacuumCommand: jest.fn(),
}));

const mockedGetStatus = getVacuumStatus as jest.Mock;
const mockedSendCommand = sendVacuumCommand as jest.Mock;

const commandLabels = [
    'VACUUM_COMMAND_START',
    'VACUUM_COMMAND_PAUSE',
    'VACUUM_COMMAND_RESUME',
    'VACUUM_COMMAND_STOP',
    'VACUUM_COMMAND_DOCK',
    'VACUUM_COMMAND_LOCATE',
];

const enabledCommandLabels = (): string[] =>
    commandLabels.filter((label) => !screen.getByRole('button', { name: label }).hasAttribute('disabled'));

const renderWithState = async (state: string) => {
    mockedGetStatus.mockResolvedValue({ state });
    const rendered = renderWithVacuumProviders(<VacuumCommandsBlock />);
    await waitFor(() => expect(mockedGetStatus).toHaveBeenCalled());
    return rendered;
};

describe('VacuumCommandsBlock', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedGetStatus.mockRejectedValue(new Error('backend absent'));
    });

    it('mounts without a backend', () => {
        expect(() => renderWithoutBackend(<VacuumCommandsBlock />)).not.toThrow();
    });

    it('offers every command while the state is not known yet', () => {
        mockedGetStatus.mockReturnValue(new Promise(() => undefined));

        renderWithVacuumProviders(<VacuumCommandsBlock />);

        expect(enabledCommandLabels()).toEqual(commandLabels);
    });

    it('offers every command when the status payload is empty', async () => {
        await renderWithState('');
        await waitFor(() => expect(enabledCommandLabels()).toEqual(commandLabels));
    });

    it.each([
        ['idle', ['VACUUM_COMMAND_START', 'VACUUM_COMMAND_DOCK', 'VACUUM_COMMAND_LOCATE']],
        [
            'cleaning',
            [
                'VACUUM_COMMAND_PAUSE',
                'VACUUM_COMMAND_STOP',
                'VACUUM_COMMAND_DOCK',
                'VACUUM_COMMAND_LOCATE',
            ],
        ],
        [
            'paused',
            [
                'VACUUM_COMMAND_RESUME',
                'VACUUM_COMMAND_STOP',
                'VACUUM_COMMAND_DOCK',
                'VACUUM_COMMAND_LOCATE',
            ],
        ],
        ['returning', ['VACUUM_COMMAND_STOP', 'VACUUM_COMMAND_LOCATE']],
        ['charging', ['VACUUM_COMMAND_START', 'VACUUM_COMMAND_LOCATE']],
        ['charged', ['VACUUM_COMMAND_START', 'VACUUM_COMMAND_LOCATE']],
        ['emptying', ['VACUUM_COMMAND_LOCATE']],
        ['sleeping', ['VACUUM_COMMAND_START', 'VACUUM_COMMAND_DOCK', 'VACUUM_COMMAND_LOCATE']],
        ['remote_control', ['VACUUM_COMMAND_DOCK', 'VACUUM_COMMAND_LOCATE']],
        ['fault', ['VACUUM_COMMAND_DOCK', 'VACUUM_COMMAND_LOCATE']],
        ['unknown', commandLabels],
    ])('enables the right buttons in state %s', async (state, expectedLabels) => {
        await renderWithState(state);

        await waitFor(() => expect(enabledCommandLabels()).toEqual(expectedLabels));
    });

    it('sends the command, reports success and revalidates the status without guessing the state', async () => {
        mockedSendCommand.mockResolvedValue(undefined);
        await renderWithState('idle');
        await waitFor(() => expect(enabledCommandLabels()).toHaveLength(3));

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_COMMAND_START' }));

        await waitFor(() => expect(mockedSendCommand).toHaveBeenCalledWith('start'));
        expect(await screen.findByText('VACUUM_COMMAND_SENT')).toBeInTheDocument();
        await waitFor(() => expect(mockedGetStatus).toHaveBeenCalledTimes(2));
        expect(screen.getByRole('button', { name: 'VACUUM_COMMAND_START' })).toBeEnabled();
    });

    it('disables the button while the command is in flight and ignores a double click', async () => {
        let finishCommand: () => void = () => undefined;
        mockedSendCommand.mockReturnValue(
            new Promise<void>((resolve) => {
                finishCommand = resolve;
            })
        );
        await renderWithState('idle');
        const startButton = screen.getByRole('button', { name: 'VACUUM_COMMAND_START' });
        await waitFor(() => expect(startButton).toBeEnabled());

        fireEvent.click(startButton);
        fireEvent.click(startButton);

        await waitFor(() => expect(startButton).toBeDisabled());
        expect(mockedSendCommand).toHaveBeenCalledTimes(1);
        finishCommand();
        await waitFor(() => expect(startButton).toBeEnabled());
    });

    it('shows the server message as it came when the command fails', async () => {
        mockedSendCommand.mockRejectedValue(createServerFailure(502, 'robot unreachable'));
        await renderWithState('idle');
        await waitFor(() => expect(enabledCommandLabels()).toHaveLength(3));

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_COMMAND_LOCATE' }));

        expect(await screen.findByText('robot unreachable')).toBeInTheDocument();
    });

    it('shows the connection message when the command gets no server response', async () => {
        mockedSendCommand.mockRejectedValue(new Error('network'));
        await renderWithState('idle');
        await waitFor(() => expect(enabledCommandLabels()).toHaveLength(3));

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_COMMAND_LOCATE' }));

        expect(await screen.findByText('CONNECTION_BANNER_UNAVAILABLE')).toBeInTheDocument();
    });
});
