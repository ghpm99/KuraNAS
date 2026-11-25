import { apiBase } from '@/service';
import type {
    VacuumCommand,
    VacuumConsumablePart,
    VacuumCloudSettings,
    VacuumCloudSettingsChanges,
    VacuumConsumables,
    VacuumDirection,
    VacuumDiscovery,
    VacuumMap,
    VacuumPath,
    VacuumPreferences,
    VacuumPreferencesPatch,
    VacuumRoom,
    VacuumRoomCleaningRequest,
    VacuumRoomRename,
    VacuumRooms,
    VacuumSettings,
    VacuumSettingsChanges,
    VacuumSpotCleaningRequest,
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

export const getVacuumMap = async (): Promise<VacuumMap> => {
    const response = await apiBase.get<VacuumMap>('/vacuum/map');
    return response.data;
};

export const getVacuumRooms = async (): Promise<VacuumRooms> => {
    const response = await apiBase.get<VacuumRooms>('/vacuum/rooms');
    return response.data;
};

export const renameVacuumRoom = async (
    roomId: number,
    rename: VacuumRoomRename
): Promise<VacuumRoom> => {
    const response = await apiBase.put<VacuumRoom>(`/vacuum/rooms/${roomId}`, rename);
    return response.data;
};

export const getVacuumPath = async (): Promise<VacuumPath> => {
    const response = await apiBase.get<VacuumPath>('/vacuum/path');
    return response.data;
};

export const startVacuumRoomCleaning = async (request: VacuumRoomCleaningRequest): Promise<void> => {
    await apiBase.post('/vacuum/room-cleaning', request);
};

export const startVacuumSpotCleaning = async (request: VacuumSpotCleaningRequest): Promise<void> => {
    await apiBase.post('/vacuum/spot-cleaning', request);
};
