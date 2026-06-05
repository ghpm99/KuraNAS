import { renderHook } from '@testing-library/react';
import { useI18n } from './i18nContext';

describe('i18n/provider/i18nContext', () => {
    it('falls back to echoing the key when used outside a provider', () => {
        const { result } = renderHook(() => useI18n());

        expect(result.current.t('SOMETHING_WENT_WRONG')).toBe('SOMETHING_WENT_WRONG');
    });
});
