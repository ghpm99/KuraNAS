jest.mock('./index', () => ({
    apiBase: {
        get: jest.fn(),
        post: jest.fn(),
        put: jest.fn(),
        delete: jest.fn(),
    },
}));

import { apiBase } from './index';
import {
    addImageAlbumItems,
    createImageAlbum,
    deleteImageAlbum,
    getImageAlbum,
    getImageAlbums,
    removeImageAlbumItems,
    updateImageAlbum,
} from './imageAlbum';

const mockedApi = apiBase as unknown as Record<'get' | 'post' | 'put' | 'delete', jest.Mock>;

describe('service/imageAlbum', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApi.get.mockResolvedValue({ data: { id: 1 } });
        mockedApi.post.mockResolvedValue({ data: { id: 1 } });
        mockedApi.put.mockResolvedValue({ data: { id: 1 } });
        mockedApi.delete.mockResolvedValue({ data: { requested: 2, changed: 2 } });
    });

    it('lists albums with the pagination params', async () => {
        await getImageAlbums(2, 10);
        expect(mockedApi.get).toHaveBeenCalledWith('/image/albums', {
            params: { page: 2, page_size: 10 },
        });
    });

    it('reads one album', async () => {
        await expect(getImageAlbum(3)).resolves.toEqual({ id: 1 });
        expect(mockedApi.get).toHaveBeenCalledWith('/image/albums/3');
    });

    it('creates an album posting only the name', async () => {
        await createImageAlbum('Trip');
        expect(mockedApi.post).toHaveBeenCalledWith('/image/albums', { name: 'Trip' });
    });

    it('renames an album and sets its cover with the update payload', async () => {
        await updateImageAlbum(3, { name: 'New' });
        expect(mockedApi.put).toHaveBeenCalledWith('/image/albums/3', { name: 'New' });
        await updateImageAlbum(3, { cover_file_id: 9 });
        expect(mockedApi.put).toHaveBeenCalledWith('/image/albums/3', { cover_file_id: 9 });
    });

    it('deletes an album by id', async () => {
        await deleteImageAlbum(3);
        expect(mockedApi.delete).toHaveBeenCalledWith('/image/albums/3');
    });

    it('adds and removes photos sending file_ids in the body', async () => {
        await addImageAlbumItems(3, [4, 5]);
        expect(mockedApi.post).toHaveBeenCalledWith('/image/albums/3/items', { file_ids: [4, 5] });

        await expect(removeImageAlbumItems(3, [4, 5])).resolves.toEqual({
            requested: 2,
            changed: 2,
        });
        expect(mockedApi.delete).toHaveBeenCalledWith('/image/albums/3/items', {
            data: { file_ids: [4, 5] },
        });
    });
});
