import { AxiosError } from 'axios';
import { shouldRetryQuery } from './queryRetryPolicy';

const buildResponseError = (status: number) => {
    const error = new AxiosError('failure');
    error.response = { status } as AxiosError['response'];
    return error;
};

describe('components/providers/queryRetryPolicy', () => {
    it('retries network and server errors once', () => {
        expect(shouldRetryQuery(0, new AxiosError('network'))).toBe(true);
        expect(shouldRetryQuery(1, new AxiosError('network'))).toBe(false);
        expect(shouldRetryQuery(0, buildResponseError(503))).toBe(true);
    });

    it('never retries client errors', () => {
        expect(shouldRetryQuery(0, buildResponseError(404))).toBe(false);
    });
});
