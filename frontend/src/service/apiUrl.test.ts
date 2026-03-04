import { getApiBaseUrl, getApiV1BaseUrl } from './apiUrl';

describe('service/apiUrl', () => {
	const originalRuntime = (globalThis as any).__KURANAS_API_URL__;
	const originalEnv = process.env.VITE_API_URL;

	afterEach(() => {
		if (originalRuntime === undefined) {
			delete (globalThis as any).__KURANAS_API_URL__;
		} else {
			(globalThis as any).__KURANAS_API_URL__ = originalRuntime;
		}
		if (originalEnv === undefined) {
			delete process.env.VITE_API_URL;
		} else {
			process.env.VITE_API_URL = originalEnv;
		}
	});

	it('prefers runtime global and normalizes trailing slashes', () => {
		(globalThis as any).__KURANAS_API_URL__ = 'http://runtime.local///';
		process.env.VITE_API_URL = 'http://env.local';
		expect(getApiBaseUrl()).toBe('http://runtime.local');
		expect(getApiV1BaseUrl()).toBe('http://runtime.local/api/v1');
	});

	it('uses env when runtime global is blank', () => {
		(globalThis as any).__KURANAS_API_URL__ = '   ';
		process.env.VITE_API_URL = 'http://env.local/';
		expect(getApiBaseUrl()).toBe('http://env.local');
		expect(getApiV1BaseUrl()).toBe('http://env.local/api/v1');
	});

	it('returns local API path when no valid base url exists', () => {
		delete (globalThis as any).__KURANAS_API_URL__;
		process.env.VITE_API_URL = '';
		expect(getApiBaseUrl()).toBe('');
		expect(getApiV1BaseUrl()).toBe('/api/v1');
	});
});
