import { apiBase } from '@/service';
import type {
    VacuumCommand,
    VacuumConsumablePart,
    VacuumConsumables,
    VacuumDirection,
    VacuumPreferences,
    VacuumPreferencesPatch,
    VacuumSettings,
    VacuumStatus,
    VacuumTotals,
} from '@/types/vacuum';

export const getVacuumSettings = async (): Promise<VacuumSettings> => {
    const response = await apiBase.get<VacuumSettings>('/vacuum/settings');
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
