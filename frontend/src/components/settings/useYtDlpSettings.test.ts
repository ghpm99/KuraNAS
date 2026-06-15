import { renderHook } from '@testing-library/react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { updateYtDlp } from '@/service/ingest';
import useYtDlpSettings from './useYtDlpSettings';

const mockEnqueueSnackbar = jest.fn();
const mockInvalidateQueries = jest.fn();

jest.mock('@tanstack/react-query', () => ({
	useQuery: jest.fn(),
	useMutation: jest.fn(),
	useQueryClient: jest.fn(),
}));

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

jest.mock('notistack', () => ({
	useSnackbar: () => ({ enqueueSnackbar: mockEnqueueSnackbar }),
}));

jest.mock('@/service/ingest', () => ({
	getYtDlpStatus: jest.fn(),
	updateYtDlp: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;
const mockedUpdateYtDlp = updateYtDlp as jest.Mock;

const sampleStatus = {
	installed: true,
	current_version: '2024.08.06',
	latest_version: '2024.09.01',
	update_available: true,
	release_url: 'http://x',
	release_date: '2024-09-01',
};

type MutationConfig = {
	mutationFn: () => Promise<unknown>;
	onSuccess?: () => void;
	onError?: (error: unknown) => void;
};

let mutationConfig: MutationConfig;

describe('components/settings/useYtDlpSettings', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseQueryClient.mockReturnValue({ invalidateQueries: mockInvalidateQueries });
		mockedUseQuery.mockReturnValue({ data: sampleStatus, isLoading: false, isError: false });
		mockedUseMutation.mockImplementation((config: MutationConfig) => {
			mutationConfig = config;
			return {
				mutate: () => config.mutationFn(),
				isPending: false,
			};
		});
	});

	it('exposes the status from the query', () => {
		const { result } = renderHook(() => useYtDlpSettings());
		expect(result.current.status).toEqual(sampleStatus);
		expect(result.current.isLoading).toBe(false);
		expect(result.current.hasError).toBe(false);
	});

	it('runs the update and reports success', async () => {
		mockedUpdateYtDlp.mockResolvedValue(undefined);
		const { result } = renderHook(() => useYtDlpSettings());

		result.current.handleUpdate();
		mutationConfig.onSuccess?.();

		expect(mockedUpdateYtDlp).toHaveBeenCalled();
		expect(mockInvalidateQueries).toHaveBeenCalledWith({ queryKey: ['ytdlp-status'] });
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('YTDLP_UPDATE_APPLIED', { variant: 'success' });
	});

	it('shows the backend error message verbatim on failure', () => {
		const { result } = renderHook(() => useYtDlpSettings());
		result.current.handleUpdate();
		mutationConfig.onError?.({ response: { data: { error: 'falha de checksum' } } });

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('falha de checksum', { variant: 'error' });
	});

	it('falls back to a generic error key when none is provided', () => {
		const { result } = renderHook(() => useYtDlpSettings());
		result.current.handleUpdate();
		mutationConfig.onError?.(new Error('boom'));

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_YTDLP_UPDATE', { variant: 'error' });
	});
});
