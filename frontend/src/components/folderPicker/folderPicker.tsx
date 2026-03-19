import { useCallback, useEffect, useState } from 'react';
import {
    Box,
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    List,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    TextField,
    Typography,
} from '@mui/material';
import { Folder } from 'lucide-react';
import { getFilesTree } from '@/service/files';
import { FileData } from '@/components/providers/fileProvider/fileContext';
import { FileType } from '@/utils';
import useI18n from '@/components/i18n/provider/i18nContext';

export type FolderPickerResult = {
    folderId?: number;
    path?: string;
};

type FolderPickerProps = {
    open: boolean;
    onClose: () => void;
    onSelect: (result: FolderPickerResult) => void;
};

type FolderEntry = {
    id: number;
    name: string;
    path: string;
};

const FolderPicker = ({ open, onClose, onSelect }: FolderPickerProps) => {
    const { t } = useI18n();
    const [folders, setFolders] = useState<FolderEntry[]>([]);
    const [loading, setLoading] = useState(false);
    const [breadcrumbs, setBreadcrumbs] = useState<FolderEntry[]>([]);
    const [pathInput, setPathInput] = useState('');
    const [selectedFolder, setSelectedFolder] = useState<FolderEntry | null>(null);

    const fetchFolders = useCallback(async (folderId?: number) => {
        setLoading(true);
        try {
            const response = await getFilesTree({
                page: 1,
                pageSize: 200,
                fileParent: folderId,
                category: 'all',
            });
            const dirs = response.items
                .filter((item: FileData) => item.type === FileType.Directory)
                .map((item: FileData) => ({ id: item.id, name: item.name, path: item.path }));
            setFolders(dirs);
        } catch {
            setFolders([]);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        if (open) {

            setBreadcrumbs([]);
            setPathInput('');
            setSelectedFolder(null);
            fetchFolders(undefined);
        }
    }, [open, fetchFolders]);

    const navigateInto = (folder: FolderEntry) => {

        setBreadcrumbs((prev) => [...prev, folder]);
        setPathInput(folder.path);
        setSelectedFolder(folder);
        fetchFolders(folder.id);
    };

    const navigateToBreadcrumb = (index: number) => {
        if (index < 0) {

            setBreadcrumbs([]);
            setPathInput('');
            setSelectedFolder(null);
            fetchFolders(undefined);
        } else {
            const target = breadcrumbs[index]!;
            setBreadcrumbs((prev) => prev.slice(0, index + 1));
            setPathInput(target.path);
            setSelectedFolder(target);
            fetchFolders(target.id);
        }
    };

    const handleConfirm = () => {
        const trimmedInput = pathInput.trim();
        if (selectedFolder && selectedFolder.path === trimmedInput) {
            onSelect({ folderId: selectedFolder.id });
        } else if (trimmedInput) {
            onSelect({ path: trimmedInput });
        } else {
            onSelect({});
        }
    };

    const handlePathInputChange = (value: string) => {
        setPathInput(value);
        if (selectedFolder && selectedFolder.path !== value.trim()) {
            setSelectedFolder(null);
        }
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>{t('SELECT_DESTINATION')}</DialogTitle>
            <DialogContent>
                <Box sx={{ display: 'flex', gap: 0.5, flexWrap: 'wrap', mb: 1, mt: 1 }}>
                    <Typography
                        variant="body2"
                        sx={{
                            cursor: 'pointer',
                            fontWeight: breadcrumbs.length === 0 ? 'bold' : 'normal',
                            '&:hover': { textDecoration: 'underline' },
                        }}
                        onClick={() => navigateToBreadcrumb(-1)}
                    >
                        {t('ROOT_FOLDER')}
                    </Typography>
                    {breadcrumbs.map((crumb, i) => (
                        <Box key={crumb.id} sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                            <Typography variant="body2" color="text.secondary">
                                /
                            </Typography>
                            <Typography
                                variant="body2"
                                sx={{
                                    cursor: 'pointer',
                                    fontWeight: i === breadcrumbs.length - 1 ? 'bold' : 'normal',
                                    '&:hover': { textDecoration: 'underline' },
                                }}
                                onClick={() => navigateToBreadcrumb(i)}
                            >
                                {crumb.name}
                            </Typography>
                        </Box>
                    ))}
                </Box>
                <Box
                    sx={{
                        minHeight: 200,
                        maxHeight: 300,
                        overflowY: 'auto',
                        border: 1,
                        borderColor: 'divider',
                        borderRadius: 1,
                        mb: 2,
                    }}
                >
                    {loading ? (
                        <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                            <CircularProgress size={24} />
                        </Box>
                    ) : folders.length === 0 ? (
                        <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{ p: 2, textAlign: 'center' }}
                        >
                            {t('EMPTY_FILE_LIST')}
                        </Typography>
                    ) : (
                        <List dense disablePadding>
                            {folders.map((folder) => (
                                <ListItemButton
                                    key={folder.id}
                                    selected={selectedFolder?.id === folder.id}
                                    onClick={() => navigateInto(folder)}
                                    sx={{ borderRadius: 1 }}
                                >
                                    <ListItemIcon sx={{ minWidth: 32 }}>
                                        <Folder size={16} />
                                    </ListItemIcon>
                                    <ListItemText
                                        primary={folder.name}
                                        primaryTypographyProps={{ variant: 'body2' }}
                                    />
                                </ListItemButton>
                            ))}
                        </List>
                    )}
                </Box>
                <TextField
                    fullWidth
                    size="small"
                    label={t('PATH')}
                    value={pathInput}
                    onChange={(e) => handlePathInputChange(e.target.value)}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('ACTION_CANCEL')}</Button>
                <Button onClick={handleConfirm} variant="contained">
                    {t('MOVE')}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default FolderPicker;
