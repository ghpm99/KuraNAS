import { renderHook } from '@testing-library/react';
import { useFile } from './fileContext';

describe('providers/fileProvider/fileContext', () => {
    it('throws when used outside provider', () => {
        expect(() => renderHook(() => useFile())).toThrow(
            'useFile must be used within a FileProvider'
        );
    });
});
