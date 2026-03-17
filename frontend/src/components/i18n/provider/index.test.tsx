import { render, screen } from '@testing-library/react';
import I18nProvider from './index';
import useI18n from './i18nContext';
import { useQuery } from '@tanstack/react-query';
import { apiBase } from '@/service';

jest.mock('@/service', () => ({
    apiBase: {
        get: jest.fn(),
    },
}));

jest.mock('@tanstack/react-query', () => ({
    useQuery: jest.fn(),
}));

const mockedUseQuery = useQuery as jest.Mock;
const mockedApiGet = apiBase.get as jest.Mock;

function Consumer({ keyName, options }: { keyName: string; options?: Record<string, string> }) {
    const { t } = useI18n();
    return <span data-testid="value">{t(keyName, options)}</span>;
}

describe('i18n/provider/index', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedApiGet.mockResolvedValue({ data: { HELLO: 'Ola' } });
    });

    it('falls back to key while query is not successful', () => {
        mockedUseQuery.mockReturnValue({ status: 'loading', data: undefined });

        render(
            <I18nProvider>
                <Consumer keyName="HELLO" />
            </I18nProvider>
        );

        expect(screen.getByTestId('value')).toHaveTextContent('HELLO');
    });

    it('translates and interpolates placeholders', () => {
        mockedUseQuery.mockReturnValue({
            status: 'success',
            data: {
                GREETING: 'Ola, {{name}}!',
            },
        });

        render(
            <I18nProvider>
                <Consumer keyName="GREETING" options={{ name: 'Joao' }} />
            </I18nProvider>
        );

        expect(screen.getByTestId('value')).toHaveTextContent('Ola, Joao!');
    });

    it('returns key when translation entry is missing', () => {
        mockedUseQuery.mockReturnValue({
            status: 'success',
            data: { KNOWN: 'Known value' },
        });

        render(
            <I18nProvider>
                <Consumer keyName="UNKNOWN" />
            </I18nProvider>
        );

        expect(screen.getByTestId('value')).toHaveTextContent('UNKNOWN');
    });

    it('executes useQuery queryFn to fetch translations', async () => {
        mockedUseQuery.mockReturnValue({ status: 'loading', data: undefined });

        render(
            <I18nProvider>
                <Consumer keyName="HELLO" />
            </I18nProvider>
        );

        const options = mockedUseQuery.mock.calls[0][0];
        await expect(options.queryFn()).resolves.toEqual({ HELLO: 'Ola' });
        expect(mockedApiGet).toHaveBeenCalledWith('/configuration/translation');
    });
});
