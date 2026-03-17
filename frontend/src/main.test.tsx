const renderMock = jest.fn();
const createRootMock = jest.fn((_container: Element | DocumentFragment) => ({
    render: renderMock,
}));

jest.mock('react-dom/client', () => ({
    createRoot: (container: Element | DocumentFragment) => createRootMock(container),
}));
jest.mock('./app/App.tsx', () => () => null);

describe('main entry', () => {
    it('creates root and renders app', async () => {
        document.body.innerHTML = '<div id="root"></div>';
        await import('./main');
        expect(createRootMock).toHaveBeenCalled();
        expect(renderMock).toHaveBeenCalled();
    });
});
