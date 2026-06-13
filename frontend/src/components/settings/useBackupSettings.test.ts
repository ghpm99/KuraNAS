import { act, renderHook } from '@testing-library/react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { updateBackupSettings } from '@/service/backup';
import useBackupSettings, { backupStatusKey } from './useBackupSettings';

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

jest.mock('@/service/backup', () => ({
	getBackupSettings: jest.fn(),
	updateBackupSettings: jest.fn(),
	getBackupStatus: jest.fn(),
	getBackupPending: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedUseMutation = useMutation as jest.Mock;
const mockedUseQueryClient = useQueryClient as jest.Mock;
const mockedUpdateBackupSettings = updateBackupSettings as jest.Mock;

const sampleSettings = {
	enabled: true,
	destination_path: '/mnt/backup',
	retention_days: 30,
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

const mockQueries = (overrides: Record<string, { data?: unknown; isLoading?: boolean; isError?: boolean }> = {}) => {
	mockedUseQuery.mockImplementation(({ queryKey }: { queryKey: string[] }) => {
		const defaults: Record<string, unknown> = {
			'backup-settings': sampleSettings,
			'backup-status': sampleStatus,
			'backup-pending': { pending_files: 3 },
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

describe('components/settings/useBackupSettings', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockedUseQueryClient.mockReturnValue({ invalidateQueries: mockInvalidateQueries });
		mockQueries();
		// Run the real mutationFn + onSuccess so the hook is exercised end to end.
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
		const { result } = renderHook(() => useBackupSettings());

		expect(result.current.form).toEqual(sampleSettings);
		expect(result.current.status).toEqual(sampleStatus);
		expect(result.current.pendingFiles).toBe(3);
		expect(result.current.hasUnsavedChanges).toBe(false);
	});

	it('falls back to the defaults before the settings load', () => {
		mockQueries({ 'backup-settings': { data: undefined } });
		const { result } = renderHook(() => useBackupSettings());

		expect(result.current.form).toEqual({
			enabled: false,
			destination_path: '',
			retention_days: 30,
			interval_hours: 24,
		});
	});

	it('tracks edits in the draft and saves them', async () => {
		mockedUpdateBackupSettings.mockResolvedValue(sampleSettings);
		const { result } = renderHook(() => useBackupSettings());

		act(() => {
			result.current.setField('destination_path', '/mnt/cold');
		});
		expect(result.current.hasUnsavedChanges).toBe(true);
		expect(result.current.form.destination_path).toBe('/mnt/cold');

		await act(async () => {
			await result.current.handleSave();
		});

		expect(mockedUpdateBackupSettings).toHaveBeenCalledWith({
			...sampleSettings,
			destination_path: '/mnt/cold',
		});
		expect(mockInvalidateQueries).toHaveBeenCalledWith({ queryKey: ['backup-settings'] });
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_BACKUP_SAVED', {
			variant: 'success',
		});
		expect(result.current.hasUnsavedChanges).toBe(false);
	});

	it('shows the backend message verbatim when the save fails', async () => {
		mockedUpdateBackupSettings.mockRejectedValue({
			response: { data: { error: 'destino inválido' } },
		});
		const { result } = renderHook(() => useBackupSettings());

		await act(async () => {
			await result.current.handleSave();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('destino inválido', {
			variant: 'error',
		});
	});

	it('falls back to the generic save error without a backend message', async () => {
		mockedUpdateBackupSettings.mockRejectedValue(new Error('network'));
		const { result } = renderHook(() => useBackupSettings());

		await act(async () => {
			await result.current.handleSave();
		});

		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('SETTINGS_BACKUP_SAVE_ERROR', {
			variant: 'error',
		});
	});

	it('maps job statuses to i18n keys with a queued fallback', () => {
		expect(backupStatusKey('completed')).toBe('SETTINGS_BACKUP_STATUS_COMPLETED');
		expect(backupStatusKey('failed')).toBe('SETTINGS_BACKUP_STATUS_FAILED');
		expect(backupStatusKey('partial_fail')).toBe('SETTINGS_BACKUP_STATUS_PARTIAL_FAIL');
		expect(backupStatusKey('running')).toBe('SETTINGS_BACKUP_STATUS_RUNNING');
		expect(backupStatusKey('canceled')).toBe('SETTINGS_BACKUP_STATUS_CANCELED');
		expect(backupStatusKey('whatever')).toBe('SETTINGS_BACKUP_STATUS_QUEUED');
	});
});
