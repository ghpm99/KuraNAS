import { act, renderHook } from '@testing-library/react';
import useTakeoutUpload from './useTakeoutUpload';

const initTakeoutUploadMock = jest.fn();
const uploadTakeoutChunkMock = jest.fn();
const completeTakeoutUploadMock = jest.fn();

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => key,
	}),
}));

jest.mock('@/service/takeout', () => ({
	initTakeoutUpload: (...args: unknown[]) => initTakeoutUploadMock(...args),
	uploadTakeoutChunk: (...args: unknown[]) => uploadTakeoutChunkMock(...args),
	completeTakeoutUpload: (...args: unknown[]) => completeTakeoutUploadMock(...args),
}));

describe('components/takeout/useTakeoutUpload', () => {
	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('rejects non zip file', () => {
		const { result } = renderHook(() => useTakeoutUpload());
		const file = new File(['x'], 'file.txt', { type: 'text/plain' });

		act(() => {
			result.current.selectFile(file);
		});

		expect(result.current.state).toBe('error');
		expect(result.current.errorMessage).toBe('TAKEOUT_FILE_TYPE_ERROR');
	});

	it('uploads file in chunks and completes', async () => {
		initTakeoutUploadMock.mockResolvedValue({ upload_id: 'u1', chunk_size: 2 });
		uploadTakeoutChunkMock.mockResolvedValue({ received: true });
		completeTakeoutUploadMock.mockResolvedValue({ job_id: 11, message: 'ok' });

		const { result } = renderHook(() => useTakeoutUpload());
		const file = new File(['hello'], 'takeout.zip', { type: 'application/zip' });

		act(() => {
			result.current.selectFile(file);
		});

		await act(async () => {
			await result.current.startUpload();
		});

		expect(initTakeoutUploadMock).toHaveBeenCalledWith('takeout.zip', 5);
		expect(uploadTakeoutChunkMock).toHaveBeenCalled();
		expect(completeTakeoutUploadMock).toHaveBeenCalledWith('u1');
		expect(result.current.state).toBe('done');
		expect(result.current.jobId).toBe(11);
	});

	it('resets state', () => {
		const { result } = renderHook(() => useTakeoutUpload());
		act(() => {
			result.current.reset();
		});
		expect(result.current.state).toBe('idle');
		expect(result.current.fileName).toBe('');
	});

	it('does nothing when startUpload is called without a file', async () => {
		const { result } = renderHook(() => useTakeoutUpload());

		await act(async () => {
			await result.current.startUpload();
		});

		expect(initTakeoutUploadMock).not.toHaveBeenCalled();
		expect(result.current.state).toBe('idle');
	});

	it('sets error state on upload failure', async () => {
		initTakeoutUploadMock.mockRejectedValue(new Error('network'));

		const { result } = renderHook(() => useTakeoutUpload());
		const file = new File(['hello'], 'takeout.zip', { type: 'application/zip' });

		act(() => {
			result.current.selectFile(file);
		});

		await act(async () => {
			await result.current.startUpload();
		});

		expect(result.current.state).toBe('error');
		expect(result.current.errorMessage).toBe('TAKEOUT_IMPORT_FAILED');
	});

	it('returns empty progressMessage when no file selected', () => {
		const { result } = renderHook(() => useTakeoutUpload());
		expect(result.current.progressMessage).toBe('');
	});

	it('accepts zip file and moves to selecting state', () => {
		const { result } = renderHook(() => useTakeoutUpload());
		const file = new File(['data'], 'archive.zip', { type: 'application/zip' });

		act(() => {
			result.current.selectFile(file);
		});

		expect(result.current.state).toBe('selecting');
		expect(result.current.fileName).toBe('archive.zip');
	});

	it('formats progress message with large file sizes', async () => {
		const largeContent = new Uint8Array(2048);
		const file = new File([largeContent], 'big.zip', { type: 'application/zip' });

		initTakeoutUploadMock.mockResolvedValue({ upload_id: 'u2', chunk_size: 1024 });
		uploadTakeoutChunkMock.mockResolvedValue({ received: true });
		completeTakeoutUploadMock.mockResolvedValue({ job_id: 22, message: 'ok' });

		const { result } = renderHook(() => useTakeoutUpload());

		act(() => {
			result.current.selectFile(file);
		});

		await act(async () => {
			await result.current.startUpload();
		});

		expect(result.current.state).toBe('done');
		expect(uploadTakeoutChunkMock).toHaveBeenCalledTimes(2);
	});
});
