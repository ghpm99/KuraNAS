import { apiBase } from '@/service';
import type {
    VacuumCommand,
    VacuumConsumablePart,
    VacuumCloudSettings,
    VacuumCloudSettingsChanges,
    VacuumConsumables,
    VacuumDirection,
    VacuumDiscovery,
    VacuumPreferences,
    VacuumPreferencesPatch,
    VacuumSettings,
    VacuumSettingsChanges,
    VacuumStatus,
    VacuumTotals,
} from '@/types/vacuum';

export const getVacuumSettings = async (): Promise<VacuumSettings> => {
    const response = await apiBase.get<VacuumSettings>('/vacuum/settings');
    return response.data;
};

export const updateVacuumSettings = async (
    changes: VacuumSettingsChanges
): Promise<VacuumSettings> => {
    const response = await apiBase.put<VacuumSettings>('/vacuum/settings', changes);
    return response.data;
};

export const testVacuumConnection = async (): Promise<void> => {
    await apiBase.post('/vacuum/connection-test');
};

export const discoverVacuumDevice = async (): Promise<VacuumDiscovery> => {
    const response = await apiBase.post<VacuumDiscovery>('/vacuum/discovery');
    return response.data;
};

export const getVacuumStatus = async (): Promise<VacuumStatus> => {
    const response = await apiBase.get<VacuumStatus>('/vacuum/status');
    return response.data;
};

export const sendVacuumCommand = async (command: VacuumCommand): Promise<void> => {
    await apiBase.post('/vacuum/commands', { command });
};

export const sendVacuumRemoteControl = async (direction: VacuumDirection): Promise<void> => {
    await apiBase.post('/vacuum/remote-control', { direction });
};

export const getVacuumPreferences = async (): Promise<VacuumPreferences> => {
    const response = await apiBase.get<VacuumPreferences>('/vacuum/preferences');
    return response.data;
};

export const updateVacuumPreferences = async (
    changes: VacuumPreferencesPatch
): Promise<VacuumPreferences> => {
    const response = await apiBase.patch<VacuumPreferences>('/vacuum/preferences', changes);
    return response.data;
};

export const getVacuumConsumables = async (): Promise<VacuumConsumables> => {
    const response = await apiBase.get<VacuumConsumables>('/vacuum/consumables');
    return response.data;
};

export const resetVacuumConsumable = async (part: VacuumConsumablePart): Promise<void> => {
    await apiBase.post(`/vacuum/consumables/${part}/reset`);
};

export const getVacuumTotals = async (): Promise<VacuumTotals> => {
    const response = await apiBase.get<VacuumTotals>('/vacuum/totals');
    return response.data;
};

export const getVacuumCloudSettings = async (): Promise<VacuumCloudSettings> => {
    const response = await apiBase.get<VacuumCloudSettings>('/vacuum/cloud-settings');
    return response.data;
};

export const updateVacuumCloudSettings = async (
    changes: VacuumCloudSettingsChanges
): Promise<VacuumCloudSettings> => {
    const response = await apiBase.put<VacuumCloudSettings>('/vacuum/cloud-settings', changes);
    return response.data;
};

export const testVacuumCloudConnection = async (): Promise<void> => {
    await apiBase.post('/vacuum/cloud-connection-test');
};
