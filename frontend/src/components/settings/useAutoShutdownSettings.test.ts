import { act, renderHook } from '@testing-library/react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getSuggestedShutdownTime, updateAutoShutdownSettings } from '@/service/autoShutdown';
import useAutoShutdownSettings from './useAutoShutdownSettings';

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

jest.mock('@/service/autoShutdown', () => ({
	getAutoShutdownSettings: jest.fn(),
	updateAutoShutdownSettings: jest.fn(),
	getSuggestedShutdownTime: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;
const mockedUpdate = updateAutoShutdownSettings as jest.Mock;
const mockedSuggest = getSuggestedShutdownTime as jest.Mock;

const sampleSettings = {
	enabled: true,
	time: '03:00',
	grace_period_seconds: 60,
};

const mockSettingsQuery = (
	overrides: { data?: unknown; isLoading?: boolean; isError?: boolean } = {}
) => {
	mockedUseQuery.mockReturnValue({
		data: sampleSettings,
		isLoading: false,
		isError: false,
		...overrides,
	});
};

// Each useMutation call returns a mutateAsync that runs its own mutationFn and
// onSuccess, in declaration order: first the save mutation, then the suggest one.
const mockMutations = () => {
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
};

describe('components/settings/useAutoShutdownSettings', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseQueryClient.mockReturnValue({ invalidateQueries: mockInvalidateQueries });
		mockSettingsQuery();
		mockMutations();
	});

	it('mirrors the persisted settings while there is no draft', () => {
		const { result } = renderHook(() => useAutoShutdownSettings());

		expect(result.current.form).toEqual(sampleSettings);
		expect(result.current.hasUnsavedChanges).toBe(false);
		expect(result.current.suggestion).toBeNull();
	});

	it('falls back to the defaults before the settings load', () => {
		mockSettingsQuery({ data: undefined });
		const { result } = renderHook(() => useAutoShutdownSettings());

		expect(result.current.form).toEqual({
			enabled: false,
			time: '03:00',
			grace_period_seconds: 60,
		});
	});

	it('tracks edits in the draft and saves them', async () => {
		mockedUpdate.mockResolvedValue(sampleSettings);
		const { result } = renderHook(() => useAutoShutdownSettings());

		act(() => {
			result.current.setField('time', '04:30');
		});
		expect(result.current.hasUnsavedChanges).toBe(true);
		expect(result.current.form.time).toBe('04:30');

		await act(async () => {
			await result.current.handleSave();
		});

		expect(mockedUpdate).toHaveBeenCalledWith({ ...sampleSettings, time: '04:30' });
		expect(mockInvalidateQueries).toHaveBeenCalledWith({ queryKey: ['auto-shutdown-settings'] });
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_AUTO_SHUTDOWN_SAVE_SUCCESS', {
			variant: 'success',
		});
		expect(result.current.hasUnsavedChanges).toBe(false);
	});

	it('shows the backend message verbatim when the save fails', async () => {
		mockedUpdate.mockRejectedValue({ response: { data: { error: 'horário inválido' } } });
		const { result } = renderHook(() => useAutoShutdownSettings());

		await act(async () => {
			await result.current.handleSave();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('horário inválido', { variant: 'error' });
	});

	it('falls back to the generic save error without a backend message', async () => {
		mockedUpdate.mockRejectedValue(new Error('network'));
		const { result } = renderHook(() => useAutoShutdownSettings());

		await act(async () => {
			await result.current.handleSave();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_AUTO_SHUTDOWN_SAVE_ERROR', {
			variant: 'error',
		});
	});

	it('applies an available suggestion to the time field', async () => {
		mockedSuggest.mockResolvedValue({ available: true, time: '02:15', sample_size: 9 });
		const { result } = renderHook(() => useAutoShutdownSettings());

		await act(async () => {
			await result.current.handleSuggest();
		});

		expect(result.current.suggestion).toEqual({ available: true, time: '02:15', sample_size: 9 });
		expect(result.current.form.time).toBe('02:15');
		expect(result.current.hasUnsavedChanges).toBe(true);
	});

	it('keeps the time when the suggestion is unavailable', async () => {
		mockedSuggest.mockResolvedValue({ available: false, time: '', sample_size: 1 });
		const { result } = renderHook(() => useAutoShutdownSettings());

		await act(async () => {
			await result.current.handleSuggest();
		});

		expect(result.current.suggestion).toEqual({ available: false, time: '', sample_size: 1 });
		expect(result.current.form.time).toBe('03:00');
		expect(result.current.hasUnsavedChanges).toBe(false);
	});

	it('reports a suggestion failure', async () => {
		mockedSuggest.mockRejectedValue(new Error('network'));
		const { result } = renderHook(() => useAutoShutdownSettings());

		await act(async () => {
			await result.current.handleSuggest();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_AUTO_SHUTDOWN_LOAD_ERROR', {
			variant: 'error',
		});
	});
});
