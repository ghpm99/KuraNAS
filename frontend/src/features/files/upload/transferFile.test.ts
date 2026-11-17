import { uploadSingleFile } from '@/service/files';
import type { FileTransferParams } from './fileTransferTypes';
import { chunkedUploadThresholdBytes, transferFile } from './transferFile';
import { uploadFileInChunks } from './uploadFileInChunks';

jest.mock('@/service/files', () => ({ uploadSingleFile: jest.fn() }));
jest.mock('./uploadFileInChunks', () => ({ uploadFileInChunks: jest.fn() }));

const fileOfSize = (size: number): File => {
    const file = new File(['x'], 'big.bin');
    Object.defineProperty(file, 'size', { value: size });
    return file;
};

const buildParams = (file: File): FileTransferParams => ({
    file,
    targetFolderId: 7,
    relativePath: 'Pasta/big.bin',
    onConflict: 'replace',
    signal: new AbortController().signal,
    onPhaseChange: jest.fn(),
    onChecksumComputed: jest.fn(),
    onSentBytes: jest.fn(),
});

describe('transferFile', () => {
    beforeEach(() => jest.clearAllMocks());

    it('keeps the threshold at 64 MiB', () => {
        expect(chunkedUploadThresholdBytes).toBe(64 * 1024 * 1024);
    });

    it('sends a file at the threshold in a single request reporting bytes', async () => {
        jest.mocked(uploadSingleFile).mockImplementation(async ({ onProgress }) => {
            onProgress?.(1234);
            return { status: 'renamed', name: 'big (2).bin' };
        });
        const params = buildParams(fileOfSize(chunkedUploadThresholdBytes));

        const outcome = await transferFile(params);

        expect(outcome).toEqual({ status: 'renamed', name: 'big (2).bin' });
        expect(uploadSingleFile).toHaveBeenCalledWith({
            file: params.file,
            targetFolderId: 7,
            relativePath: 'Pasta/big.bin',
            onConflict: 'replace',
            signal: params.signal,
            onProgress: params.onSentBytes,
        });
        expect(params.onPhaseChange).toHaveBeenCalledWith('uploading');
        expect(params.onSentBytes).toHaveBeenCalledWith(1234);
        expect(uploadFileInChunks).not.toHaveBeenCalled();
    });

    it('sends a file above the threshold through chunked upload with the same params', async () => {
        jest.mocked(uploadFileInChunks).mockResolvedValue({ status: 'alreadyExisted' });
        const params = buildParams(fileOfSize(chunkedUploadThresholdBytes + 1));

        await expect(transferFile(params)).resolves.toEqual({ status: 'alreadyExisted' });

        expect(uploadFileInChunks).toHaveBeenCalledWith(params);
        expect(uploadSingleFile).not.toHaveBeenCalled();
    });
});
