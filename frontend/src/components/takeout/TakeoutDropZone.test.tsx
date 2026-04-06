import { render, screen, fireEvent } from '@testing-library/react';
import TakeoutDropZone from './TakeoutDropZone';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (key: string) => key }),
}));

const getZone = () => screen.getByText('TAKEOUT_DRAG_DROP').closest('[role="button"]') as HTMLElement;

const createFileList = (files: File[]): FileList => ({
	length: files.length,
	item: (i: number) => files[i] ?? null,
	...Object.fromEntries(files.map((f, i) => [i, f])),
	[Symbol.iterator]: files[Symbol.iterator].bind(files),
} as unknown as FileList);

describe('TakeoutDropZone', () => {
	const onSelectFile = jest.fn();

	beforeEach(() => {
		jest.clearAllMocks();
	});

	it('renders instruction text', () => {
		render(<TakeoutDropZone onSelectFile={onSelectFile} />);
		expect(screen.getByText('TAKEOUT_DRAG_DROP')).toBeInTheDocument();
		expect(screen.getAllByText('TAKEOUT_SELECT_FILE').length).toBeGreaterThan(0);
	});

	it('calls onSelectFile when a file is dropped', () => {
		render(<TakeoutDropZone onSelectFile={onSelectFile} />);
		const zone = getZone();
		const file = new File(['content'], 'takeout.zip', { type: 'application/zip' });

		fireEvent.drop(zone, {
			dataTransfer: { files: createFileList([file]) },
		});

		expect(onSelectFile).toHaveBeenCalledWith(file);
	});

	it('does not call onSelectFile when drop has no files', () => {
		render(<TakeoutDropZone onSelectFile={onSelectFile} />);
		const zone = getZone();

		fireEvent.drop(zone, {
			dataTransfer: { files: createFileList([]) },
		});

		expect(onSelectFile).not.toHaveBeenCalled();
	});

	it('handles file input change', () => {
		render(<TakeoutDropZone onSelectFile={onSelectFile} />);
		const input = document.querySelector('input[type="file"]') as HTMLInputElement;
		const file = new File(['data'], 'test.zip', { type: 'application/zip' });
		const fileList = { length: 1, item: (i: number) => (i === 0 ? file : null), 0: file } as unknown as FileList;

		Object.defineProperty(input, 'files', { value: fileList, writable: false });
		fireEvent.change(input);
		expect(onSelectFile).toHaveBeenCalledWith(file);
	});

	it('handles keyboard Enter to open file dialog', () => {
		render(<TakeoutDropZone onSelectFile={onSelectFile} />);
		const zone = getZone();
		const input = document.querySelector('input[type="file"]') as HTMLInputElement;
		const clickSpy = jest.spyOn(input, 'click');

		fireEvent.keyDown(zone, { key: 'Enter' });
		expect(clickSpy).toHaveBeenCalled();
	});

	it('handles keyboard Space to open file dialog', () => {
		render(<TakeoutDropZone onSelectFile={onSelectFile} />);
		const zone = getZone();
		const input = document.querySelector('input[type="file"]') as HTMLInputElement;
		const clickSpy = jest.spyOn(input, 'click');

		fireEvent.keyDown(zone, { key: ' ' });
		expect(clickSpy).toHaveBeenCalled();
	});

	it('prevents default on drag over', () => {
		render(<TakeoutDropZone onSelectFile={onSelectFile} />);
		const zone = getZone();
		const event = new Event('dragover', { bubbles: true });
		const preventDefaultSpy = jest.spyOn(event, 'preventDefault');

		zone.dispatchEvent(event);
		expect(preventDefaultSpy).toHaveBeenCalled();
	});

	it('ignores other key presses', () => {
		render(<TakeoutDropZone onSelectFile={onSelectFile} />);
		const zone = getZone();
		const input = document.querySelector('input[type="file"]') as HTMLInputElement;
		const clickSpy = jest.spyOn(input, 'click');

		fireEvent.keyDown(zone, { key: 'Tab' });
		expect(clickSpy).not.toHaveBeenCalled();
	});
});
