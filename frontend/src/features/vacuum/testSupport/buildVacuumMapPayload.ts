import type { VacuumMap } from '@/types/vacuum';

export const buildVacuumMapPayload = (overrides: Partial<VacuumMap> = {}): VacuumMap => ({
    map_id: 77,
    width: 4,
    height: 3,
    resolution: 50,
    origin_x: -100,
    origin_y: -50,
    cells: btoa(String.fromCharCode(0, 0, 0, 127, 0, 1, 2, 255, 0, 1, 2, 127)),
    dock: { x: 25, y: 25 },
    forbidden_zones: [
        {
            vertices: [
                [0, 0],
                [50, 0],
                [50, 50],
            ],
        },
    ],
    read_at: '2026-11-25T10:00:00Z',
    ...overrides,
});
