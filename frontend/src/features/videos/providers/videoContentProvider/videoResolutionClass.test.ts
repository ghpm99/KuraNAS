import { classifyVideoResolution } from './videoResolutionClass';

describe('classifyVideoResolution', () => {
    it.each([
        [1, 'sd'],
        [719, 'sd'],
        [720, 'hd'],
        [1079, 'hd'],
        [1080, 'fhd'],
        [2159, 'fhd'],
        [2160, 'uhd'],
        [4320, 'uhd'],
    ])('classifies height %d as %s', (height, expectedClass) => {
        expect(classifyVideoResolution(height)).toBe(expectedClass);
    });

    it.each([[undefined], [null], [0], [-1], [NaN]])('returns null for %p', (height) => {
        expect(classifyVideoResolution(height as number | null | undefined)).toBeNull();
    });
});
