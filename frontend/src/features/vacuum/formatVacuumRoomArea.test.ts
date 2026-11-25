import { formatVacuumRoomArea } from './formatVacuumRoomArea';

describe('formatVacuumRoomArea', () => {
    it('uses one decimal place and the language decimal separator', () => {
        expect(formatVacuumRoomArea(3.84, ',')).toBe('3,8');
        expect(formatVacuumRoomArea(12, '.')).toBe('12.0');
    });

    it('falls back to a dot when the separator is not a single character', () => {
        expect(formatVacuumRoomArea(1.5, 'VACUUM_ROOM_AREA_DECIMAL_SEPARATOR')).toBe('1.5');
    });

    it('treats a missing or invalid area as zero', () => {
        expect(formatVacuumRoomArea(undefined, ',')).toBe('0,0');
        expect(formatVacuumRoomArea(Number.NaN, ',')).toBe('0,0');
        expect(formatVacuumRoomArea(-2, ',')).toBe('0,0');
    });
});
