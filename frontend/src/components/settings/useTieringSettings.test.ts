import { act, renderHook } from '@testing-library/react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { updateTieringSettings } from '@/service/tiering';
import useTieringSettings, { tieringStatusKey } from './useTieringSettings';

const mockEnqueueSnackbar = jest.fn();
const mockInvalidateQueries = jest.fn();

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

jest.mock('@/service/tiering', () => ({
	getTieringSettings: jest.fn(),
	updateTieringSettings: jest.fn(),
	getTieringStatus: jest.fn(),
	getTieringUsage: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;
const mockedUpdateTieringSettings = updateTieringSettings as jest.Mock;

const sampleSettings = {
	enabled: true,
	cold_dir_path: '/mnt/cold',
	min_age_days: 90,
	min_size_bytes: 1048576,
	interval_hours: 24,
};

const sampleStatus = {
	enabled: true,
	has_run: true,
	status: 'completed',
	started_at: '2026-06-12T10:00:00Z',
	ended_at: '2026-06-12T10:05:00Z',
	last_error: '',
};

const sampleUsage = { hot_files: 10, hot_bytes: 1000, cold_files: 4, cold_bytes: 400 };

const mockQueries = (
	overrides: Record<string, { data?: unknown; isLoading?: boolean; isError?: boolean }> = {}
) => {
	mockedUseQuery.mockImplementation(({ queryKey }: { queryKey: string[] }) => {
		const defaults: Record<string, unknown> = {
			'tiering-settings': sampleSettings,
			'tiering-status': sampleStatus,
			'tiering-usage': sampleUsage,
		};
		const key = queryKey[0] as string;
		return {
			data: defaults[key],
			isLoading: false,
			isError: false,
			...(overrides[key] ?? {}),
		};
	});
};

describe('components/settings/useTieringSettings', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseQueryClient.mockReturnValue({ invalidateQueries: mockInvalidateQueries });
		mockQueries();
		mockedUseMutation.mockImplementation(
			({
				mutationFn,
				onSuccess,
			}: {
				mutationFn: (args: unknown) => Promise<unknown>;
				onSuccess?: (result: unknown) => void;
			}) => ({
				mutateAsync: async (args: unknown) => {
					const result = await mutationFn(args);
					onSuccess?.(result);
					return result;
				},
				isPending: false,
			})
		);
	});

	it('mirrors the persisted settings while there is no draft', () => {
		const { result } = renderHook(() => useTieringSettings());

		expect(result.current.form).toEqual(sampleSettings);
		expect(result.current.status).toEqual(sampleStatus);
		expect(result.current.usage).toEqual(sampleUsage);
		expect(result.current.hasUnsavedChanges).toBe(false);
	});

	it('falls back to the defaults before the settings load', () => {
		mockQueries({ 'tiering-settings': { data: undefined } });
		const { result } = renderHook(() => useTieringSettings());

		expect(result.current.form).toEqual({
			enabled: false,
			cold_dir_path: '',
			min_age_days: 90,
			min_size_bytes: 1048576,
			interval_hours: 24,
		});
	});

	it('tracks edits in the draft and saves them', async () => {
		mockedUpdateTieringSettings.mockResolvedValue(sampleSettings);
		const { result } = renderHook(() => useTieringSettings());

		act(() => {
			result.current.setField('cold_dir_path', '/mnt/frio');
		});
		expect(result.current.hasUnsavedChanges).toBe(true);
		expect(result.current.form.cold_dir_path).toBe('/mnt/frio');

		await act(async () => {
			await result.current.handleSave();
		});

		expect(mockedUpdateTieringSettings).toHaveBeenCalledWith({
			...sampleSettings,
			cold_dir_path: '/mnt/frio',
		});
		expect(mockInvalidateQueries).toHaveBeenCalledWith({ queryKey: ['tiering-settings'] });
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_TIERING_SAVED', {
			variant: 'success',
		});
		expect(result.current.hasUnsavedChanges).toBe(false);
	});

	it('shows the backend message verbatim when the save fails', async () => {
		mockedUpdateTieringSettings.mockRejectedValue({
			response: { data: { error: 'diretório inválido' } },
		});
		const { result } = renderHook(() => useTieringSettings());

		await act(async () => {
			await result.current.handleSave();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('diretório inválido', {
			variant: 'error',
		});
	});

	it('falls back to the generic save error without a backend message', async () => {
		mockedUpdateTieringSettings.mockRejectedValue(new Error('network'));
		const { result } = renderHook(() => useTieringSettings());

		await act(async () => {
			await result.current.handleSave();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_TIERING_SAVE_ERROR', {
			variant: 'error',
		});
	});

	it('maps job statuses to i18n keys with a queued fallback', () => {
		expect(tieringStatusKey('completed')).toBe('SETTINGS_TIERING_STATUS_COMPLETED');
		expect(tieringStatusKey('failed')).toBe('SETTINGS_TIERING_STATUS_FAILED');
		expect(tieringStatusKey('partial_fail')).toBe('SETTINGS_TIERING_STATUS_PARTIAL_FAIL');
		expect(tieringStatusKey('running')).toBe('SETTINGS_TIERING_STATUS_RUNNING');
		expect(tieringStatusKey('canceled')).toBe('SETTINGS_TIERING_STATUS_CANCELED');
		expect(tieringStatusKey('whatever')).toBe('SETTINGS_TIERING_STATUS_QUEUED');
	});
});
