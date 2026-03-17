import { renderHook } from '@testing-library/react';
import { I18nContextProvider, useI18n } from './i18nContext';

describe('i18n/provider/i18nContext', () => {
    it('throws when used outside provider', () => {
        expect(() => renderHook(() => useI18n())).toThrow(
            'useI18n must be used within an I18nProvider'
        );
    });

    it('returns translator from provider', () => {
        const wrapper = ({ children }: { children: React.ReactNode }) => (
            <I18nContextProvider value={{ t: (key: string) => `x-${key}` }}>
                {children}
            </I18nContextProvider>
        );
        const { result } = renderHook(() => useI18n(), { wrapper });

        expect(result.current.t('HELLO')).toBe('x-HELLO');
    });
});
