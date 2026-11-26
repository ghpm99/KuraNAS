import { formatRemainingTime } from './formatRemainingTime';

const translate = (key: string, variables: Record<string, string> = {}) =>
    `${key}:${Object.entries(variables)
        .map(([name, variable]) => `${name}=${variable}`)
        .join(',')}`;

describe('formatRemainingTime', () => {
    it.each([
        [0, 'FILES_UPLOAD_TIME_SECONDS:seconds=0'],
        [44.2, 'FILES_UPLOAD_TIME_SECONDS:seconds=45'],
        [59, 'FILES_UPLOAD_TIME_SECONDS:seconds=59'],
        [60, 'FILES_UPLOAD_TIME_MINUTES:minutes=1'],
        [125, 'FILES_UPLOAD_TIME_MINUTES:minutes=2'],
        [3600, 'FILES_UPLOAD_TIME_HOURS_MINUTES:hours=1,minutes=0'],
        [3900, 'FILES_UPLOAD_TIME_HOURS_MINUTES:hours=1,minutes=5'],
        [3590, 'FILES_UPLOAD_TIME_HOURS_MINUTES:hours=1,minutes=0'],
        [7170, 'FILES_UPLOAD_TIME_HOURS_MINUTES:hours=2,minutes=0'],
        [-5, 'FILES_UPLOAD_TIME_SECONDS:seconds=0'],
    ])('formats %s seconds', (remainingSeconds, expected) => {
        expect(formatRemainingTime(remainingSeconds, translate)).toBe(expected);
    });
});
