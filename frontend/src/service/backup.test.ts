jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
        put: jest.fn(),
    },
}));

import { apiBase } from './index';
import {
    getBackupPending,
    getBackupSettings,
    getBackupStatus,
    updateBackupSettings,
} from './backup';
import type { BackupSettings } from '@/types/backup';

const mockedApi = apiBase as unknown as {
    get: jest.Mock;
    put: jest.Mock;
};

const sampleSettings: BackupSettings = {
    enabled: true,
    destination_path: '/mnt/backup',
    retention_days: 30,
    interval_hours: 24,
};

describe('service/backup', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('loads the backup settings', async () => {
        mockedApi.get.mockResolvedValue({ data: sampleSettings });
        const result = await getBackupSettings();
        expect(mockedApi.get).toHaveBeenCalledWith('/backup/settings');
        expect(result).toEqual(sampleSettings);
    });

    it('saves the backup settings', async () => {
        mockedApi.put.mockResolvedValue({ data: sampleSettings });
        const result = await updateBackupSettings(sampleSettings);
        expect(mockedApi.put).toHaveBeenCalledWith('/backup/settings', sampleSettings);
        expect(result).toEqual(sampleSettings);
    });

    it('loads the backup status', async () => {
        const status = {
            enabled: true,
            has_run: true,
            status: 'completed',
            started_at: '2026-06-12T10:00:00Z',
            ended_at: '2026-06-12T10:05:00Z',
            last_error: '',
        };
        mockedApi.get.mockResolvedValue({ data: status });
        const result = await getBackupStatus();
        expect(mockedApi.get).toHaveBeenCalledWith('/backup/status');
        expect(result).toEqual(status);
    });

    it('loads the pending files count', async () => {
        mockedApi.get.mockResolvedValue({ data: { pending_files: 12 } });
        const result = await getBackupPending();
        expect(mockedApi.get).toHaveBeenCalledWith('/backup/pending');
        expect(result).toEqual({ pending_files: 12 });
    });
});
