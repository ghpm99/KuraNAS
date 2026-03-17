import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import ActionBar from './actionBar';
import { FileType } from '@/utils';

const mockUseFile = jest.fn();
const mockNavigate = jest.fn();
const mockEnqueueSnackbar = jest.fn();
const mockDownloadFileBlob = jest.fn();

const createFileContext = (overrides = {}) => ({
	selectedItem: null,
	uploadFiles: jest.fn(),
	createFolder: jest.fn().mockResolvedValue(undefined),
	movePath: jest.fn().mockResolvedValue(undefined),
	copyPath: jest.fn().mockResolvedValue(undefined),
	renamePath: jest.fn().mockResolvedValue(undefined),
	deletePath: jest.fn().mockResolvedValue(undefined),
	rescanFiles: jest.fn(),
	fileListFilter: 'recent',
	...overrides,
});

jest.mock('../providers/fileProvider/fileContext', () => ({
	__esModule: true,
	default: () => mockUseFile(),
}));

jest.mock('../i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => {
			const map: Record<string, string> = {
				FILES: 'FILES',
				RECENT_FILES: 'RECENT_FILES',
				STARRED_FILES: 'STARRED_FILES',
				NEW_FILE: 'NEW_FILE',
				UPLOAD_FILE: 'UPLOAD_FILE',
				NEW_FOLDER: 'NEW_FOLDER',
				MOVE: 'MOVE',
				COPY: 'COPY',
				RENAME: 'RENAME',
				DELETE: 'DELETE',
				DOWNLOAD: 'DOWNLOAD',
				NAME: 'NAME',
				PATH: 'PATH',
				ACTION_CANCEL: 'ACTION_CANCEL',
				CONFIRM_DELETE: 'CONFIRM_DELETE',
				ACTION_CREATE_FOLDER_SUCCESS: 'ACTION_CREATE_FOLDER_SUCCESS',
				ACTION_COPY_SUCCESS: 'ACTION_COPY_SUCCESS',
				ACTION_RENAME_SUCCESS: 'ACTION_RENAME_SUCCESS',
				ACTION_DELETE_SUCCESS: 'ACTION_DELETE_SUCCESS',
				ERROR_LOADING_FILES: 'ERROR_LOADING_FILES',
				COPY_SUFFIX: '_copy',
			};
			return map[key] ?? key;
		},
	}),
}));

jest.mock('react-router-dom', () => {
	const actual = jest.requireActual('react-router-dom');
	return {
		...actual,
		useNavigate: () => mockNavigate,
	};
});

jest.mock('notistack', () => ({
	useSnackbar: () => ({
		enqueueSnackbar: mockEnqueueSnackbar,
	}),
}));

jest.mock('@/service/files', () => ({
	downloadFileBlob: (...args: unknown[]) => mockDownloadFileBlob(...args),
}));

describe('components/actionBar', () => {
	beforeEach(() => {
		jest.clearAllMocks();
		mockDownloadFileBlob.mockResolvedValue(new Blob(['file']));
		mockUseFile.mockReturnValue(createFileContext());
	});

it('shows the filtered list title and creates folders from the dialog', async () => {
		const createFolder = jest.fn().mockResolvedValue(undefined);
		mockUseFile.mockReturnValue(createFileContext({ createFolder }));

		render(<ActionBar />);

		expect(screen.getByText('RECENT_FILES')).toBeInTheDocument();
		expect(screen.queryByRole('button', { name: 'MOVE' })).not.toBeInTheDocument();

		fireEvent.click(screen.getByRole('button', { name: 'NEW_FOLDER' }));

		const dialog = screen.getByRole('dialog');
		fireEvent.change(within(dialog).getByLabelText('NAME'), { target: { value: 'Docs' } });
		fireEvent.click(within(dialog).getAllByRole('button', { name: 'NEW_FOLDER' })[0]!);

		await waitFor(() => {
			expect(createFolder).toHaveBeenCalledWith('Docs', undefined);
		});
		expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ACTION_CREATE_FOLDER_SUCCESS', { variant: 'success' });
	});

it('opens copy/rename/delete flows and downloads the selected file', async () => {
		const movePath = jest.fn().mockResolvedValue(undefined);
		const copyPath = jest.fn().mockResolvedValue(undefined);
		const renamePath = jest.fn().mockResolvedValue(undefined);
		const deletePath = jest.fn().mockResolvedValue(undefined);
		Object.assign(URL, {
			createObjectURL: URL.createObjectURL ?? jest.fn(),
			revokeObjectURL: URL.revokeObjectURL ?? jest.fn(),
		});
		const createObjectURLSpy = jest.spyOn(URL, 'createObjectURL').mockReturnValue('blob:url');
		const revokeObjectURLSpy = jest.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined);
		const clickSpy = jest.fn();
		const removeSpy = jest.fn();
		const originalCreateElement = document.createElement.bind(document);
		const createElementSpy = jest.spyOn(document, 'createElement').mockImplementation((tagName: string) => {
			if (tagName === 'a') {
				const anchor = originalCreateElement('a');
				anchor.click = clickSpy;
				anchor.remove = removeSpy;
				return anchor;
			}

			return originalCreateElement(tagName);
		});

		mockUseFile.mockReturnValue(createFileContext({
			selectedItem: {
				id: 7,
				name: 'movie.mp4',
				path: '/media/movie.mp4',
				parent_path: '/media',
				type: FileType.File,
			},
			fileListFilter: 'all',
			movePath,
			copyPath,
			renamePath,
			deletePath,
		}));

		render(<ActionBar />);

		fireEvent.click(screen.getByRole('button', { name: 'MOVE' }));
		let dialog = screen.getByRole('dialog');
		expect(within(dialog).getByLabelText('PATH')).toHaveValue('/media');
		fireEvent.change(within(dialog).getByLabelText('PATH'), { target: { value: '/archive/' } });
		fireEvent.click(within(dialog).getAllByRole('button', { name: 'MOVE' })[0]!);
		await waitFor(() => {
			expect(movePath).toHaveBeenCalledWith('/media/movie.mp4', '/archive/movie.mp4');
		});
		await waitFor(() => {
			expect(screen.queryByRole('dialog', { name: 'MOVE' })).not.toBeInTheDocument();
		});

		fireEvent.click(screen.getByRole('button', { name: 'COPY' }));
		dialog = screen.getByRole('dialog');
		expect(within(dialog).getByLabelText('PATH')).toHaveValue('/media/movie.mp4_copy');
	fireEvent.change(within(dialog).getByLabelText('PATH'), { target: { value: '/target/movie.mp4' } });
	fireEvent.click(within(dialog).getAllByRole('button', { name: 'COPY' })[0]!);
	await waitFor(() => {
		expect(copyPath).toHaveBeenCalledWith('/media/movie.mp4', '/target/movie.mp4');
	});
	await waitFor(() => {
		expect(screen.queryByRole('dialog', { name: 'COPY' })).not.toBeInTheDocument();
	});

		fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));
		dialog = screen.getByRole('dialog');
		fireEvent.change(within(dialog).getByLabelText('NAME'), { target: { value: 'movie-new.mp4' } });
		fireEvent.click(within(dialog).getAllByRole('button', { name: 'RENAME' })[0]!);
		await waitFor(() => {
			expect(renamePath).toHaveBeenCalledWith('/media/movie.mp4', 'movie-new.mp4');
		});
		await waitFor(() => {
			expect(screen.queryByRole('dialog', { name: 'RENAME' })).not.toBeInTheDocument();
		});

		fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));
		dialog = screen.getByRole('dialog');
		fireEvent.click(within(dialog).getAllByRole('button', { name: 'DELETE' })[0]!);
		await waitFor(() => {
			expect(deletePath).toHaveBeenCalledWith('/media/movie.mp4');
		});
		await waitFor(() => {
			expect(screen.queryByRole('dialog', { name: 'DELETE' })).not.toBeInTheDocument();
		});

	fireEvent.click(screen.getByRole('button', { name: 'DOWNLOAD' }));
	await waitFor(() => {
		expect(mockDownloadFileBlob).toHaveBeenCalledWith(7);
	});

		expect(createObjectURLSpy).toHaveBeenCalled();
		expect(clickSpy).toHaveBeenCalled();
		expect(removeSpy).toHaveBeenCalled();
expect(revokeObjectURLSpy).toHaveBeenCalledWith('blob:url');

		createElementSpy.mockRestore();
		createObjectURLSpy.mockRestore();
		revokeObjectURLSpy.mockRestore();
	});
});

it('shows error snackbars when operations fail', async () => {
	const error = new Error('boom');
	const context = createFileContext({
		selectedItem: {
			id: 7,
			name: 'movie.mp4',
			path: '/media/movie.mp4',
			parent_path: '/media',
			type: FileType.File,
		},
		createFolder: jest.fn().mockRejectedValue(error),
		movePath: jest.fn().mockRejectedValue(error),
		copyPath: jest.fn().mockRejectedValue(error),
		renamePath: jest.fn().mockRejectedValue(error),
		deletePath: jest.fn().mockRejectedValue(error),
	});

	mockUseFile.mockReturnValue(context);

	render(<ActionBar />);

	fireEvent.click(screen.getByRole('button', { name: 'NEW_FOLDER' }));
	let dialog = screen.getByRole('dialog', { name: 'NEW_FOLDER' });
	fireEvent.change(within(dialog).getByLabelText('NAME'), { target: { value: 'Docs' } });
	fireEvent.click(within(dialog).getAllByRole('button', { name: 'NEW_FOLDER' })[0]!);
	await waitFor(() => expect(context.createFolder).toHaveBeenCalled());
	await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_CREATE_FOLDER_FAILED', { variant: 'error' }));
	fireEvent.click(within(dialog).getByRole('button', { name: 'ACTION_CANCEL' }));
	await waitFor(() => expect(screen.queryByRole('dialog', { name: 'NEW_FOLDER' })).not.toBeInTheDocument());

	fireEvent.click(screen.getByRole('button', { name: 'MOVE' }));
	dialog = screen.getByRole('dialog', { name: 'MOVE' });
	fireEvent.change(within(dialog).getByLabelText('PATH'), { target: { value: '/archive/' } });
	fireEvent.click(within(dialog).getAllByRole('button', { name: 'MOVE' })[0]!);
	await waitFor(() => expect(context.movePath).toHaveBeenCalled());
	await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_MOVE_FAILED', { variant: 'error' }));
	fireEvent.click(within(dialog).getByRole('button', { name: 'ACTION_CANCEL' }));
	await waitFor(() => expect(screen.queryByRole('dialog', { name: 'MOVE' })).not.toBeInTheDocument());

	fireEvent.click(screen.getByRole('button', { name: 'COPY' }));
	dialog = screen.getByRole('dialog', { name: 'COPY' });
	fireEvent.click(within(dialog).getAllByRole('button', { name: 'COPY' })[0]!);
	await waitFor(() => expect(context.copyPath).toHaveBeenCalled());
	await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_COPY_FAILED', { variant: 'error' }));
	fireEvent.click(within(dialog).getByRole('button', { name: 'ACTION_CANCEL' }));
	await waitFor(() => expect(screen.queryByRole('dialog', { name: 'COPY' })).not.toBeInTheDocument());

	fireEvent.click(screen.getByRole('button', { name: 'RENAME' }));
	dialog = screen.getByRole('dialog', { name: 'RENAME' });
	fireEvent.change(within(dialog).getByLabelText('NAME'), { target: { value: 'movie-new.mp4' } });
	fireEvent.click(within(dialog).getAllByRole('button', { name: 'RENAME' })[0]!);
	await waitFor(() => expect(context.renamePath).toHaveBeenCalled());
	await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_RENAME_FAILED', { variant: 'error' }));
	fireEvent.click(within(dialog).getByRole('button', { name: 'ACTION_CANCEL' }));
	await waitFor(() => expect(screen.queryByRole('dialog', { name: 'RENAME' })).not.toBeInTheDocument());

	fireEvent.click(screen.getByRole('button', { name: 'DELETE' }));
	dialog = screen.getByRole('dialog', { name: 'DELETE' });
	fireEvent.click(within(dialog).getAllByRole('button', { name: 'DELETE' })[0]!);
	await waitFor(() => expect(context.deletePath).toHaveBeenCalled());
	await waitFor(() => expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_DELETE_FAILED', { variant: 'error' }));
	fireEvent.click(within(dialog).getByRole('button', { name: 'ACTION_CANCEL' }));
	await waitFor(() => expect(screen.queryByRole('dialog', { name: 'DELETE' })).not.toBeInTheDocument());
});

it('shows upload errors and resets the file input', async () => {
	const uploadFiles = jest.fn().mockRejectedValue(new Error('upload failed'));
	mockUseFile.mockReturnValue(createFileContext({ uploadFiles }));

	render(<ActionBar />);

	const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
	const blob = new File(['content'], 'doc.txt', { type: 'text/plain' });
	fireEvent.change(fileInput, { target: { files: [blob] } });

	await waitFor(() => {
		expect(uploadFiles).toHaveBeenCalledWith([blob], undefined);
	});

	expect(mockEnqueueSnackbar).toHaveBeenCalledWith('ERROR_UPLOAD_FAILED', { variant: 'error' });
	expect(fileInput.value).toBe('');
});
