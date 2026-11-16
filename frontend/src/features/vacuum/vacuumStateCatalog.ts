import type { VacuumState } from '@/types/vacuum';

export const vacuumStateLabelKeys: Record<VacuumState, string> = {
    idle: 'VACUUM_STATE_IDLE',
    cleaning: 'VACUUM_STATE_CLEANING',
    paused: 'VACUUM_STATE_PAUSED',
    returning: 'VACUUM_STATE_RETURNING',
    charging: 'VACUUM_STATE_CHARGING',
    charged: 'VACUUM_STATE_CHARGED',
    emptying: 'VACUUM_STATE_EMPTYING',
    sleeping: 'VACUUM_STATE_SLEEPING',
    remote_control: 'VACUUM_STATE_REMOTE_CONTROL',
    fault: 'VACUUM_STATE_FAULT',
    unknown: 'VACUUM_STATE_UNKNOWN',
};

export const toVacuumState = (reportedState: string | undefined): VacuumState => {
    if (reportedState === undefined) return 'unknown';
    return reportedState in vacuumStateLabelKeys ? (reportedState as VacuumState) : 'unknown';
};

export const isCleaningSessionState = (state: VacuumState): boolean =>
    state === 'cleaning' || state === 'paused';
