import { createTestFile } from '@/features/files/selection/testFileFactory';
import { extractBackendErrorMessage, runBulkOperation } from './bulkOutcome';

describe('bulkOutcome', () => {
    it('extracts the backend error message only when present as text', () => {
        expect(extractBackendErrorMessage({ response: { data: { error: 'denied' } } })).toBe('denied');
        expect(extractBackendErrorMessage({ response: { data: { error: '' } } })).toBeUndefined();
        expect(extractBackendErrorMessage(new Error('x'))).toBeUndefined();
        expect(extractBackendErrorMessage(null)).toBeUndefined();
    });

    it('splits succeeded and failed files and keeps the first failure message', async () => {
        const files = [1, 2, 3, 4].map((id) => createTestFile(id));
        const outcome = await runBulkOperation(files, async (file) => {
            if (file.id === 2) throw { response: { data: { error: 'first failure' } } };
            if (file.id === 3) throw { response: { data: { error: 'second failure' } } };
        });

        expect(outcome.succeededFiles.map((file) => file.id)).toEqual([1, 4]);
        expect(outcome.failedFiles.map((file) => file.id)).toEqual([2, 3]);
        expect(outcome.firstFailureMessage).toBe('first failure');
    });
});
