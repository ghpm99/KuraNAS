import { apiBase } from '@/service';
import type { LibraryCategory, LibraryDto, UpdateLibraryRequest } from '@/types/libraries';

export const getLibraries = async (): Promise<LibraryDto[]> => {
	const response = await apiBase.get<LibraryDto[]>('/libraries');
	return response.data;
};

export const updateLibrary = async (
	category: LibraryCategory,
	request: UpdateLibraryRequest
): Promise<LibraryDto> => {
	const response = await apiBase.put<LibraryDto>(`/libraries/${encodeURIComponent(category)}`, request);
	return response.data;
};
