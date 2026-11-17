import { createSha256 } from './sha256';

const DEFAULT_SLICE_SIZE_BYTES = 4 * 1024 * 1024;

interface ComputeFileSha256Options {
    signal?: AbortSignal;
    sliceSize?: number;
    onProgress?: (hashedBytes: number) => void;
}

const readSliceBytes = (slice: Blob): Promise<Uint8Array> =>
    new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(new Uint8Array(reader.result as ArrayBuffer));
        reader.onerror = () => reject(reader.error);
        reader.readAsArrayBuffer(slice);
    });

const throwIfAborted = (signal?: AbortSignal): void => {
    if (!signal?.aborted) return;
    throw signal.reason ?? new DOMException('Aborted', 'AbortError');
};

/**
 * Computes the SHA-256 of a Blob streaming it in slices, so memory stays bounded by `sliceSize`.
 * Rejects with the signal reason (AbortError by default) when `signal` aborts between slices.
 * `onProgress` receives the cumulative number of hashed bytes after each slice.
 */
export const computeFileSha256 = async (
    file: Blob,
    { signal, sliceSize = DEFAULT_SLICE_SIZE_BYTES, onProgress }: ComputeFileSha256Options = {}
): Promise<string> => {
    const hasher = createSha256();
    throwIfAborted(signal);
    for (let start = 0; start < file.size; start += sliceSize) {
        const end = Math.min(start + sliceSize, file.size);
        hasher.update(await readSliceBytes(file.slice(start, end)));
        throwIfAborted(signal);
        onProgress?.(end);
    }
    return hasher.digestHex();
};
