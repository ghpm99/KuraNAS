import { apiBase } from '.';
import { getServerHealth } from './health';

jest.mock('.', () => ({ apiBase: { get: jest.fn() } }));

describe('service/health', () => {
    it('requests the health endpoint and returns its payload', async () => {
        (apiBase.get as jest.Mock).mockResolvedValue({ data: { status: 'ok', service: 'kuranas' } });

        await expect(getServerHealth()).resolves.toEqual({ status: 'ok', service: 'kuranas' });
        expect(apiBase.get).toHaveBeenCalledWith('/health', { timeout: 5000 });
    });
});
