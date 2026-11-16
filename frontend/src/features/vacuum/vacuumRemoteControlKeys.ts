import type { VacuumDirection } from '@/types/vacuum';

export type VacuumMovementDirection = Exclude<VacuumDirection, 'stop'>;

const directionsByArrowKey: Record<string, VacuumMovementDirection> = {
    ArrowUp: 'forward',
    ArrowDown: 'backward',
    ArrowLeft: 'left',
    ArrowRight: 'right',
};

export const directionForArrowKey = (key: string): VacuumMovementDirection | undefined =>
    directionsByArrowKey[key];
