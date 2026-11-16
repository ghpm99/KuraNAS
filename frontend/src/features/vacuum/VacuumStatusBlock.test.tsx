import { act, screen, waitFor } from '@testing-library/react';
import { focusManager } from '@tanstack/react-query';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { getVacuumStatus } from '@/service/vacuum';
import type { VacuumStatus } from '@/types/vacuum';
import { createServerFailure } from './testSupport/createServerFailure';
import { renderWithVacuumProviders } from './testSupport/renderWithVacuumProviders';
import VacuumStatusBlock from './VacuumStatusBlock';
import { vacuumStatusPollingIntervalMs } from './useVacuumStatus';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);
jest.mock('@/service/vacuum', () => ({
    ...jest.requireActual('@/service/vacuum'),
    getVacuumStatus: jest.fn(),
}));

const mockedGetStatus = getVacuumStatus as jest.Mock;

const cleaningStatus: VacuumStatus = {
    state: 'cleaning',
    battery_percent: 76,
    cleaning_seconds: 754,
    cleaning_area: 9,
    fault_codes: [],
    read_at: '2026-11-16T10:00:00Z',
};

describe('VacuumStatusBlock', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    afterEach(() => {
        focusManager.setFocused(undefined);
        jest.useRealTimers();
    });

    it('mounts without a backend and shows the loading state', () => {
        mockedGetStatus.mockReturnValue(new Promise(() => undefined));

        expect(() => renderWithoutBackend(<VacuumStatusBlock />)).not.toThrow();
        expect(screen.getByRole('status')).toHaveTextContent('LOADING');
    });

    it('shows the load error with the server message when there is no previous state', async () => {
        mockedGetStatus.mockRejectedValue(createServerFailure(502, 'robot offline'));

        renderWithVacuumProviders(<VacuumStatusBlock />);

        expect(await screen.findByText('robot offline')).toBeInTheDocument();
    });

    it('shows state, battery, elapsed time and area while cleaning', async () => {
        mockedGetStatus.mockResolvedValue(cleaningStatus);

        renderWithVacuumProviders(<VacuumStatusBlock />);

        expect(await screen.findByText('VACUUM_STATE_CLEANING')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_STATUS_BATTERY:76')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_STATUS_ELAPSED:12:34')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_STATUS_AREA:9')).toBeInTheDocument();
    });

    it('hides elapsed time and area when the robot is docked', async () => {
        mockedGetStatus.mockResolvedValue({ ...cleaningStatus, state: 'charging' });

        renderWithVacuumProviders(<VacuumStatusBlock />);

        expect(await screen.findByText('VACUUM_STATE_CHARGING')).toBeInTheDocument();
        expect(screen.queryByText(/VACUUM_STATUS_ELAPSED/)).not.toBeInTheDocument();
        expect(screen.queryByText(/VACUUM_STATUS_AREA/)).not.toBeInTheDocument();
    });

    it('shows an unrecognized state as unknown', async () => {
        mockedGetStatus.mockResolvedValue({ ...cleaningStatus, state: 'warp_drive' });

        renderWithVacuumProviders(<VacuumStatusBlock />);

        expect(await screen.findByText('VACUUM_STATE_UNKNOWN')).toBeInTheDocument();
    });

    it('survives an empty payload', async () => {
        mockedGetStatus.mockResolvedValue({});

        renderWithVacuumProviders(<VacuumStatusBlock />);

        expect(await screen.findByText('VACUUM_STATE_UNKNOWN')).toBeInTheDocument();
        expect(screen.queryByText(/VACUUM_STATUS_BATTERY/)).not.toBeInTheDocument();
    });

    it('shows paused sessions with zeroed counters when the payload omits them', async () => {
        mockedGetStatus.mockResolvedValue({ state: 'paused' });

        renderWithVacuumProviders(<VacuumStatusBlock />);

        expect(await screen.findByText('VACUUM_STATUS_ELAPSED:0:00')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_STATUS_AREA:0')).toBeInTheDocument();
    });

    it('highlights every fault code', async () => {
        mockedGetStatus.mockResolvedValue({ ...cleaningStatus, state: 'fault', fault_codes: [5, 9] });

        renderWithVacuumProviders(<VacuumStatusBlock />);

        expect(await screen.findByText('VACUUM_FAULT_CODE:5')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_FAULT_CODE:9')).toBeInTheDocument();
    });

    it('survives a payload without fault codes', async () => {
        mockedGetStatus.mockResolvedValue({ state: 'fault' });

        renderWithVacuumProviders(<VacuumStatusBlock />);

        expect(await screen.findByText('VACUUM_STATE_FAULT')).toBeInTheDocument();
        expect(screen.queryByText(/VACUUM_FAULT_CODE/)).not.toBeInTheDocument();
    });

    it('polls every five seconds and keeps the last state marked stale when a poll fails', async () => {
        jest.useFakeTimers();
        mockedGetStatus.mockResolvedValueOnce(cleaningStatus);
        renderWithVacuumProviders(<VacuumStatusBlock />);
        expect(await screen.findByText('VACUUM_STATE_CLEANING')).toBeInTheDocument();
        expect(screen.queryByText(/VACUUM_STATUS_STALE/)).not.toBeInTheDocument();

        mockedGetStatus.mockRejectedValueOnce(new Error('timeout'));
        await act(async () => {
            await jest.advanceTimersByTimeAsync(vacuumStatusPollingIntervalMs);
        });

        expect(mockedGetStatus).toHaveBeenCalledTimes(2);
        expect(screen.getByText('VACUUM_STATE_CLEANING')).toBeInTheDocument();
        expect(screen.getByText(/VACUUM_STATUS_STALE/)).toBeInTheDocument();
        expect(screen.queryByText('LOADING')).not.toBeInTheDocument();
    });

    it('does not swap the content for a loading state while revalidating', async () => {
        jest.useFakeTimers();
        mockedGetStatus.mockResolvedValueOnce(cleaningStatus);
        renderWithVacuumProviders(<VacuumStatusBlock />);
        await screen.findByText('VACUUM_STATE_CLEANING');

        mockedGetStatus.mockReturnValueOnce(new Promise(() => undefined));
        await act(async () => {
            await jest.advanceTimersByTimeAsync(vacuumStatusPollingIntervalMs);
        });

        expect(screen.getByText('VACUUM_STATE_CLEANING')).toBeInTheDocument();
        expect(screen.queryByText('LOADING')).not.toBeInTheDocument();
    });

    it('stops polling while the tab is hidden and resumes when it is visible again', async () => {
        jest.useFakeTimers();
        mockedGetStatus.mockResolvedValue(cleaningStatus);
        renderWithVacuumProviders(<VacuumStatusBlock />);
        await screen.findByText('VACUUM_STATE_CLEANING');
        expect(mockedGetStatus).toHaveBeenCalledTimes(1);

        focusManager.setFocused(false);
        await act(async () => {
            await jest.advanceTimersByTimeAsync(vacuumStatusPollingIntervalMs * 3);
        });
        expect(mockedGetStatus).toHaveBeenCalledTimes(1);

        focusManager.setFocused(true);
        await act(async () => {
            await jest.advanceTimersByTimeAsync(vacuumStatusPollingIntervalMs);
        });
        await waitFor(() => expect(mockedGetStatus.mock.calls.length).toBeGreaterThan(1));
    });

    it('stops polling once the robot is reported as not configured', async () => {
        jest.useFakeTimers();
        mockedGetStatus.mockRejectedValue(createServerFailure(409, 'not configured'));
        renderWithVacuumProviders(<VacuumStatusBlock />);
        await screen.findByText('not configured');

        await act(async () => {
            await jest.advanceTimersByTimeAsync(vacuumStatusPollingIntervalMs * 3);
        });

        expect(mockedGetStatus).toHaveBeenCalledTimes(1);
    });
});
