import { buildVacuumMapPixels } from './buildVacuumMapPixels';
import { buildVacuumMapPayload } from './testSupport/buildVacuumMapPayload';
import { vacuumMapPalette } from './vacuumMapPalette';
import { toVacuumMapModel } from './vacuumMapModel';
import { vacuumRoomColor } from './vacuumRoomColor';

const map = toVacuumMapModel(buildVacuumMapPayload());
if (!map) throw new Error('fixture must build a map');

const pixelAt = (pixels: Uint8ClampedArray, column: number, pixelRow: number): number[] => {
    const offset = (pixelRow * map.width + column) * 4;
    return Array.from(pixels.slice(offset, offset + 4));
};

describe('buildVacuumMapPixels', () => {
    it('has one RGBA pixel per cell', () => {
        expect(buildVacuumMapPixels(map, new Set())).toHaveLength(map.width * map.height * 4);
    });

    it('flips the rows so the lowest y is the bottom row', () => {
        const pixels = buildVacuumMapPixels(map, new Set());

        expect(pixelAt(pixels, 1, 1)).toEqual([...vacuumRoomColor(1, false), 255]);
        expect(pixelAt(pixels, 2, 1)).toEqual([...vacuumRoomColor(2, false), 255]);
        expect(pixelAt(pixels, 1, 0)).toEqual([...vacuumRoomColor(1, false), 255]);
    });

    it('paints walls and obstacles and leaves unknown transparent', () => {
        const pixels = buildVacuumMapPixels(map, new Set());

        expect(pixelAt(pixels, 0, 0)).toEqual([...vacuumMapPalette.wall, 255]);
        expect(pixelAt(pixels, 3, 1)).toEqual([...vacuumMapPalette.obstacle, 255]);
        expect(pixelAt(pixels, 3, 2)).toEqual([0, 0, 0, 0]);
        expect(pixelAt(pixels, 3, 0)).toEqual([0, 0, 0, 0]);
    });

    it('highlights the selected room only', () => {
        const pixels = buildVacuumMapPixels(map, new Set([2]));

        expect(pixelAt(pixels, 2, 1)).toEqual([...vacuumRoomColor(2, true), 255]);
        expect(pixelAt(pixels, 1, 1)).toEqual([...vacuumRoomColor(1, false), 255]);
    });
});
