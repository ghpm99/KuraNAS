import { directionForArrowKey } from './vacuumRemoteControlKeys';

describe('directionForArrowKey', () => {
    it.each([
        ['ArrowUp', 'forward'],
        ['ArrowDown', 'backward'],
        ['ArrowLeft', 'left'],
        ['ArrowRight', 'right'],
    ])('maps %s to %s', (key, direction) => {
        expect(directionForArrowKey(key)).toBe(direction);
    });

    it('ignores keys that are not arrows', () => {
        expect(directionForArrowKey('Enter')).toBeUndefined();
        expect(directionForArrowKey('a')).toBeUndefined();
    });
});
