import { act, renderHook } from '@testing-library/react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import useLibrarySettings from './useLibrarySettings';

const mockMutateAsync = jest.fn();
const mockSetQueryData = jest.fn();
const mockEnqueueSnackbar = jest.fn();

jest.mock('@tanstack/react-query', () => ({
	useQuery: jest.fn(),
	useMutation: jest.fn(),
	useQueryClient: jest.fn(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => key,
	}),
}));

jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/service/libraries', () => ({
	getLibraries: jest.fn(),
	updateLibrary: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;

describe('components/settings/useLibrarySettings', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseQueryClient.mockReturnValue({
			setQueryData: mockSetQueryData,
		});
		mockedUseQuery.mockReturnValue({
			data: [
				{ category: 'images', path: '/data/Imagens' },
				{ category: 'music', path: '/data/Musicas' },
			],
			isLoading: false,
			isError: false,
		});
		mockedUseMutation.mockReturnValue({
			mutateAsync: mockMutateAsync,
			isPending: false,
		});
	});

	it('exposes ordered libraries and category labels', () => {
		const { result } = renderHook(() => useLibrarySettings());

		expect(result.current.libraries).toHaveLength(4);
		expect(result.current.libraries[0].category).toBe('images');
		expect(result.current.libraries[0].path).toBe('/data/Imagens');
		expect(result.current.getCategoryLabel('videos')).toBe('LIBRARY_VIDEOS');
	});

	it('updates local path state', () => {
		const { result } = renderHook(() => useLibrarySettings());

		act(() => {
			result.current.setPath('documents', '/data/Documentos');
		});

		expect(result.current.libraries[3].path).toBe('/data/Documentos');
	});

	it('saves category and shows success snackbar', async () => {
		mockMutateAsync.mockResolvedValue(undefined);
		const { result } = renderHook(() => useLibrarySettings());

		await act(async () => {
			await result.current.handleSave('images');
		});

		expect(mockMutateAsync).toHaveBeenCalledWith({
			category: 'images',
			path: '/data/Imagens',
		});
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_LIBRARY_SAVED', {
			variant: 'success',
		});
	});

	it('shows error snackbar when save fails', async () => {
		mockMutateAsync.mockRejectedValue(new Error('fail'));
		const { result } = renderHook(() => useLibrarySettings());

		await act(async () => {
			await result.current.handleSave('music');
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_LIBRARY_SAVE_ERROR', {
			variant: 'error',
		});
	});
});
