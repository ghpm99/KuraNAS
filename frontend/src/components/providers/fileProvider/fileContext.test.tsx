import { renderHook } from '@testing-library/react';
import { FileContextProvider, useFile, type FileContextType } from './fileContext';

describe('providers/fileProvider/fileContext', () => {
	it('throws when used outside provider', () => {
		expect(() => renderHook(() => useFile())).toThrow('useFile must be used within a FileProvider');
	});

	it('returns context inside provider', () => {
		const value: FileContextType = {
			files: [],
			recentAccessFiles: [],
			isLoadingAccessData: false,
			status: 'success',
			selectedItem: null,
			handleSelectItem: jest.fn(),
			handleStarredItem: jest.fn(),
			uploadFiles: jest.fn().mockResolvedValue(undefined),
			createFolder: jest.fn().mockResolvedValue(undefined),
			movePath: jest.fn().mockResolvedValue(undefined),
			copyPath: jest.fn().mockResolvedValue(undefined),
			renamePath: jest.fn().mockResolvedValue(undefined),
			deletePath: jest.fn().mockResolvedValue(undefined),
			rescanFiles: jest.fn().mockResolvedValue(undefined),
			expandedItems: [],
			fileListFilter: 'all' as const,
			setFileListFilter: jest.fn(),
		};

		const wrapper = ({ children }: { children: React.ReactNode }) => (
			<FileContextProvider value={value}>{children}</FileContextProvider>
		);

		const { result } = renderHook(() => useFile(), { wrapper });
		expect(result.current.status).toBe('success');
		expect(result.current.fileListFilter).toBe('all');
	});
});
