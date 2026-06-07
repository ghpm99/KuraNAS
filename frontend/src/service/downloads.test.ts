jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

jest.mock('./apiUrl', () => ({
    getApiBaseUrl: jest.fn(),
}));

import { apiBase } from './index';
import { getApiBaseUrl } from './apiUrl';
import { buildDownloadHref, getDownloads } from './downloads';
import type { DownloadItem } from '@/types/downloads';

const mockedApi = apiBase as unknown as { get: jest.Mock };
const mockedGetApiBaseUrl = getApiBaseUrl as jest.Mock;

const sampleItem: DownloadItem = {
    id: 'android',
    name: 'Android App',
    description: 'Native app',
    platform: 'android',
    version: '1.0.0',
    min_os: 'Android 13',
    size_bytes: 1024,
    sha256: 'abc',
    download_url: '/api/v1/downloads/android',
};

describe('service/downloads', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('fetches the download catalog', async () => {
        mockedApi.get.mockResolvedValue({ data: [sampleItem] });
        const result = await getDownloads();
        expect(mockedApi.get).toHaveBeenCalledWith('/downloads');
        expect(result).toEqual([sampleItem]);
    });

    it('builds a same-origin href when no base url is set', () => {
        mockedGetApiBaseUrl.mockReturnValue('');
        expect(buildDownloadHref(sampleItem)).toBe('/api/v1/downloads/android');
    });

    it('builds an absolute href against a remote server', () => {
        mockedGetApiBaseUrl.mockReturnValue('http://nas.local:8000');
        expect(buildDownloadHref(sampleItem)).toBe('http://nas.local:8000/api/v1/downloads/android');
    });
});
