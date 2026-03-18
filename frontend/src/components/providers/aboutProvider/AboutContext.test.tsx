import { renderHook } from '@testing-library/react';
import { useAbout } from './AboutContext';

describe('providers/aboutProvider/AboutContext', () => {
    it('throws when used outside provider', () => {
        expect(() => renderHook(() => useAbout())).toThrow(
            'useAbout must be used within an AboutProvider'
        );
    });
});
