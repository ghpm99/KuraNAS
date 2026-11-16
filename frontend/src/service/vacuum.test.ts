jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
        post: jest.fn(),
        patch: jest.fn(),
    },
}));

import { apiBase } from './index';
import {
    getVacuumConsumables,
    getVacuumPreferences,
    getVacuumSettings,
    getVacuumStatus,
    getVacuumTotals,
    resetVacuumConsumable,
    sendVacuumCommand,
    sendVacuumRemoteControl,
    updateVacuumPreferences,
} from './vacuum';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    post: jest.Mock;
    patch: jest.Mock;
};

describe('service/vacuum', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('reads', () => {
        it('loads the settings', async () => {
            const settings = { address: '10.0.0.5', device_id: 'abc', has_local_key: true };
            mockedApi.get.mockResolvedValue({ data: settings });

            await expect(getVacuumSettings()).resolves.toEqual(settings);
            expect(mockedApi.get).toHaveBeenCalledWith('/vacuum/settings');
        });

        it('loads the status', async () => {
            const status = {
                state: 'cleaning',
                battery_percent: 80,
                cleaning_seconds: 120,
                cleaning_area: 6,
                fault_codes: [],
                read_at: '2026-11-16T10:00:00Z',
            };
            mockedApi.get.mockResolvedValue({ data: status });

            await expect(getVacuumStatus()).resolves.toEqual(status);
            expect(mockedApi.get).toHaveBeenCalledWith('/vacuum/status');
        });

        it('loads the preferences', async () => {
            const preferences = {
                suction: 'auto',
                water_level: null,
                volume: 4,
                is_do_not_disturb: false,
                is_carpet_boost: null,
                is_resume_after_charge: true,
                is_auto_empty: null,
                auto_empty_frequency: null,
            };
            mockedApi.get.mockResolvedValue({ data: preferences });

            await expect(getVacuumPreferences()).resolves.toEqual(preferences);
            expect(mockedApi.get).toHaveBeenCalledWith('/vacuum/preferences');
        });

        it('loads the consumables', async () => {
            const consumables = {
                items: [{ part: 'main_brush', seconds: 803013, max_seconds: 1080000 }],
            };
            mockedApi.get.mockResolvedValue({ data: consumables });

            await expect(getVacuumConsumables()).resolves.toEqual(consumables);
            expect(mockedApi.get).toHaveBeenCalledWith('/vacuum/consumables');
        });

        it('loads the totals', async () => {
            const totals = { cleaning_seconds: 7200, cleaning_area: 90, cleaning_count: 12 };
            mockedApi.get.mockResolvedValue({ data: totals });

            await expect(getVacuumTotals()).resolves.toEqual(totals);
            expect(mockedApi.get).toHaveBeenCalledWith('/vacuum/totals');
        });
    });

    describe('mutations', () => {
        it('posts the command with the command field only', async () => {
            mockedApi.post.mockResolvedValue({ status: 204 });

            await sendVacuumCommand('start');

            expect(mockedApi.post).toHaveBeenCalledTimes(1);
            expect(mockedApi.post).toHaveBeenCalledWith('/vacuum/commands', { command: 'start' });
        });

        it('posts the remote control direction with the direction field only', async () => {
            mockedApi.post.mockResolvedValue({ status: 204 });

            await sendVacuumRemoteControl('forward');

            expect(mockedApi.post).toHaveBeenCalledTimes(1);
            expect(mockedApi.post).toHaveBeenCalledWith('/vacuum/remote-control', {
                direction: 'forward',
            });
        });

        it('patches only the changed preference and returns the confirmed preferences', async () => {
            const confirmed = { suction: 'max', volume: 3 };
            mockedApi.patch.mockResolvedValue({ data: confirmed });

            await expect(updateVacuumPreferences({ suction: 'max' })).resolves.toEqual(confirmed);

            expect(mockedApi.patch).toHaveBeenCalledTimes(1);
            expect(mockedApi.patch).toHaveBeenCalledWith('/vacuum/preferences', {
                suction: 'max',
            });
        });

        it('posts the consumable reset to the part route without a body', async () => {
            mockedApi.post.mockResolvedValue({ status: 204 });

            await resetVacuumConsumable('main_brush');

            expect(mockedApi.post).toHaveBeenCalledTimes(1);
            expect(mockedApi.post).toHaveBeenCalledWith('/vacuum/consumables/main_brush/reset');
        });
    });
});
