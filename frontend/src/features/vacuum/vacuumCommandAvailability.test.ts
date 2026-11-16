import type { VacuumCommand, VacuumState } from '@/types/vacuum';
import { isVacuumCommandAvailable, vacuumCommandsInDisplayOrder } from './vacuumCommandAvailability';

const concreteStates: VacuumState[] = [
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
];

const expectedAvailableCommands: Record<VacuumState, VacuumCommand[]> = {
    idle: ['start', 'dock', 'locate'],
    cleaning: ['pause', 'stop', 'dock', 'locate'],
    paused: ['resume', 'stop', 'dock', 'locate'],
    returning: ['stop', 'locate'],
    charging: ['start', 'locate'],
    charged: ['start', 'locate'],
    emptying: ['locate'],
    sleeping: ['start', 'dock', 'locate'],
    remote_control: ['dock', 'locate'],
    fault: ['dock', 'locate'],
    unknown: ['start', 'pause', 'resume', 'stop', 'dock', 'locate'],
};

describe('vacuumCommandAvailability', () => {
    it.each([...concreteStates, 'unknown' as VacuumState])(
        'offers the right commands in state %s',
        (state) => {
            const availableCommands = vacuumCommandsInDisplayOrder.filter((command) =>
                isVacuumCommandAvailable(command, state)
            );

            expect(availableCommands).toEqual(
                vacuumCommandsInDisplayOrder.filter((command) =>
                    expectedAvailableCommands[state].includes(command)
                )
            );
        }
    );

    it('always offers locate', () => {
        concreteStates.forEach((state) =>
            expect(isVacuumCommandAvailable('locate', state)).toBe(true)
        );
    });
});
