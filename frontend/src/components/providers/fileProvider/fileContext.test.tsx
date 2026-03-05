import { renderHook } from '@testing-library/react';
import { FileContextProvider, useFile } from './fileContext';

describe('providers/fileProvider/fileContext', () => {
	it('throws when used outside provider', () => {
		expect(() => renderHook(() => useFile())).toThrow('useFile must be used within a FileProvider');
	});

	it('returns context inside provider', () => {
		const value = {
			files: [],
			recentAccessFiles: [],
			isLoadingAccessData: false,
			status: 'success',
			selectedItem: null,
			handleSelectItem: jest.fn(),
			handleStarredItem: jest.fn(),
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
