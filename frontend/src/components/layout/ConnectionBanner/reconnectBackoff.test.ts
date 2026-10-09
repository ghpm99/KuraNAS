import { getReconnectDelay } from './reconnectBackoff';

describe('layout/ConnectionBanner/reconnectBackoff', () => {
    it('grows exponentially up to the ceiling', () => {
        expect(getReconnectDelay(0)).toBe(2000);
        expect(getReconnectDelay(1)).toBe(4000);
        expect(getReconnectDelay(8)).toBe(30000);
    });
});
