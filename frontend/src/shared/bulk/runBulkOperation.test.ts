import { runBulkOperation } from './runBulkOperation';

describe('runBulkOperation', () => {
    it('splits succeeded and failed items and keeps the first backend failure message', async () => {
        const outcome = await runBulkOperation([1, 2, 3, 4], async (itemId) => {
            if (itemId === 2) throw { response: { data: { error: 'first failure' } } };
            if (itemId === 3) throw { response: { data: { error: 'second failure' } } };
        });

        expect(outcome.succeededItems).toEqual([1, 4]);
        expect(outcome.failedItems).toEqual([2, 3]);
        expect(outcome.firstFailureMessage).toBe('first failure');
    });

    it('returns empty lists for no items', async () => {
        const outcome = await runBulkOperation<number>([], async () => undefined);

        expect(outcome).toEqual({ succeededItems: [], failedItems: [], firstFailureMessage: undefined });
    });
});
