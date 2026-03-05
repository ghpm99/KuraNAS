import {
	formatDate,
	formatDateTime,
	formatDuration,
	formatSize,
	getFileTypeInfo,
} from './index';

describe('utils/index', () => {
	it('formats sizes for bytes and larger units', () => {
		expect(formatSize(512)).toBe('512 B');
		expect(formatSize(1024)).toBe('1.00 KB');
		expect(formatSize(1024 * 1024)).toBe('1.00 MB');
	});

	it('formats duration across all branches', () => {
		expect(formatDuration(undefined)).toBe('Em andamento');
		expect(formatDuration(0)).toBe('Em andamento');
		expect(formatDuration(59)).toBe('59s');
		expect(formatDuration(120)).toBe('2m 0s');
		expect(formatDuration(3661)).toBe('1h 1m 1s');
	});

	it('formats date and handles Date constructor failure', () => {
		const realDate = Date;
		const consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => {});

		expect(typeof formatDate('2025-01-01T00:00:00Z')).toBe('string');

		(global as any).Date = class extends realDate {
			constructor(value?: string | number | Date) {
				if (value === 'boom') {
					throw new Error('invalid date');
				}
				super(value as any);
			}
		} as DateConstructor;

		expect(formatDate('boom')).toBe('boom');
		expect(consoleErrorSpy).toHaveBeenCalled();

		(global as any).Date = realDate;
		consoleErrorSpy.mockRestore();
	});

	it('formats datetime in pt-BR locale', () => {
		const value = formatDateTime(new Date('2025-01-02T03:04:05Z'));
		expect(value).toContain('/');
		expect(value).toContain(':');
	});

	it('returns file type metadata and unknown fallback', () => {
		expect(getFileTypeInfo('.MP3')).toEqual({
			type: 'audio',
			mime: 'audio/mpeg',
			description: 'AUDIO_MP3',
		});
		expect(getFileTypeInfo('.unknown')).toEqual({
			type: 'unknown',
			mime: '',
			description: 'UNKNOWN_FORMAT',
		});
	});
});
