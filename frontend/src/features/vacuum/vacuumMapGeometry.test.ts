import { buildVacuumMapPayload } from './testSupport/buildVacuumMapPayload';
import {
    buildSpotCleaningSquare,
    cellIndexAt,
    findRoomIdAt,
    mapFractionToMillimeters,
    millimetersToCanvasPixels,
} from './vacuumMapGeometry';
import { toVacuumMapModel } from './vacuumMapModel';

const map = toVacuumMapModel(buildVacuumMapPayload());
if (!map) throw new Error('fixture must build a map');

describe('vacuumMapGeometry', () => {
    it('puts the lowest y at the bottom of the canvas', () => {
        expect(millimetersToCanvasPixels(map, { x: -100, y: -50 }, 10)).toEqual({ x: 0, y: 30 });
        expect(millimetersToCanvasPixels(map, { x: 100, y: 100 }, 10)).toEqual({ x: 40, y: 0 });
    });

    it('converts a position in the drawn map back to millimeters', () => {
        expect(mapFractionToMillimeters(map, 0, 1)).toEqual({ x: -100, y: -50 });
        expect(mapFractionToMillimeters(map, 1, 0)).toEqual({ x: 100, y: 100 });
    });

    it('finds the cell index with row 0 being the lowest y', () => {
        expect(cellIndexAt(map, { x: -100, y: -50 })).toBe(0);
        expect(cellIndexAt(map, { x: 99, y: 99 })).toBe(11);
        expect(cellIndexAt(map, { x: 100, y: 0 })).toBe(-1);
        expect(cellIndexAt(map, { x: -101, y: 0 })).toBe(-1);
        expect(cellIndexAt(map, { x: 0, y: -51 })).toBe(-1);
        expect(cellIndexAt(map, { x: 0, y: 100 })).toBe(-1);
    });

    it('finds the room only on floor cells', () => {
        expect(findRoomIdAt(map, { x: -25, y: 25 })).toBe(1);
        expect(findRoomIdAt(map, { x: 25, y: 25 })).toBe(2);
        expect(findRoomIdAt(map, { x: 75, y: 25 })).toBeNull();
        expect(findRoomIdAt(map, { x: 75, y: -25 })).toBeNull();
        expect(findRoomIdAt(map, { x: -75, y: 25 })).toBeNull();
        expect(findRoomIdAt(map, { x: 5000, y: 0 })).toBeNull();
    });

    it('builds a 1.5 m square centered on the point', () => {
        expect(buildSpotCleaningSquare({ x: 100, y: 200 })).toEqual([
            [-650, -550],
            [850, -550],
            [850, 950],
            [-650, 950],
        ]);
    });
});
