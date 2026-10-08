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
import useFileActionFlow from '@/features/files/fileActions/useFileActionFlow';
import useFileOperations from '@/features/files/fileActions/useFileOperations';

export const ActionBar = () => {
    const {
        selectedItem,
        uploadFiles,
        createFolder,
        rescanFiles,
        fileListFilter,
    } = useFile();
    const { t } = useI18n();
    const navigate = useNavigate();
    const { enqueueSnackbar } = useSnackbar();
    const uploadInputRef = useRef<HTMLInputElement | null>(null);
    const { startAction, dialogs } = useFileActionFlow();
    const { downloadFiles } = useFileOperations();
    const [isCreateFolderOpen, setIsCreateFolderOpen] = useState(false);
    const [folderName, setFolderName] = useState('');
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
            setIsCreateFolderOpen(false);
            setFolderName('');
        } catch {
            enqueueSnackbar(t('ERROR_CREATE_FOLDER_FAILED'), { variant: 'error' });
        }
    };

    const openCreateFolderDialog = () => {
        setFolderName('');
        setIsCreateFolderOpen(true);
    };

    const startActionOnOpenedItem = (action: 'move' | 'copy' | 'rename' | 'delete') => {
        if (!selectedItem) return;
        startAction(action, [selectedItem]);
    };

    const handleDownloadSelected = () => {
        if (!selectedItem) return;
        downloadFiles([selectedItem]);
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
                        onClick={() => startActionOnOpenedItem('move')}
                    >
                        {t('MOVE')}
                    </Button>
                )}
                {selectedItem && (
                    <Button
                        variant="outlined"
                        size="small"
                        startIcon={<Copy size={16} />}
                        onClick={() => startActionOnOpenedItem('copy')}
                    >
                        {t('COPY')}
                    </Button>
                )}
                {selectedItem && (
                    <Button
                        variant="outlined"
                        size="small"
                        startIcon={<Pencil size={16} />}
                        onClick={() => startActionOnOpenedItem('rename')}
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
                        onClick={() => startActionOnOpenedItem('delete')}
                    >
                        {t('DELETE')}
                    </Button>
                )}
                {selectedItem && (
                    <Button variant="outlined" size="small" onClick={handleDownloadSelected}>
                        {t('DOWNLOAD')}
                    </Button>
                )}
            </Box>
            <Dialog
                open={isCreateFolderOpen}
                onClose={() => setIsCreateFolderOpen(false)}
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
                    <Button onClick={() => setIsCreateFolderOpen(false)}>{t('ACTION_CANCEL')}</Button>
                    <Button
                        onClick={handleCreateFolder}
                        variant="contained"
                        disabled={folderName.trim() === ''}
                    >
                        {t('NEW_FOLDER')}
                    </Button>
                </DialogActions>
            </Dialog>
            {dialogs}
        </Box>
    );
};

export default ActionBar;
