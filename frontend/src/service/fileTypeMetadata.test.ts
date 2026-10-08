jest.mock('./index', () => ({
    apiBase: { get: jest.fn() },
}));

import { apiBase } from './index';
import { getAudioSummary, getImageSummary, getVideoSummary } from './fileTypeMetadata';

const mockedApi = apiBase as unknown as { get: jest.Mock };

describe('service/fileTypeMetadata', () => {
    beforeEach(() => jest.clearAllMocks());

    it('gets the image summary sending only the file id in the url', async () => {
        mockedApi.get.mockResolvedValue({ data: { width: 10 } });

        await expect(getImageSummary(4)).resolves.toEqual({ width: 10 });
        expect(mockedApi.get).toHaveBeenCalledWith('/image/metadata/4');
    });

    it('gets the audio summary sending only the file id in the url', async () => {
        mockedApi.get.mockResolvedValue({ data: { title: 'T' } });

        await expect(getAudioSummary(5)).resolves.toEqual({ title: 'T' });
        expect(mockedApi.get).toHaveBeenCalledWith('/music/metadata/5');
    });

    it('gets the video summary sending only the file id in the url', async () => {
        mockedApi.get.mockResolvedValue({ data: { codec_name: 'h264' } });

        await expect(getVideoSummary(6)).resolves.toEqual({ codec_name: 'h264' });
        expect(mockedApi.get).toHaveBeenCalledWith('/video/metadata/6');
    });
});
