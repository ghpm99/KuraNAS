import { AxiosError } from 'axios';
import { isClientError, isNetworkError } from './networkError';

const buildAxiosError = (status?: number) => {
    const error = new AxiosError('failure');
    if (status !== undefined) {
        error.response = { status } as AxiosError['response'];
    }
    return error;
};

describe('service/networkError', () => {
    it('detects axios errors without a response as network errors', () => {
        expect(isNetworkError(buildAxiosError())).toBe(true);
        expect(isNetworkError(buildAxiosError(500))).toBe(false);
        expect(isNetworkError(new Error('plain'))).toBe(false);
    });

    it('detects 4xx responses as client errors', () => {
        expect(isClientError(buildAxiosError(404))).toBe(true);
        expect(isClientError(buildAxiosError(500))).toBe(false);
        expect(isClientError(buildAxiosError())).toBe(false);
        expect(isClientError(new Error('plain'))).toBe(false);
    });
});
