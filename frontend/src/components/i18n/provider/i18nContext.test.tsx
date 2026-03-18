import { renderHook } from '@testing-library/react';
import { useI18n } from './i18nContext';

describe('i18n/provider/i18nContext', () => {
    it('throws when used outside provider', () => {
        expect(() => renderHook(() => useI18n())).toThrow(
            'useI18n must be used within an I18nProvider'
        );
    });
});
