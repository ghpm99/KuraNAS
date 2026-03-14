import { Pagination } from '@/types/pagination';
import { MusicAlbum, MusicArtist, MusicFolder, MusicGenre, MusicHomeCatalog } from '@/types/music';
import { IMusicData } from '@/components/providers/musicProvider/musicProvider';
import { apiBase } from '.';

export const getMusicHomeCatalog = async (limit: number) => {
	const response = await apiBase.get<MusicHomeCatalog>('/music/library/home', {
		params: { limit },
	});
	return response.data;
};

export const getMusicArtists = async (page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<MusicArtist>>('/music/library/artists', {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicByArtist = async (artistKey: string, page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<IMusicData>>(`/music/library/artists/${encodeURIComponent(artistKey)}/tracks`, {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicAlbums = async (page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<MusicAlbum>>('/music/library/albums', {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicByAlbum = async (albumKey: string, page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<IMusicData>>(`/music/library/albums/${encodeURIComponent(albumKey)}/tracks`, {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicGenres = async (page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<MusicGenre>>('/music/library/genres', {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicByGenre = async (genreKey: string, page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<IMusicData>>(`/music/library/genres/${encodeURIComponent(genreKey)}/tracks`, {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusicFolders = async (page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<MusicFolder>>('/music/library/folders', {
		params: { page, page_size: pageSize },
	});
	return response.data;
};

export const getMusic = async (page: number, pageSize: number) => {
	const response = await apiBase.get<Pagination<IMusicData>>('/music/library', {
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
