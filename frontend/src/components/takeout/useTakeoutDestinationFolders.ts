import { getFileByDiskPath } from '@/service/files';
import { getLibraries } from '@/service/libraries';
import type { LibraryCategory } from '@/types/libraries';
import { useQuery } from '@tanstack/react-query';

const TAKEOUT_FOLDER_NAME = 'takeout';

export type TakeoutDestinationFolders = {
	imagesFolder: string;
	videosFolder: string;
};

const resolveTakeoutFolder = async (libraryPath: string | undefined): Promise<string> => {
	if (!libraryPath) return '';
	const libraryFolder = await getFileByDiskPath(libraryPath);
	return `${libraryFolder.path.replace(/\/$/, '')}/${TAKEOUT_FOLDER_NAME}`;
};

const findLibraryPath = (
	libraries: { category: LibraryCategory; path: string }[],
	category: LibraryCategory
): string | undefined => libraries.find((library) => library?.category === category)?.path;

const resolveTakeoutDestinationFolders = async (): Promise<TakeoutDestinationFolders> => {
	const libraries = await getLibraries();
	const knownLibraries = Array.isArray(libraries) ? libraries : [];
	const [imagesFolder, videosFolder] = await Promise.all([
		resolveTakeoutFolder(findLibraryPath(knownLibraries, 'images')),
		resolveTakeoutFolder(findLibraryPath(knownLibraries, 'videos')),
	]);
	return { imagesFolder, videosFolder };
};

const useTakeoutDestinationFolders = (): TakeoutDestinationFolders => {
	const { data } = useQuery({
		queryKey: ['takeout', 'destination-folders'],
		queryFn: resolveTakeoutDestinationFolders,
		retry: false,
	});
	return data ?? { imagesFolder: '', videosFolder: '' };
};

export default useTakeoutDestinationFolders;
