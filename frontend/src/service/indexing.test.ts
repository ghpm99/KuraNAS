import { requestReindex } from './indexing';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: { post: jest.fn() },
}));

const mockedApi = apiBase as unknown as { post: jest.Mock };

describe('service/indexing', () => {
    beforeEach(() => jest.clearAllMocks());

    it('returns the job id of an accepted folder reindex', async () => {
        mockedApi.post.mockResolvedValue({ data: { job_id: 7 } });

        const result = await requestReindex({ scope: 'folder', folderId: 12 });

        expect(mockedApi.post).toHaveBeenCalledWith('/indexing/reindex', {
            scope: 'folder',
            folder_id: 12,
        });
        expect(result).toEqual({ jobId: 7, isAlreadyRunning: false });
    });

    it('sends root and all scopes with their own payload', async () => {
        mockedApi.post.mockResolvedValue({ data: { job_id: 1 } });

        await requestReindex({ scope: 'root', rootId: 3 });
        await requestReindex({ scope: 'all' });

        expect(mockedApi.post).toHaveBeenNthCalledWith(1, '/indexing/reindex', {
            scope: 'root',
            root_id: 3,
        });
        expect(mockedApi.post).toHaveBeenNthCalledWith(2, '/indexing/reindex', { scope: 'all' });
    });

    it('treats 409 with job_id as an already running job', async () => {
        mockedApi.post.mockRejectedValue({
            response: { status: 409, data: { error: 'em andamento', job_id: 9 } },
        });

        const result = await requestReindex({ scope: 'all' });

        expect(result).toEqual({ jobId: 9, isAlreadyRunning: true });
    });

    it('rethrows other failures', async () => {
        const failure = { response: { status: 404, data: { error: 'não existe' } } };
        mockedApi.post.mockRejectedValue(failure);

        await expect(requestReindex({ scope: 'folder', folderId: 1 })).rejects.toBe(failure);
    });

    it('rethrows 409 without job_id', async () => {
        const failure = { response: { status: 409, data: { error: 'conflito' } } };
        mockedApi.post.mockRejectedValue(failure);

        await expect(requestReindex({ scope: 'all' })).rejects.toBe(failure);
    });
});
