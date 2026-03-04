const renderMock = jest.fn();
const createRootMock = jest.fn(() => ({ render: renderMock }));

jest.mock('react-dom/client', () => ({ createRoot: (...args: any[]) => createRootMock(...args) }));
jest.mock('./app/App.tsx', () => () => null);

describe('main entry', () => {
	it('creates root and renders app', async () => {
		document.body.innerHTML = '<div id="root"></div>';
		await import('./main');
		expect(createRootMock).toHaveBeenCalled();
		expect(renderMock).toHaveBeenCalled();
	});
});
