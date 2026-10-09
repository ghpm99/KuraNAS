import { renderHook } from '@testing-library/react';
import { useDocumentTitle } from './useDocumentTitle';

describe('useDocumentTitle', () => {
    beforeEach(() => {
        document.title = 'original';
    });

    it('mounts without any provider and still composes a title', () => {
        renderHook(() => useDocumentTitle('Arquivos'));

        expect(document.title).toBe('Arquivos · APP_NAME');
    });

    it('updates when the page title changes and restores the previous title on unmount', () => {
        const { rerender, unmount } = renderHook(({ pageTitle }) => useDocumentTitle(pageTitle), {
            initialProps: { pageTitle: 'Arquivos' },
        });

        rerender({ pageTitle: 'Imagens' });
        expect(document.title).toBe('Imagens · APP_NAME');

        unmount();
        expect(document.title).toBe('original');
    });

    it('uses only the app name when the page title is empty', () => {
        renderHook(() => useDocumentTitle(''));

        expect(document.title).toBe('APP_NAME');
    });
});
