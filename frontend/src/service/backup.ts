import { apiBase } from '@/service';
import type { BackupPending, BackupSettings, BackupStatus } from '@/types/backup';

export const getBackupSettings = async (): Promise<BackupSettings> => {
	const response = await apiBase.get<BackupSettings>('/backup/settings');
	return response.data;
};

export const updateBackupSettings = async (settings: BackupSettings): Promise<BackupSettings> => {
	const response = await apiBase.put<BackupSettings>('/backup/settings', settings);
	return response.data;
};

export const getBackupStatus = async (): Promise<BackupStatus> => {
	const response = await apiBase.get<BackupStatus>('/backup/status');
	return response.data;
};

export const getBackupPending = async (): Promise<BackupPending> => {
	const response = await apiBase.get<BackupPending>('/backup/pending');
	return response.data;
};
