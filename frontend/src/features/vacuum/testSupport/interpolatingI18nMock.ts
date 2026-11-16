const actualI18nContext = jest.requireActual('@/components/i18n/provider/i18nContext');

const translateToKeyWithArguments = (key: string, options?: Record<string, string>): string =>
    options ? `${key}:${Object.values(options).join(',')}` : key;

export const interpolatingI18nMock = {
    ...actualI18nContext,
    __esModule: true,
    default: () => ({ t: translateToKeyWithArguments }),
};
