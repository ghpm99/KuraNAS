import { Pagination } from '@/types/pagination';
import { MusicAlbum, MusicArtist, MusicFolder, MusicGenre } from '@/types/music';
import { IMusicData } from '@/components/providers/musicProvider/musicProvider';
import { apiBase } from '.';

export const getMusicArtists = async (page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<MusicArtist>>('/files/music/artists', {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicByArtist = async (artist: string, page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<IMusicData>>(`/files/music/artists/${encodeURIComponent(artist)}`, {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicAlbums = async (page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<MusicAlbum>>('/files/music/albums', {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicByAlbum = async (album: string, page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<IMusicData>>(`/files/music/albums/${encodeURIComponent(album)}`, {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicGenres = async (page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<MusicGenre>>('/files/music/genres', {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicByGenre = async (genre: string, page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<IMusicData>>(`/files/music/genres/${encodeURIComponent(genre)}`, {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicFolders = async (page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<MusicFolder>>('/files/music/folders', {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusic = async (page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<IMusicData>>('/files/music', {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicByFolder = async (folder: string, page: number, pageSize: number) => {
	const data = await getMusic(page, pageSize);
	const items = data.items.filter((item) => item.path.startsWith(folder));

	return {
		...data,
		items,
		pagination: {
			...data.pagination,
			total_items: items.length,
		},
	};
};
