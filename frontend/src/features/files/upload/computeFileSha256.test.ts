import { createHash, randomBytes } from 'node:crypto';
import { computeFileSha256 } from './computeFileSha256';

const EMPTY_SHA256 = 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855';

const nodeDigest = (bytes: Uint8Array): string => createHash('sha256').update(bytes).digest('hex');

describe('computeFileSha256', () => {
    it('matches node:crypto reading in small slices', async () => {
        const bytes = randomBytes(5300);
        const digest = await computeFileSha256(new Blob([bytes]), { sliceSize: 1000 });
        expect(digest).toBe(nodeDigest(bytes));
    });

    it('matches node:crypto with the default slice size', async () => {
        const bytes = randomBytes(3000);
        expect(await computeFileSha256(new Blob([bytes]))).toBe(nodeDigest(bytes));
    });

    it('reports increasing progress ending at the total size', async () => {
        const progressValues: number[] = [];
        await computeFileSha256(new Blob([randomBytes(2500)]), {
            sliceSize: 1000,
            onProgress: (hashedBytes) => progressValues.push(hashedBytes),
        });
        expect(progressValues).toEqual([1000, 2000, 2500]);
    });

    it('returns the hash of the empty input for an empty Blob', async () => {
        const onProgress = jest.fn();
        expect(await computeFileSha256(new Blob([]), { onProgress })).toBe(EMPTY_SHA256);
        expect(onProgress).not.toHaveBeenCalled();
    });

    it('rejects with AbortError when the signal is already aborted', async () => {
        const controller = new AbortController();
        controller.abort();
        const onProgress = jest.fn();
        await expect(
            computeFileSha256(new Blob([randomBytes(100)]), {
                signal: controller.signal,
                onProgress,
            })
        ).rejects.toMatchObject({ name: 'AbortError' });
        expect(onProgress).not.toHaveBeenCalled();
    });

    it('rejects with AbortError when aborted midway and stops reporting', async () => {
        const controller = new AbortController();
        const progressValues: number[] = [];
        await expect(
            computeFileSha256(new Blob([randomBytes(5000)]), {
                signal: controller.signal,
                sliceSize: 1000,
                onProgress: (hashedBytes) => {
                    progressValues.push(hashedBytes);
                    if (hashedBytes === 2000) controller.abort();
                },
            })
        ).rejects.toMatchObject({ name: 'AbortError' });
        expect(progressValues).toEqual([1000, 2000]);
    });
});
