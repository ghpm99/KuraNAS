import { fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { getVacuumConsumables, resetVacuumConsumable } from '@/service/vacuum';
import { createServerFailure } from './testSupport/createServerFailure';
import { renderWithVacuumProviders } from './testSupport/renderWithVacuumProviders';
import VacuumConsumablesBlock from './VacuumConsumablesBlock';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);
jest.mock('@/service/vacuum', () => ({
    ...jest.requireActual('@/service/vacuum'),
    getVacuumConsumables: jest.fn(),
    resetVacuumConsumable: jest.fn(),
}));

const mockedGetConsumables = getVacuumConsumables as jest.Mock;
const mockedResetConsumable = resetVacuumConsumable as jest.Mock;

const mainBrushOnly = { items: [{ part: 'main_brush', seconds: 803013, max_seconds: 1080000 }] };

const resetButtonName = 'VACUUM_RESET_PART_LABEL:VACUUM_PART_MAIN_BRUSH';

describe('VacuumConsumablesBlock', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedGetConsumables.mockRejectedValue(new Error('backend absent'));
    });

    it('mounts without a backend', () => {
        expect(() => renderWithoutBackend(<VacuumConsumablesBlock />)).not.toThrow();
    });

    it('shows the load error with the server message', async () => {
        mockedGetConsumables.mockRejectedValue(createServerFailure(502, 'robot offline'));

        renderWithVacuumProviders(<VacuumConsumablesBlock />);

        expect(await screen.findByText('robot offline')).toBeInTheDocument();
    });

    it('shows each reported part with hours to one decimal and no used or remaining wording', async () => {
        mockedGetConsumables.mockResolvedValue(mainBrushOnly);

        renderWithVacuumProviders(<VacuumConsumablesBlock />);

        expect(await screen.findByText('VACUUM_PART_MAIN_BRUSH')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_CONSUMABLE_HOURS:223.1,300.0')).toBeInTheDocument();
        expect(screen.queryByText('VACUUM_PART_FILTER')).not.toBeInTheDocument();
        expect(screen.queryByText('VACUUM_PART_SIDE_BRUSH')).not.toBeInTheDocument();
        expect(screen.queryByText('VACUUM_CONSUMABLES_EMPTY')).not.toBeInTheDocument();
    });

    it('says the robot reported no part when the list is empty', async () => {
        mockedGetConsumables.mockResolvedValue({ items: [] });

        renderWithVacuumProviders(<VacuumConsumablesBlock />);

        expect(await screen.findByText('VACUUM_CONSUMABLES_EMPTY')).toBeInTheDocument();
    });

    it('survives a payload without items', async () => {
        mockedGetConsumables.mockResolvedValue({});

        renderWithVacuumProviders(<VacuumConsumablesBlock />);

        expect(await screen.findByText('VACUUM_CONSUMABLES_EMPTY')).toBeInTheDocument();
    });

    it('survives a part with missing numbers and offers no reset for an unknown part', async () => {
        mockedGetConsumables.mockResolvedValue({ items: [{ part: 'hepa' }] });

        renderWithVacuumProviders(<VacuumConsumablesBlock />);

        expect(await screen.findByText('VACUUM_PART_UNKNOWN')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_CONSUMABLE_HOURS:0.0,0.0')).toBeInTheDocument();
        expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('asks for confirmation before resetting and sends nothing until confirmed', async () => {
        mockedGetConsumables.mockResolvedValue(mainBrushOnly);
        renderWithVacuumProviders(<VacuumConsumablesBlock />);

        fireEvent.click(await screen.findByRole('button', { name: resetButtonName }));

        expect(await screen.findByRole('dialog')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_RESET_CONFIRM_MESSAGE')).toBeInTheDocument();
        expect(mockedResetConsumable).not.toHaveBeenCalled();
    });

    it('does not reset when the confirmation is cancelled', async () => {
        mockedGetConsumables.mockResolvedValue(mainBrushOnly);
        renderWithVacuumProviders(<VacuumConsumablesBlock />);
        fireEvent.click(await screen.findByRole('button', { name: resetButtonName }));

        fireEvent.click(await screen.findByRole('button', { name: 'ACTION_CANCEL' }));

        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
        expect(mockedResetConsumable).not.toHaveBeenCalled();
    });

    it('resets the confirmed part, reports success and reloads the list', async () => {
        mockedGetConsumables.mockResolvedValue(mainBrushOnly);
        mockedResetConsumable.mockResolvedValue(undefined);
        renderWithVacuumProviders(<VacuumConsumablesBlock />);
        fireEvent.click(await screen.findByRole('button', { name: resetButtonName }));

        mockedGetConsumables.mockResolvedValue({
            items: [{ part: 'main_brush', seconds: 0, max_seconds: 1080000 }],
        });
        fireEvent.click(await screen.findByRole('button', { name: 'VACUUM_RESET_ACTION' }));

        await waitFor(() => expect(mockedResetConsumable).toHaveBeenCalledWith('main_brush'));
        expect(mockedResetConsumable).toHaveBeenCalledTimes(1);
        expect(await screen.findByText('VACUUM_RESET_DONE')).toBeInTheDocument();
        expect(await screen.findByText('VACUUM_CONSUMABLE_HOURS:0.0,300.0')).toBeInTheDocument();
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });

    it('shows the server message when the reset fails', async () => {
        mockedGetConsumables.mockResolvedValue(mainBrushOnly);
        mockedResetConsumable.mockRejectedValue(createServerFailure(502, 'robot unreachable'));
        renderWithVacuumProviders(<VacuumConsumablesBlock />);
        fireEvent.click(await screen.findByRole('button', { name: resetButtonName }));

        fireEvent.click(await screen.findByRole('button', { name: 'VACUUM_RESET_ACTION' }));

        expect(await screen.findByText('robot unreachable')).toBeInTheDocument();
    });

    it('ignores a second confirmation while the reset is in flight', async () => {
        mockedGetConsumables.mockResolvedValue(mainBrushOnly);
        mockedResetConsumable.mockReturnValue(new Promise(() => undefined));
        renderWithVacuumProviders(<VacuumConsumablesBlock />);
        fireEvent.click(await screen.findByRole('button', { name: resetButtonName }));
        const confirmButton = await screen.findByRole('button', { name: 'VACUUM_RESET_ACTION' });

        fireEvent.click(confirmButton);
        fireEvent.click(confirmButton);

        await waitFor(() => expect(confirmButton).toBeDisabled());
        expect(mockedResetConsumable).toHaveBeenCalledTimes(1);
    });
});
