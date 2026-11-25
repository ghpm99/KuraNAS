import type { VacuumRoom } from '@/types/vacuum';

type Translate = (key: string, options?: Record<string, string>) => string;

export const resolveVacuumRoomName = (room: VacuumRoom, translate: Translate): string => {
    const givenName = typeof room.name === 'string' ? room.name.trim() : '';
    return givenName || translate('VACUUM_ROOM_DEFAULT_NAME', { number: String(room.id) });
};
