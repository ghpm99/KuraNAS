import { isJobFinished } from './jobs';

describe('types/jobs isJobFinished', () => {
	it('treats terminal statuses as finished', () => {
		expect(isJobFinished('completed')).toBe(true);
		expect(isJobFinished('failed')).toBe(true);
		expect(isJobFinished('canceled')).toBe(true);
	});

	it('treats in-flight statuses as not finished', () => {
		expect(isJobFinished('queued')).toBe(false);
		expect(isJobFinished('running')).toBe(false);
	});
});
