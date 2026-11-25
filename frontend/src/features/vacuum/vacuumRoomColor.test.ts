import { toCssColor, vacuumRoomColor } from './vacuumRoomColor';

describe('vacuumRoomColor', () => {
    it('is stable for the same room', () => {
        expect(vacuumRoomColor(5, false)).toEqual(vacuumRoomColor(5, false));
    });

    it('gives different rooms different colors', () => {
        const colors = [1, 2, 3, 4, 5, 6].map((roomId) => toCssColor(vacuumRoomColor(roomId, false)));

        expect(new Set(colors).size).toBe(colors.length);
    });

    it('makes the selected room more saturated than the idle one', () => {
        const spread = (channels: readonly number[]) =>
            Math.max(...channels) - Math.min(...channels);

        expect(spread(vacuumRoomColor(3, true))).toBeGreaterThan(spread(vacuumRoomColor(3, false)));
    });

    it('keeps every channel inside 0..255', () => {
        for (let roomId = 1; roomId <= 126; roomId += 1) {
            [...vacuumRoomColor(roomId, true), ...vacuumRoomColor(roomId, false)].forEach((channel) => {
                expect(channel).toBeGreaterThanOrEqual(0);
                expect(channel).toBeLessThanOrEqual(255);
            });
        }
    });

    it('formats a css color', () => {
        expect(toCssColor([1, 2, 3])).toBe('rgb(1, 2, 3)');
    });
});
