import type { VacuumCommand, VacuumState } from '@/types/vacuum';

export const vacuumCommandsInDisplayOrder: readonly VacuumCommand[] = [
    'start',
    'pause',
    'resume',
    'stop',
    'dock',
    'locate',
];

export const vacuumCommandLabelKeys: Record<VacuumCommand, string> = {
    start: 'VACUUM_COMMAND_START',
    pause: 'VACUUM_COMMAND_PAUSE',
    resume: 'VACUUM_COMMAND_RESUME',
    stop: 'VACUUM_COMMAND_STOP',
    dock: 'VACUUM_COMMAND_DOCK',
    locate: 'VACUUM_COMMAND_LOCATE',
};

const statesThatAllowCommand: Record<VacuumCommand, readonly VacuumState[]> = {
    start: ['idle', 'charging', 'charged', 'sleeping'],
    pause: ['cleaning'],
    resume: ['paused'],
    stop: ['cleaning', 'paused', 'returning'],
    dock: ['idle', 'cleaning', 'paused', 'sleeping', 'remote_control', 'fault'],
    locate: [
        'idle',
        'cleaning',
        'paused',
        'returning',
        'charging',
        'charged',
        'emptying',
        'sleeping',
        'remote_control',
        'fault',
    ],
};

export const isVacuumCommandAvailable = (command: VacuumCommand, state: VacuumState): boolean =>
    state === 'unknown' || statesThatAllowCommand[command].includes(state);
