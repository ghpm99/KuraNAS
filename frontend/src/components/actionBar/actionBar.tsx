import { ArrowLeft, Copy, FolderPlus, MoveRight, Pencil, RefreshCcw, Trash2, Upload } from 'lucide-react';
import useI18n from '../i18n/provider/i18nContext';
import useFile from '../providers/fileProvider/fileContext';
import { FileType } from '@/utils';
import {
	Box,
	Button,
	Dialog,
	DialogActions,
	DialogContent,
	DialogTitle,
	IconButton,
	TextField,
	Typography,
} from '@mui/material';
import { Link } from 'react-router-dom';
import { useRef, useState, type ChangeEvent } from 'react';
import { useSnackbar } from 'notistack';
import { apiBase } from '@/service';

export const ActionBar = () => {
	const { selectedItem, uploadFiles, createFolder, movePath, copyPath, renamePath, deletePath, rescanFiles } =
		useFile();
	const { t } = useI18n();
	const { enqueueSnackbar } = useSnackbar();
	const uploadInputRef = useRef<HTMLInputElement | null>(null);
	const [createFolderOpen, setCreateFolderOpen] = useState(false);
	const [moveOpen, setMoveOpen] = useState(false);
	const [copyOpen, setCopyOpen] = useState(false);
	const [renameOpen, setRenameOpen] = useState(false);
	const [deleteOpen, setDeleteOpen] = useState(false);
	const [folderName, setFolderName] = useState('');
	const [moveTargetDir, setMoveTargetDir] = useState('');
	const [copyDestinationPath, setCopyDestinationPath] = useState('');
	const [renameName, setRenameName] = useState('');

	const currentDirectoryPath = selectedItem
		? selectedItem.type === FileType.Directory
			? selectedItem.path
			: selectedItem.parent_path
		: undefined;

	const handleUploadClick = () => uploadInputRef.current?.click();

	const handleUploadChange = async (event: ChangeEvent<HTMLInputElement>) => {
		const inputFiles = event.target.files;
		if (!inputFiles || inputFiles.length === 0) return;
		try {
			await uploadFiles(inputFiles, currentDirectoryPath);
			enqueueSnackbar(t('ACTION_UPLOAD_SUCCESS'), { variant: 'success' });
		} catch {
			enqueueSnackbar(t('ERROR_UPLOAD_FAILED'), { variant: 'error' });
		} finally {
			event.target.value = '';
		}
	};

	const handleCreateFolder = async () => {
		if (folderName.trim() === '') return;
		try {
			await createFolder(folderName.trim(), currentDirectoryPath);
			enqueueSnackbar(t('ACTION_CREATE_FOLDER_SUCCESS'), { variant: 'success' });
			setCreateFolderOpen(false);
			setFolderName('');
		} catch {
			enqueueSnackbar(t('ERROR_CREATE_FOLDER_FAILED'), { variant: 'error' });
		}
	};

	const handleMoveSelected = async () => {
		if (!selectedItem) return;
		if (moveTargetDir.trim() === '') return;
		const destinationPath = `${moveTargetDir.trim().replace(/[\\/]+$/, '')}/${selectedItem.name}`;
		try {
			await movePath(selectedItem.path, destinationPath);
			enqueueSnackbar(t('ACTION_MOVE_SUCCESS'), { variant: 'success' });
			setMoveOpen(false);
		} catch {
			enqueueSnackbar(t('ERROR_MOVE_FAILED'), { variant: 'error' });
		}
	};

	const handleDeleteSelected = async () => {
		if (!selectedItem) return;
		try {
			await deletePath(selectedItem.path);
			enqueueSnackbar(t('ACTION_DELETE_SUCCESS'), { variant: 'success' });
			setDeleteOpen(false);
		} catch {
			enqueueSnackbar(t('ERROR_DELETE_FAILED'), { variant: 'error' });
		}
	};

	const handleCopySelected = async () => {
		if (!selectedItem) return;
		if (copyDestinationPath.trim() === '') return;
		try {
			await copyPath(selectedItem.path, copyDestinationPath.trim());
			enqueueSnackbar(t('ACTION_COPY_SUCCESS'), { variant: 'success' });
			setCopyOpen(false);
		} catch {
			enqueueSnackbar(t('ERROR_COPY_FAILED'), { variant: 'error' });
		}
	};

	const handleRenameSelected = async () => {
		if (!selectedItem) return;
		if (renameName.trim() === '' || renameName.trim() === selectedItem.name) return;
		try {
			await renamePath(selectedItem.path, renameName.trim());
			enqueueSnackbar(t('ACTION_RENAME_SUCCESS'), { variant: 'success' });
			setRenameOpen(false);
		} catch {
			enqueueSnackbar(t('ERROR_RENAME_FAILED'), { variant: 'error' });
		}
	};

	const openCreateFolderDialog = () => {
		setFolderName('');
		setCreateFolderOpen(true);
	};

	const openMoveDialog = () => {
		if (!selectedItem) return;
		setMoveTargetDir(selectedItem.parent_path);
		setMoveOpen(true);
	};

	const openCopyDialog = () => {
		if (!selectedItem) return;
		const defaultTarget = `${selectedItem.parent_path.replace(/[\\/]+$/, '')}/${selectedItem.name}${t('COPY_SUFFIX')}`;
		setCopyDestinationPath(defaultTarget);
		setCopyOpen(true);
	};

	const openRenameDialog = () => {
		if (!selectedItem) return;
		setRenameName(selectedItem.name);
		setRenameOpen(true);
	};

	const handleDownloadSelected = async () => {
		if (!selectedItem || selectedItem.type !== FileType.File) return;
		try {
			const response = await apiBase.get(`/files/blob/${selectedItem.id}`, {
				responseType: 'blob',
			});
			const blobUrl = URL.createObjectURL(response.data);
			const link = document.createElement('a');
			link.href = blobUrl;
			link.download = selectedItem.name;
			document.body.appendChild(link);
			link.click();
			link.remove();
			URL.revokeObjectURL(blobUrl);
		} catch {
			enqueueSnackbar(t('ERROR_LOADING_FILES'), { variant: 'error' });
		}
	};

	return (
		<Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
			<Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
				{selectedItem && (
					<IconButton component={Link} to='/' size='small'>
						<ArrowLeft size={16} />
					</IconButton>
				)}
				<Typography variant='h6'>{selectedItem?.name ?? t('FILES')}</Typography>
			</Box>
			<Box sx={{ display: 'flex', gap: 1 }}>
				<input ref={uploadInputRef} type='file' multiple style={{ display: 'none' }} onChange={handleUploadChange} />
				<Button variant='contained' size='small' startIcon={<RefreshCcw size={16} />} onClick={rescanFiles}>
					{t('NEW_FILE')}
				</Button>
				<Button variant='outlined' size='small' startIcon={<Upload size={16} />} onClick={handleUploadClick}>
					{t('UPLOAD_FILE')}
				</Button>
				<Button variant='outlined' size='small' startIcon={<FolderPlus size={16} />} onClick={openCreateFolderDialog}>
					{t('NEW_FOLDER')}
				</Button>
				{selectedItem && (
					<Button variant='outlined' size='small' startIcon={<MoveRight size={16} />} onClick={openMoveDialog}>
						{t('MOVE')}
					</Button>
				)}
				{selectedItem && (
					<Button variant='outlined' size='small' startIcon={<Copy size={16} />} onClick={openCopyDialog}>
						{t('COPY')}
					</Button>
				)}
				{selectedItem && (
					<Button variant='outlined' size='small' startIcon={<Pencil size={16} />} onClick={openRenameDialog}>
						{t('RENAME')}
					</Button>
				)}
				{selectedItem && (
					<Button
						color='error'
						variant='outlined'
						size='small'
						startIcon={<Trash2 size={16} />}
						onClick={() => setDeleteOpen(true)}
					>
						{t('DELETE')}
					</Button>
				)}
				{selectedItem?.type === FileType.File && (
					<Button variant='outlined' size='small' onClick={handleDownloadSelected}>
						{t('DOWNLOAD')}
					</Button>
				)}
			</Box>
			<Dialog open={createFolderOpen} onClose={() => setCreateFolderOpen(false)} maxWidth='sm' fullWidth>
				<DialogTitle>{t('NEW_FOLDER')}</DialogTitle>
				<DialogContent>
					<TextField
						autoFocus
						margin='dense'
						label={t('NAME')}
						fullWidth
						value={folderName}
						onChange={(event) => setFolderName(event.target.value)}
					/>
				</DialogContent>
				<DialogActions>
					<Button onClick={() => setCreateFolderOpen(false)}>{t('ACTION_CANCEL')}</Button>
					<Button onClick={handleCreateFolder} variant='contained' disabled={folderName.trim() === ''}>
						{t('NEW_FOLDER')}
					</Button>
				</DialogActions>
			</Dialog>
			<Dialog open={moveOpen} onClose={() => setMoveOpen(false)} maxWidth='md' fullWidth>
				<DialogTitle>{t('MOVE')}</DialogTitle>
				<DialogContent>
					<TextField
						autoFocus
						margin='dense'
						label={t('PATH')}
						fullWidth
						value={moveTargetDir}
						onChange={(event) => setMoveTargetDir(event.target.value)}
					/>
				</DialogContent>
				<DialogActions>
					<Button onClick={() => setMoveOpen(false)}>{t('ACTION_CANCEL')}</Button>
					<Button onClick={handleMoveSelected} variant='contained' disabled={moveTargetDir.trim() === ''}>
						{t('MOVE')}
					</Button>
				</DialogActions>
			</Dialog>
			<Dialog open={copyOpen} onClose={() => setCopyOpen(false)} maxWidth='md' fullWidth>
				<DialogTitle>{t('COPY')}</DialogTitle>
				<DialogContent>
					<TextField
						autoFocus
						margin='dense'
						label={t('PATH')}
						fullWidth
						value={copyDestinationPath}
						onChange={(event) => setCopyDestinationPath(event.target.value)}
					/>
				</DialogContent>
				<DialogActions>
					<Button onClick={() => setCopyOpen(false)}>{t('ACTION_CANCEL')}</Button>
					<Button onClick={handleCopySelected} variant='contained' disabled={copyDestinationPath.trim() === ''}>
						{t('COPY')}
					</Button>
				</DialogActions>
			</Dialog>
			<Dialog open={renameOpen} onClose={() => setRenameOpen(false)} maxWidth='sm' fullWidth>
				<DialogTitle>{t('RENAME')}</DialogTitle>
				<DialogContent>
					<TextField
						autoFocus
						margin='dense'
						label={t('NAME')}
						fullWidth
						value={renameName}
						onChange={(event) => setRenameName(event.target.value)}
					/>
				</DialogContent>
				<DialogActions>
					<Button onClick={() => setRenameOpen(false)}>{t('ACTION_CANCEL')}</Button>
					<Button
						onClick={handleRenameSelected}
						variant='contained'
						disabled={renameName.trim() === '' || renameName.trim() === selectedItem?.name}
					>
						{t('RENAME')}
					</Button>
				</DialogActions>
			</Dialog>
			<Dialog open={deleteOpen} onClose={() => setDeleteOpen(false)} maxWidth='xs' fullWidth>
				<DialogTitle>{t('DELETE')}</DialogTitle>
				<DialogContent>
					<Typography variant='body2'>{t('CONFIRM_DELETE')}</Typography>
				</DialogContent>
				<DialogActions>
					<Button onClick={() => setDeleteOpen(false)}>{t('ACTION_CANCEL')}</Button>
					<Button onClick={handleDeleteSelected} variant='contained' color='error'>
						{t('DELETE')}
					</Button>
				</DialogActions>
			</Dialog>
		</Box>
	);
};

export default ActionBar;
