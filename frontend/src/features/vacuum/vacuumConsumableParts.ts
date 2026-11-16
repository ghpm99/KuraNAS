import type { VacuumConsumablePart } from '@/types/vacuum';

export const vacuumPartLabelKeys: Record<VacuumConsumablePart, string> = {
    filter: 'VACUUM_PART_FILTER',
    main_brush: 'VACUUM_PART_MAIN_BRUSH',
    side_brush: 'VACUUM_PART_SIDE_BRUSH',
};

export const isVacuumConsumablePart = (part: string): part is VacuumConsumablePart =>
    part in vacuumPartLabelKeys;
