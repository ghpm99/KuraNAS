import { appRoutes } from '@/app/routes';
import { FileData } from './fileContext';

const FILES_PREFIX = appRoutes.files;

export const extractFilePath = (pathname: string): string => {
	if (!pathname.startsWith(FILES_PREFIX)) return '';
	const rest = pathname.slice(FILES_PREFIX.length);
	if (!rest || rest === '/') return '';
	return decodeURIComponent(rest);
};

export const buildFilesUrl = (filePath: string): string => {
	if (!filePath) return FILES_PREFIX;
	const encoded = filePath
		.split('/')
		.map((segment) => encodeURIComponent(segment))
		.join('/');
	return `${FILES_PREFIX}${encoded.startsWith('/') ? '' : '/'}${encoded}`;
};

export const findItemInTree = (data: FileData[], itemId: number | null): FileData | null => {
	if (!itemId) return null;
	for (const item of data) {
		if (item.id === itemId) {
			return item;
		}
		if (item?.file_children && item?.file_children?.length > 0) {
			const itemChildren = findItemInTree(item?.file_children, itemId);
			if (itemChildren) {
				return itemChildren;
			}
		}
	}

	return null;
};

export const addChildrenToTree = (tree: FileData[], parentId: number, children?: FileData[]): FileData[] => {
	return tree.map((node) => {
		if (node.id === parentId) {
			return { ...node, file_children: children };
		}
		if (node.file_children) {
			return { ...node, file_children: addChildrenToTree(node.file_children, parentId, children) };
		}
		return node;
	});
};

export const findTrailByIdInTree = (nodes: FileData[], targetId: number): FileData[] | null => {
	for (const node of nodes) {
		if (node.id === targetId) {
			return [node];
		}
		if (node.file_children?.length) {
			const branch = findTrailByIdInTree(node.file_children, targetId);
			if (branch) {
				return [node, ...branch];
			}
		}
	}
	return null;
};
