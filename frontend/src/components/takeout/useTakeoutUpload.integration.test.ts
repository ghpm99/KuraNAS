import { act, renderHook, waitFor } from '@testing-library/react';
import useTakeoutUpload from './useTakeoutUpload';
import { apiBase } from '@/service';

// Seam test: only apiBase is mocked; the real useTakeoutUpload + service/takeout.ts
// run, so the upload handshake asserts the exact init/chunk/complete endpoints and
// payloads the backend takeout handlers decode.
jest.mock('@/service', () => ({
	apiBase: { post: jest.fn() },
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const mockedApi = apiBase as unknown as { post: jest.Mock };

describe('components/takeout/useTakeoutUpload (seam)', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedApi.post.mockImplementation((url: string) => {
			if (url === '/takeout/upload/init') {
				return Promise.resolve({ data: { upload_id: 'u1', chunk_size: 10 * 1024 * 1024 } });
			}
			if (url === '/takeout/upload/chunk') return Promise.resolve({ data: { received: true } });
			if (url === '/takeout/upload/complete') {
				return Promise.resolve({ data: { job_id: 42, message: 'ok' } });
			}
			return Promise.reject(new Error(`unexpected POST ${url}`));
		});
	});

	it('startUpload runs init -> chunk -> complete against the takeout endpoints', async () => {
		const { result } = renderHook(() => useTakeoutUpload());

		const file = new File(['hello'], 'takeout.zip', { type: 'application/zip' });
		act(() => result.current.selectFile(file));

		await act(async () => {
			await result.current.startUpload();
		});

		await waitFor(() =>
			expect(mockedApi.post).toHaveBeenCalledWith('/takeout/upload/init', {
				file_name: 'takeout.zip',
				size: 5,
			})
		);
		expect(mockedApi.post).toHaveBeenCalledWith(
			'/takeout/upload/chunk',
			expect.any(FormData),
			expect.objectContaining({ headers: { 'Content-Type': 'multipart/form-data' } })
		);
		expect(mockedApi.post).toHaveBeenCalledWith('/takeout/upload/complete', { upload_id: 'u1' });
	});
});
