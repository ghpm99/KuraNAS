export type LibraryCategory = 'images' | 'music' | 'videos' | 'documents';

export type LibraryDto = {
	category: LibraryCategory;
	path: string;
};

export type UpdateLibraryRequest = {
	path: string;
};
