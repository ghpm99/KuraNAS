import {
	addChildrenToTree,
	buildFilesUrl,
	extractFilePath,
	findItemInTree,
	findTrailByIdInTree,
} from './fileProviderUtils';
import type { FileData } from './fileContext';

describe('fileProvider utilities', () => {
	const createNode = (id: number, children?: FileData[]): FileData => ({
		id,
		name: `node-${id}`,
		path: `/node-${id}`,
		parent_path: '/',
		type: 1,
		format: 'folder',
		size: 0,
		updated_at: '',
		created_at: '',
		deleted_at: '',
		last_interaction: '',
		last_backup: '',
		check_sum: '',
		directory_content_count: 0,
		starred: false,
		file_children: children,
	});

	it('extracts paths from the file route', () => {
		expect(extractFilePath('/files')).toBe('');
		expect(extractFilePath('/files/')).toBe('');
		expect(extractFilePath('/files/docs/new')).toBe('/docs/new');
		expect(extractFilePath('/other')).toBe('');
	});

	it('builds encoded file URLs', () => {
		expect(buildFilesUrl('docs/new')).toBe('/files/docs/new');
		expect(buildFilesUrl('folder name/file')).toBe('/files/folder%20name/file');
		expect(buildFilesUrl('')).toBe('/files');
	});

	it('finds items in nested trees', () => {
		const tree: FileData[] = [
			createNode(1, [createNode(2), createNode(3, [createNode(4)])]),
			createNode(5),
		];

		expect(findItemInTree(tree, 3)?.id).toBe(3);
		expect(findItemInTree(tree, 4)?.id).toBe(4);
		expect(findItemInTree(tree, 99)).toBeNull();
	});

	it('adds children to the correct parent node', () => {
		const tree: FileData[] = [
			createNode(1),
			createNode(2),
		];
		const updated = addChildrenToTree(tree, 2, [createNode(6)]);
		expect(updated[1].file_children?.[0].id).toBe(6);
	});

	it('builds the trail of nodes to a target id', () => {
		const nested = [
			createNode(1, [createNode(2, [createNode(3)])]),
			createNode(4),
		];
		const trail = findTrailByIdInTree(nested, 3);
		expect(trail?.map((node) => node.id)).toEqual([1, 2, 3]);
		expect(findTrailByIdInTree(nested, 99)).toBeNull();
	});
});
