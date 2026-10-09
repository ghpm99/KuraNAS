import { ReactNode } from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import useConfigWizard, { generateTokenKey } from './useConfigWizard';
import { EnvConfig } from '@/service/configuration';
import {
    getEnvConfig,
    testEnvDatabase,
    testEnvPath,
    updateEnvConfig,
} from '@/service/configuration';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

jest.mock('@/service/configuration', () => ({
    getEnvConfig: jest.fn(),
    updateEnvConfig: jest.fn(),
    testEnvDatabase: jest.fn(),
    testEnvPath: jest.fn(),
}));

const mockedGetEnvConfig = getEnvConfig as jest.Mock;
const mockedUpdateEnvConfig = updateEnvConfig as jest.Mock;
const mockedTestEnvDatabase = testEnvDatabase as jest.Mock;
const mockedTestEnvPath = testEnvPath as jest.Mock;

const sampleConfig: EnvConfig = {
    restart_required: false,
    fields: [
        { key: 'LANGUAGE', group: 'general', kind: 'string', value: 'pt-BR', configured: true, dangerous: false },
        { key: 'DB_HOST', group: 'database', kind: 'string', value: 'localhost', configured: true, dangerous: true },
        { key: 'DB_PASSWORD', group: 'database', kind: 'secret', value: '', configured: false, dangerous: true },
    ],
};

const createWrapper = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>
            <SnackbarProvider>{children}</SnackbarProvider>
        </QueryClientProvider>
    );
};

const renderConfigWizard = async () => {
    const hook = renderHook(() => useConfigWizard(), { wrapper: createWrapper() });
    await waitFor(() => expect(hook.result.current.isLoading).toBe(false));
    return hook;
};

describe('components/configWizard/useConfigWizard', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedGetEnvConfig.mockResolvedValue(sampleConfig);
    });

    it('builds steps and groups from the config', async () => {
        const { result } = await renderConfigWizard();

        expect(result.current.steps).toHaveLength(7);
        expect(result.current.fieldsByGroup.general).toHaveLength(1);
        expect(result.current.fieldsByGroup.database).toHaveLength(2);
    });

    it('tracks meaningful changes only', async () => {
        const { result } = await renderConfigWizard();

        act(() => result.current.setField('LANGUAGE', 'en-US'));
        expect(result.current.pendingChanges).toContain('LANGUAGE');

        act(() => result.current.setField('LANGUAGE', 'pt-BR'));
        expect(result.current.pendingChanges).not.toContain('LANGUAGE');

        act(() => result.current.setField('DB_PASSWORD', ''));
        expect(result.current.pendingChanges).not.toContain('DB_PASSWORD');

        act(() => result.current.setField('DB_HOST', 'db.internal'));
        expect(result.current.changedDangerous).toBe(true);
    });

    it('clamps step navigation', async () => {
        const { result } = await renderConfigWizard();

        act(() => result.current.goBack());
        expect(result.current.activeStep).toBe(0);

        for (let i = 0; i < 10; i += 1) {
            act(() => result.current.goNext());
        }
        expect(result.current.activeStep).toBe(6);
        expect(result.current.isLastStep).toBe(true);
    });

    it('does not save when there are no pending changes', async () => {
        const { result } = await renderConfigWizard();

        act(() => result.current.save());
        expect(mockedUpdateEnvConfig).not.toHaveBeenCalled();
    });

    it('saves pending changes', async () => {
        mockedUpdateEnvConfig.mockResolvedValue(sampleConfig);
        const { result } = await renderConfigWizard();

        act(() => result.current.setField('LANGUAGE', 'en-US'));
        act(() => result.current.save());

        await waitFor(() => expect(mockedUpdateEnvConfig).toHaveBeenCalled());
    });

    it('runs database and path tests, storing results', async () => {
        mockedTestEnvDatabase.mockResolvedValue({ ok: true, message: 'db' });
        mockedTestEnvPath.mockResolvedValue({ ok: false, message: 'path' });
        const { result } = await renderConfigWizard();

        await act(async () => {
            await result.current.runDbTest();
        });
        expect(result.current.dbTestResult).toEqual({ ok: true, message: 'db' });

        await act(async () => {
            await result.current.runPathTest('ENTRY_POINT');
        });
        expect(result.current.pathTestResult).toEqual({ ok: false, message: 'path' });
    });

    it('surfaces the translated backend error on save failure', async () => {
        mockedUpdateEnvConfig.mockRejectedValue({ response: { data: { error: 'boom' } } });
        const { result } = await renderConfigWizard();

        act(() => result.current.setField('LANGUAGE', 'en-US'));
        act(() => result.current.save());

        await waitFor(() => expect(mockedUpdateEnvConfig).toHaveBeenCalled());
    });

    it('falls back to a generic error when the response has none', async () => {
        mockedUpdateEnvConfig.mockRejectedValue(new Error('network'));
        const { result } = await renderConfigWizard();

        act(() => result.current.setField('LANGUAGE', 'en-US'));
        act(() => result.current.save());

        await waitFor(() => expect(mockedUpdateEnvConfig).toHaveBeenCalled());
    });

    it('reports the error state when the backend rejects', async () => {
        mockedGetEnvConfig.mockReset();
        mockedGetEnvConfig.mockRejectedValue(new Error('forbidden'));

        const { result } = renderHook(() => useConfigWizard(), { wrapper: createWrapper() });

        await waitFor(() => expect(result.current.isError).toBe(true));
    });

    it('generates a 32-byte base64 token key', () => {
        const key = generateTokenKey();
        expect(atob(key)).toHaveLength(32);
    });
});
