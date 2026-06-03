import { appRoutes } from '@/app/routes';
import {
    ArrowLeft,
    Copy,
    FolderPlus,
    MoveRight,
    Pencil,
    RefreshCcw,
    Trash2,
    Upload,
} from 'lucide-react';
import useI18n from '../i18n/provider/i18nContext';
import useFile from '@/features/files/providers/fileProvider/fileContext';
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
import { useNavigate } from 'react-router-dom';
import { useRef, useState, type ChangeEvent } from 'react';
import { useSnackbar } from 'notistack';
import { downloadFileBlob } from '@/service/files';
import FolderPicker, { type FolderPickerResult } from '@/components/folderPicker/folderPicker';

export const ActionBar = () => {
    const {
        selectedItem,
        uploadFiles,
        createFolder,
        moveFile,
        copyFile,
        renameFile,
        deleteFile,
        rescanFiles,
        fileListFilter,
    } = useFile();
    const { t } = useI18n();
    const navigate = useNavigate();
    const { enqueueSnackbar } = useSnackbar();
    const uploadInputRef = useRef<HTMLInputElement | null>(null);
    const [openDialog, setOpenDialog] = useState<
        'createFolder' | 'move' | 'copy' | 'rename' | 'delete' | null
    >(null);
    const [folderName, setFolderName] = useState('');
    const [renameName, setRenameName] = useState('');
    const currentListTitle =
        fileListFilter === 'starred'
            ? t('STARRED_FILES')
            : fileListFilter === 'recent'
              ? t('RECENT_FILES')
              : t('FILES');

    const currentFolderId =
        selectedItem && selectedItem.type === FileType.Directory
            ? selectedItem.id
            : undefined;

    const handleUploadClick = () => uploadInputRef.current?.click();

    const handleUploadChange = async (event: ChangeEvent<HTMLInputElement>) => {
        const inputFiles = event.target.files;
        if (!inputFiles || inputFiles.length === 0) return;
        try {
            await uploadFiles(inputFiles, currentFolderId);
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
            await createFolder(folderName.trim(), currentFolderId);
            enqueueSnackbar(t('ACTION_CREATE_FOLDER_SUCCESS'), {
                variant: 'success',
            });
            setOpenDialog(null);
            setFolderName('');
        } catch {
            enqueueSnackbar(t('ERROR_CREATE_FOLDER_FAILED'), { variant: 'error' });
        }
    };

    const handleMoveSelected = async (result: FolderPickerResult) => {
        if (!selectedItem) return;
        try {
            await moveFile(selectedItem.id, result.folderId, result.path);
            enqueueSnackbar(t('ACTION_MOVE_SUCCESS'), { variant: 'success' });
            setOpenDialog(null);
        } catch {
            enqueueSnackbar(t('ERROR_MOVE_FAILED'), { variant: 'error' });
        }
    };

    const handleDeleteSelected = async () => {
        if (!selectedItem) return;
        try {
            await deleteFile(selectedItem.id);
            enqueueSnackbar(t('ACTION_DELETE_SUCCESS'), { variant: 'success' });
            setOpenDialog(null);
        } catch {
            enqueueSnackbar(t('ERROR_DELETE_FAILED'), { variant: 'error' });
        }
    };

    const handleCopySelected = async (result: FolderPickerResult) => {
        if (!selectedItem) return;
        try {
            await copyFile(selectedItem.id, result.folderId, result.path);
            enqueueSnackbar(t('ACTION_COPY_SUCCESS'), { variant: 'success' });
            setOpenDialog(null);
        } catch {
            enqueueSnackbar(t('ERROR_COPY_FAILED'), { variant: 'error' });
        }
    };

    const handleRenameSelected = async () => {
        if (!selectedItem) return;
        if (renameName.trim() === '' || renameName.trim() === selectedItem.name) return;
        try {
            await renameFile(selectedItem.id, renameName.trim());
            enqueueSnackbar(t('ACTION_RENAME_SUCCESS'), { variant: 'success' });
            setOpenDialog(null);
        } catch {
            enqueueSnackbar(t('ERROR_RENAME_FAILED'), { variant: 'error' });
        }
    };

    const openCreateFolderDialog = () => {
        setFolderName('');
        setOpenDialog('createFolder');
    };

    const openMoveDialog = () => {
        if (!selectedItem) return;
        setOpenDialog('move');
    };

    const openCopyDialog = () => {
        if (!selectedItem) return;
        setOpenDialog('copy');
    };

    const openRenameDialog = () => {
        if (!selectedItem) return;
        setRenameName(selectedItem.name);
        setOpenDialog('rename');
    };

    const handleDownloadSelected = async () => {
        if (!selectedItem || selectedItem.type !== FileType.File) return;
        try {
            const fileBlob = await downloadFileBlob(selectedItem.id);
            const blobUrl = URL.createObjectURL(fileBlob);
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
        <Box
            sx={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                mb: 2,
            }}
        >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                {selectedItem && (
                    <IconButton
                        size="small"
                        onClick={() => {
                            const parentPath = selectedItem.parent_path;
                            const url =
                                parentPath && parentPath !== '/'
                                    ? `${appRoutes.files}${parentPath}`
                                    : appRoutes.files;
                            navigate(url);
                        }}
                    >
                        <ArrowLeft size={16} />
                    </IconButton>
                )}
                <Typography variant="h6">{selectedItem?.name ?? currentListTitle}</Typography>
            </Box>
            <Box sx={{ display: 'flex', gap: 1 }}>
                <input
                    ref={uploadInputRef}
                    type="file"
                    multiple
                    style={{ display: 'none' }}
                    onChange={handleUploadChange}
                />
                <Button
                    variant="contained"
                    size="small"
                    startIcon={<RefreshCcw size={16} />}
                    onClick={rescanFiles}
                >
                    {t('NEW_FILE')}
                </Button>
                <Button
                    variant="outlined"
                    size="small"
                    startIcon={<Upload size={16} />}
                    onClick={handleUploadClick}
                >
                    {t('UPLOAD_FILE')}
                </Button>
                <Button
                    variant="outlined"
                    size="small"
                    startIcon={<FolderPlus size={16} />}
                    onClick={openCreateFolderDialog}
                >
                    {t('NEW_FOLDER')}
                </Button>
                {selectedItem && (
                    <Button
                        variant="outlined"
                        size="small"
                        startIcon={<MoveRight size={16} />}
                        onClick={openMoveDialog}
                    >
                        {t('MOVE')}
                    </Button>
                )}
                {selectedItem && (
                    <Button
                        variant="outlined"
                        size="small"
                        startIcon={<Copy size={16} />}
                        onClick={openCopyDialog}
                    >
                        {t('COPY')}
                    </Button>
                )}
                {selectedItem && (
                    <Button
                        variant="outlined"
                        size="small"
                        startIcon={<Pencil size={16} />}
                        onClick={openRenameDialog}
                    >
                        {t('RENAME')}
                    </Button>
                )}
                {selectedItem && (
                    <Button
                        color="error"
                        variant="outlined"
                        size="small"
                        startIcon={<Trash2 size={16} />}
                        onClick={() => setOpenDialog('delete')}
                    >
                        {t('DELETE')}
                    </Button>
                )}
                {selectedItem?.type === FileType.File && (
                    <Button variant="outlined" size="small" onClick={handleDownloadSelected}>
                        {t('DOWNLOAD')}
                    </Button>
                )}
            </Box>
            <Dialog
                open={openDialog === 'createFolder'}
                onClose={() => setOpenDialog(null)}
                maxWidth="sm"
                fullWidth
            >
                <DialogTitle>{t('NEW_FOLDER')}</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        margin="dense"
                        label={t('NAME')}
                        fullWidth
                        value={folderName}
                        onChange={(event) => setFolderName(event.target.value)}
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setOpenDialog(null)}>{t('ACTION_CANCEL')}</Button>
                    <Button
                        onClick={handleCreateFolder}
                        variant="contained"
                        disabled={folderName.trim() === ''}
                    >
                        {t('NEW_FOLDER')}
                    </Button>
                </DialogActions>
            </Dialog>
            <FolderPicker
                open={openDialog === 'move'}
                onClose={() => setOpenDialog(null)}
                onSelect={handleMoveSelected}
            />
            <FolderPicker
                open={openDialog === 'copy'}
                onClose={() => setOpenDialog(null)}
                onSelect={handleCopySelected}
            />
            <Dialog
                open={openDialog === 'rename'}
                onClose={() => setOpenDialog(null)}
                maxWidth="sm"
                fullWidth
            >
                <DialogTitle>{t('RENAME')}</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        margin="dense"
                        label={t('NAME')}
                        fullWidth
                        value={renameName}
                        onChange={(event) => setRenameName(event.target.value)}
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setOpenDialog(null)}>{t('ACTION_CANCEL')}</Button>
                    <Button
                        onClick={handleRenameSelected}
                        variant="contained"
                        disabled={
                            renameName.trim() === '' || renameName.trim() === selectedItem?.name
                        }
                    >
                        {t('RENAME')}
                    </Button>
                </DialogActions>
            </Dialog>
            <Dialog
                open={openDialog === 'delete'}
                onClose={() => setOpenDialog(null)}
                maxWidth="xs"
                fullWidth
            >
                <DialogTitle>{t('DELETE')}</DialogTitle>
                <DialogContent>
                    <Typography variant="body2">{t('CONFIRM_DELETE')}</Typography>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setOpenDialog(null)}>{t('ACTION_CANCEL')}</Button>
                    <Button onClick={handleDeleteSelected} variant="contained" color="error">
                        {t('DELETE')}
                    </Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default ActionBar;
