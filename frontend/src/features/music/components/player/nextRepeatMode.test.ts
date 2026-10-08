import { nextRepeatMode } from './nextRepeatMode';

describe('nextRepeatMode', () => {
    it.each([
        ['none', 'all'],
        ['all', 'one'],
        ['one', 'none'],
        ['unknown', 'none'],
    ])('goes from %s to %s', (currentMode, expectedMode) => {
        expect(nextRepeatMode(currentMode)).toBe(expectedMode);
    });
});
