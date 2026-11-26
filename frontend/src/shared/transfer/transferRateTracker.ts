export type TransferRateEstimate = {
    bytesPerSecond?: number;
    remainingSeconds?: number;
};

export type TransferRateTracker = {
    record: (sentBytes: number, totalBytes: number) => TransferRateEstimate;
    reset: () => void;
};

type ByteSample = { recordedAt: number; sentBytes: number };

const sampleWindowMs = 5000;
const minimumSampleSpanMs = 1000;

export const createTransferRateTracker = (now: () => number = Date.now): TransferRateTracker => {
    let samples: ByteSample[] = [];

    const record = (sentBytes: number, totalBytes: number): TransferRateEstimate => {
        const recordedAt = now();
        const isRealigned =
            samples.length > 0 && sentBytes < samples[samples.length - 1]!.sentBytes;
        const recentSamples = isRealigned
            ? []
            : samples.filter((sample) => recordedAt - sample.recordedAt <= sampleWindowMs);
        samples = [...recentSamples, { recordedAt, sentBytes }];

        const oldestSample = samples[0]!;
        const spanMs = recordedAt - oldestSample.recordedAt;
        if (spanMs < minimumSampleSpanMs) return {};

        const bytesPerSecond = ((sentBytes - oldestSample.sentBytes) * 1000) / spanMs;
        if (bytesPerSecond <= 0) return {};

        return {
            bytesPerSecond,
            remainingSeconds: Math.max(0, (totalBytes - sentBytes) / bytesPerSecond),
        };
    };

    const reset = () => {
        samples = [];
    };

    return { record, reset };
};
