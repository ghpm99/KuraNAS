import type { RgbColor } from './vacuumRoomColor';

type VacuumMapPalette = {
    backdrop: string;
    wall: RgbColor;
    obstacle: RgbColor;
    forbiddenFill: string;
    forbiddenStroke: string;
    dock: string;
    path: string;
    robot: string;
    robotRing: string;
    spotFill: string;
    spotStroke: string;
};

export const vacuumMapPalette: VacuumMapPalette = {
    backdrop: '#eef2f7',
    wall: [31, 41, 55],
    obstacle: [180, 83, 9],
    forbiddenFill: 'rgba(220, 38, 38, 0.3)',
    forbiddenStroke: '#b91c1c',
    dock: '#15803d',
    path: '#6d28d9',
    robot: '#1d4ed8',
    robotRing: '#ffffff',
    spotFill: 'rgba(14, 165, 233, 0.3)',
    spotStroke: '#0369a1',
};
