import { appRoutes } from '@/app/routes';
import {
    ArrowLeft,
    Copy,
    Download,
    FolderPlus,
    FolderUp,
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
import { useRef, useState, type ChangeEvent, type InputHTMLAttributes } from 'react';
import ConflictPolicySelect from '@/features/files/upload/conflictPolicySelect';
import { entriesFromFileList } from '@/features/files/upload/entriesFromFileList';
import useUploadToCurrentFolder from '@/features/files/upload/useUploadToCurrentFolder';
import { useSnackbar } from 'notistack';
import useFileActionFlow from '@/features/files/fileActions/useFileActionFlow';
import useFileOperations from '@/features/files/fileActions/useFileOperations';
import ActionBarButton from './actionBarButton';
import ActionBarMoreMenu, { type ActionBarMenuEntry } from './actionBarMoreMenu';
import useActionBarLayout from './useActionBarLayout';

const folderInputAttributes = {
    webkitdirectory: '',
} as unknown as InputHTMLAttributes<HTMLInputElement>;

export const ActionBar = () => {
    const {
        selectedItem,
        createFolder,
        rescanFiles,
        fileListFilter,
    } = useFile();
    const { t } = useI18n();
    const navigate = useNavigate();
    const { enqueueSnackbar } = useSnackbar();
    const uploadInputRef = useRef<HTMLInputElement | null>(null);
    const uploadFolderInputRef = useRef<HTMLInputElement | null>(null);
    const { uploadEntries } = useUploadToCurrentFolder();
    const { startAction, dialogs } = useFileActionFlow();
    const { downloadFiles } = useFileOperations();
    const { isIconOnly, isSecondaryCollapsed } = useActionBarLayout();
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

    const handleUploadFolderClick = () => uploadFolderInputRef.current?.click();

    const handleUploadChange = async (event: ChangeEvent<HTMLInputElement>) => {
        const inputFiles = event.target.files;
        if (!inputFiles || inputFiles.length === 0) return;
        const entries = entriesFromFileList(inputFiles);
        event.target.value = '';
        await uploadEntries(entries);
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

    const openedItemEntries: ActionBarMenuEntry[] = selectedItem
        ? [
              {
                  key: 'move',
                  label: t('MOVE'),
                  icon: <MoveRight size={16} />,
                  onSelect: () => startActionOnOpenedItem('move'),
              },
              {
                  key: 'copy',
                  label: t('COPY'),
                  icon: <Copy size={16} />,
                  onSelect: () => startActionOnOpenedItem('copy'),
              },
              {
                  key: 'rename',
                  label: t('RENAME'),
                  icon: <Pencil size={16} />,
                  onSelect: () => startActionOnOpenedItem('rename'),
              },
              {
                  key: 'delete',
                  label: t('DELETE'),
                  icon: <Trash2 size={16} />,
                  onSelect: () => startActionOnOpenedItem('delete'),
                  isDestructive: true,
              },
              {
                  key: 'download',
                  label: t('DOWNLOAD'),
                  icon: <Download size={16} />,
                  onSelect: handleDownloadSelected,
              },
          ]
        : [];

    const rescanEntry: ActionBarMenuEntry = {
        key: 'rescan',
        label: t('FILES_RESCAN_FOLDER'),
        icon: <RefreshCcw size={16} />,
        onSelect: rescanFiles,
    };

    const moreMenuEntries = isSecondaryCollapsed
        ? [...openedItemEntries, rescanEntry]
        : [rescanEntry];

    return (
        <Box
            sx={{
                display: 'flex',
                flexWrap: 'wrap',
                justifyContent: 'space-between',
                alignItems: 'center',
                gap: 1,
                mb: 2,
            }}
        >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, minWidth: 0 }}>
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
                <Typography variant="h6" noWrap>{selectedItem?.name ?? currentListTitle}</Typography>
            </Box>
            <Box sx={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: 1, minWidth: 0 }}>
                <input
                    ref={uploadInputRef}
                    type="file"
                    multiple
                    style={{ display: 'none' }}
                    onChange={handleUploadChange}
                />
                <input
                    ref={uploadFolderInputRef}
                    type="file"
                    multiple
                    style={{ display: 'none' }}
                    onChange={handleUploadChange}
                    {...folderInputAttributes}
                />
                <ActionBarButton
                    label={t('UPLOAD_FILE')}
                    icon={<Upload size={16} />}
                    onClick={handleUploadClick}
                    isIconOnly={isIconOnly}
                />
                <ActionBarButton
                    label={t('FILES_UPLOAD_FOLDER')}
                    icon={<FolderUp size={16} />}
                    onClick={handleUploadFolderClick}
                    isIconOnly={isIconOnly}
                />
                <ConflictPolicySelect />
                <ActionBarButton
                    label={t('NEW_FOLDER')}
                    icon={<FolderPlus size={16} />}
                    onClick={openCreateFolderDialog}
                    isIconOnly={isIconOnly}
                />
                {isSecondaryCollapsed
                    ? null
                    : openedItemEntries.map((entry) => (
                          <ActionBarButton
                              key={entry.key}
                              label={entry.label}
                              icon={entry.icon}
                              onClick={entry.onSelect}
                              isIconOnly={isIconOnly}
                              isDestructive={entry.isDestructive}
                          />
                      ))}
                <ActionBarMoreMenu entries={moreMenuEntries} />
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
