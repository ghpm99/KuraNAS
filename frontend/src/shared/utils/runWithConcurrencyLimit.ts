export type SettledTask<TInput, TOutput> =
    | { input: TInput; status: 'fulfilled'; output: TOutput }
    | { input: TInput; status: 'rejected'; reason: unknown };

export const runWithConcurrencyLimit = async <TInput, TOutput>(
    inputs: readonly TInput[],
    concurrencyLimit: number,
    task: (input: TInput) => Promise<TOutput>
): Promise<SettledTask<TInput, TOutput>[]> => {
    const settledTasks: SettledTask<TInput, TOutput>[] = new Array(inputs.length);
    let nextIndex = 0;

    const consumeQueue = async () => {
        while (nextIndex < inputs.length) {
            const currentIndex = nextIndex;
            nextIndex += 1;
            const input = inputs[currentIndex] as TInput;
            try {
                settledTasks[currentIndex] = {
                    input,
                    status: 'fulfilled',
                    output: await task(input),
                };
            } catch (reason) {
                settledTasks[currentIndex] = { input, status: 'rejected', reason };
            }
        }
    };

    const workerCount = Math.max(1, Math.min(Math.floor(concurrencyLimit), inputs.length));
    await Promise.all(Array.from({ length: workerCount }, consumeQueue));
    return settledTasks;
};
