import { decodeVacuumMapCells } from './decodeVacuumMapCells';

describe('decodeVacuumMapCells', () => {
    it('returns an empty grid when the cells are absent', () => {
        expect(decodeVacuumMapCells(undefined)).toHaveLength(0);
        expect(decodeVacuumMapCells('')).toHaveLength(0);
    });

    it('returns an empty grid when the text is not base64', () => {
        expect(decodeVacuumMapCells('***not base64***')).toHaveLength(0);
    });

    it('decodes every byte, including walls, rooms, unknown and obstacles', () => {
        const encodedCells = btoa(String.fromCharCode(0, 1, 126, 127, 255));

        expect(Array.from(decodeVacuumMapCells(encodedCells))).toEqual([0, 1, 126, 127, 255]);
    });
});
