import { parseRepeatMode } from './repeatMode';

describe('parseRepeatMode', () => {
    it.each([
        ['all', 'all'],
        ['one', 'one'],
        ['none', 'none'],
        ['bogus', 'none'],
        [undefined, 'none'],
    ] as const)('maps %s to %s', (rawRepeatMode, expectedRepeatMode) => {
        expect(parseRepeatMode(rawRepeatMode)).toBe(expectedRepeatMode);
    });
});
