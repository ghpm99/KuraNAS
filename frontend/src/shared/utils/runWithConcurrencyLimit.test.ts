import { runWithConcurrencyLimit } from './runWithConcurrencyLimit';

describe('runWithConcurrencyLimit', () => {
    it('returns an empty list for empty input', async () => {
        await expect(runWithConcurrencyLimit([], 4, async () => 1)).resolves.toEqual([]);
    });

    it('keeps input order and captures failures without aborting the rest', async () => {
        const settledTasks = await runWithConcurrencyLimit([1, 2, 3], 2, async (input) => {
            if (input === 2) throw new Error('boom');
            return input * 10;
        });

        expect(settledTasks.map((settledTask) => settledTask.status)).toEqual([
            'fulfilled',
            'rejected',
            'fulfilled',
        ]);
        expect(settledTasks[0]).toMatchObject({ output: 10 });
        expect(settledTasks[2]).toMatchObject({ output: 30 });
    });

    it('never runs more tasks at once than the limit', async () => {
        let runningTasks = 0;
        let highestRunningTasks = 0;
        await runWithConcurrencyLimit([1, 2, 3, 4, 5, 6, 7, 8, 9], 3, async () => {
            runningTasks += 1;
            highestRunningTasks = Math.max(highestRunningTasks, runningTasks);
            await new Promise((resolve) => setTimeout(resolve, 5));
            runningTasks -= 1;
        });

        expect(highestRunningTasks).toBe(3);
    });

    it('treats a limit below one as sequential execution', async () => {
        let runningTasks = 0;
        let highestRunningTasks = 0;
        await runWithConcurrencyLimit([1, 2, 3], 0, async () => {
            runningTasks += 1;
            highestRunningTasks = Math.max(highestRunningTasks, runningTasks);
            await new Promise((resolve) => setTimeout(resolve, 1));
            runningTasks -= 1;
        });

        expect(highestRunningTasks).toBe(1);
    });
});
