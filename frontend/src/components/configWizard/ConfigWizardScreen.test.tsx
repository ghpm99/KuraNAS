import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import ConfigWizardScreen from './ConfigWizardScreen';
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
        { key: 'ENABLE_WORKERS', group: 'general', kind: 'bool', value: 'false', configured: true, dangerous: false },
        { key: 'ENTRY_POINT', group: 'general', kind: 'string', value: '/data', configured: true, dangerous: false },
        { key: 'LOG_MAX_SIZE_MB', group: 'general', kind: 'int', value: '50', configured: true, dangerous: false },
        { key: 'DB_HOST', group: 'database', kind: 'string', value: 'localhost', configured: true, dangerous: true },
        { key: 'DB_PASSWORD', group: 'database', kind: 'secret', value: '', configured: true, dangerous: true },
        { key: 'ALLOWED_ORIGINS', group: 'access', kind: 'string', value: 'http://localhost', configured: true, dangerous: true },
        { key: 'EMAIL_TOKEN_KEY', group: 'email', kind: 'secret', value: '', configured: false, dangerous: true },
        { key: 'AI_OPENAI_API_KEY', group: 'ai', kind: 'secret', value: '', configured: false, dangerous: false },
        { key: 'WORKER_MAX_CONCURRENT_JOBS', group: 'workers', kind: 'int', value: '4', configured: true, dangerous: false },
    ],
};

const renderWizard = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={queryClient}>
            <SnackbarProvider>
                <MemoryRouter>
                    <ConfigWizardScreen />
                </MemoryRouter>
            </SnackbarProvider>
        </QueryClientProvider>
    );
};

const clickNext = (times: number) => {
    for (let i = 0; i < times; i += 1) {
        fireEvent.click(screen.getByRole('button', { name: 'ENV_NEXT' }));
    }
};

describe('components/configWizard/ConfigWizardScreen', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('renders an error fallback when the backend rejects (loopback-gated)', async () => {
        mockedGetEnvConfig.mockRejectedValue(new Error('forbidden'));

        renderWizard();

        await waitFor(() => {
            expect(screen.getByText('ENV_WIZARD_LOAD_ERROR')).toBeInTheDocument();
        });
        expect(screen.getByText('ENV_WIZARD_LOOPBACK_HELP')).toBeInTheDocument();
    });

    it('renders the first step fields and edits a value', async () => {
        mockedGetEnvConfig.mockResolvedValue(sampleConfig);

        renderWizard();

        await waitFor(() => {
            expect(screen.getByLabelText('ENV_FIELD_LANGUAGE')).toBeInTheDocument();
        });

        fireEvent.change(screen.getByLabelText('ENV_FIELD_LANGUAGE'), {
            target: { value: 'en-US' },
        });
        expect(screen.getByLabelText('ENV_FIELD_LANGUAGE')).toHaveValue('en-US');

        const workersSwitch = screen.getByLabelText('ENV_FIELD_ENABLE_WORKERS');
        fireEvent.click(workersSwitch);
        expect(workersSwitch).toBeChecked();
        fireEvent.click(workersSwitch);
        expect(workersSwitch).not.toBeChecked();
    });

    it('runs a path test on the general step', async () => {
        mockedGetEnvConfig.mockResolvedValue(sampleConfig);
        mockedTestEnvPath.mockResolvedValue({ ok: true, message: 'path-ok' });

        renderWizard();
        await waitFor(() => screen.getByLabelText('ENV_FIELD_ENTRY_POINT'));

        fireEvent.click(screen.getByRole('button', { name: 'ENV_TEST_PATH_BUTTON' }));

        await waitFor(() => {
            expect(screen.getByText('path-ok')).toBeInTheDocument();
        });
        expect(mockedTestEnvPath).toHaveBeenCalledWith({ path: '/data' });
    });

    it('runs a database test on the database step', async () => {
        mockedGetEnvConfig.mockResolvedValue(sampleConfig);
        mockedTestEnvDatabase.mockResolvedValue({ ok: false, message: 'db-failed' });

        renderWizard();
        await waitFor(() => screen.getByLabelText('ENV_FIELD_LANGUAGE'));

        clickNext(1);
        fireEvent.click(screen.getByRole('button', { name: 'ENV_TEST_DB_BUTTON' }));

        await waitFor(() => {
            expect(screen.getByText('db-failed')).toBeInTheDocument();
        });
        expect(mockedTestEnvDatabase).toHaveBeenCalledWith({
            host: 'localhost',
            port: '',
            user: '',
            name: '',
            password: '',
        });
    });

    it('generates an email token key', async () => {
        mockedGetEnvConfig.mockResolvedValue(sampleConfig);

        renderWizard();
        await waitFor(() => screen.getByLabelText('ENV_FIELD_LANGUAGE'));

        clickNext(3);
        const tokenInput = screen.getByLabelText('ENV_FIELD_EMAIL_TOKEN_KEY');
        fireEvent.change(tokenInput, { target: { value: 'typed-secret' } });
        expect(tokenInput).toHaveValue('typed-secret');

        fireEvent.click(screen.getByRole('button', { name: 'ENV_GENERATE_TOKEN_KEY' }));

        expect(screen.getByLabelText('ENV_FIELD_EMAIL_TOKEN_KEY')).not.toHaveValue('typed-secret');
        expect(screen.getByText('ENV_TOKEN_KEY_WARNING')).toBeInTheDocument();
    });

    it('requires confirmation before saving a dangerous change', async () => {
        mockedGetEnvConfig.mockResolvedValue(sampleConfig);
        mockedUpdateEnvConfig.mockResolvedValue({ ...sampleConfig, restart_required: true });

        renderWizard();
        await waitFor(() => screen.getByLabelText('ENV_FIELD_LANGUAGE'));

        clickNext(1);
        fireEvent.change(screen.getByLabelText('ENV_FIELD_DB_HOST'), {
            target: { value: 'db.internal' },
        });
        fireEvent.change(screen.getByLabelText('ENV_FIELD_DB_PASSWORD'), {
            target: { value: 'new-secret' },
        });

        clickNext(5);

        // Secret changes are masked in the review list.
        expect(screen.getByText('••••••')).toBeInTheDocument();

        const saveButton = screen.getByRole('button', { name: 'ENV_SAVE' });
        expect(saveButton).toBeDisabled();

        fireEvent.click(screen.getByLabelText('ENV_CONFIRM_DANGEROUS'));
        expect(saveButton).toBeEnabled();

        fireEvent.click(saveButton);
        await waitFor(() => {
            expect(mockedUpdateEnvConfig).toHaveBeenCalledWith({
                changes: { DB_HOST: 'db.internal', DB_PASSWORD: 'new-secret' },
                confirmed: true,
            });
        });
    });

    it('shows an empty review and a disabled save with no changes', async () => {
        mockedGetEnvConfig.mockResolvedValue(sampleConfig);

        renderWizard();
        await waitFor(() => screen.getByLabelText('ENV_FIELD_LANGUAGE'));

        clickNext(6);

        expect(screen.getByText('ENV_REVIEW_NO_CHANGES')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'ENV_SAVE' })).toBeDisabled();

        fireEvent.click(screen.getByRole('button', { name: 'ENV_BACK' }));
        expect(screen.getByLabelText('ENV_FIELD_WORKER_MAX_CONCURRENT_JOBS')).toBeInTheDocument();
    });

    it('shows the restart banner when a write is pending', async () => {
        mockedGetEnvConfig.mockResolvedValue({ ...sampleConfig, restart_required: true });

        renderWizard();

        await waitFor(() => {
            expect(screen.getByText('ENV_RESTART_REQUIRED')).toBeInTheDocument();
        });
    });
});
