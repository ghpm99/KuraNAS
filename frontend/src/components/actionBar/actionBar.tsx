import { appRoutes } from '@/app/routes';
import {
    ArrowLeft,
    Copy,
    Download,
    FolderPlus,
    FolderUp,
    Info,
    MoveRight,
    Pencil,
    RefreshCcw,
    Trash2,
    Upload,
} from 'lucide-react';
import useI18n from '../i18n/provider/i18nContext';
import useFile from '@/features/files/providers/fileProvider/fileContext';
import { Box, IconButton, Typography } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import ConflictPolicySelect from '@/features/files/upload/conflictPolicySelect';
import useUploadPickers from '@/features/files/upload/useUploadPickers';
import useFileDetails from '@/features/files/fileDetails/useFileDetails';
import useFileActionFlow from '@/features/files/fileActions/useFileActionFlow';
import useFileOperations from '@/features/files/fileActions/useFileOperations';
import useCreateFolderFlow from '@/features/files/fileActions/useCreateFolderFlow';
import ActionBarButton from './actionBarButton';
import ActionBarMoreMenu, { type ActionBarMenuEntry } from './actionBarMoreMenu';
import useActionBarLayout from './useActionBarLayout';

export const ActionBar = () => {
    const { selectedItem, rescanFiles, fileListFilter } = useFile();
    const { t } = useI18n();
    const navigate = useNavigate();
    const { openFilePicker, openFolderPicker, pickerInputs } = useUploadPickers();
    const { openCreateFolderDialog, createFolderDialog } = useCreateFolderFlow();
    const { startAction, dialogs } = useFileActionFlow();
    const { downloadFiles } = useFileOperations();
    const { isAvailable: isDetailsAvailable, openDetails } = useFileDetails();
    const { isIconOnly, isSecondaryCollapsed } = useActionBarLayout();
    const currentListTitle =
        fileListFilter === 'starred'
            ? t('STARRED_FILES')
            : fileListFilter === 'recent'
              ? t('RECENT_FILES')
              : t('FILES');

    const startActionOnOpenedItem = (action: 'move' | 'copy' | 'rename' | 'delete') => {
        if (!selectedItem) return;
        startAction(action, [selectedItem]);
    };

    const handleDownloadSelected = () => {
        if (!selectedItem) return;
        downloadFiles([selectedItem]);
    };

    const detailsEntries: ActionBarMenuEntry[] =
        selectedItem && isDetailsAvailable
            ? [
                  {
                      key: 'details',
                      label: t('FILES_DETAILS'),
                      icon: <Info size={16} />,
                      onSelect: () => openDetails(selectedItem),
                  },
              ]
            : [];

    const openedItemEntries: ActionBarMenuEntry[] = selectedItem
        ? [
              ...detailsEntries,
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
                {pickerInputs}
                <ActionBarButton
                    label={t('UPLOAD_FILE')}
                    icon={<Upload size={16} />}
                    onClick={openFilePicker}
                    isIconOnly={isIconOnly}
                />
                <ActionBarButton
                    label={t('FILES_UPLOAD_FOLDER')}
                    icon={<FolderUp size={16} />}
                    onClick={openFolderPicker}
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
            {createFolderDialog}
            {dialogs}
        </Box>
    );
};

export default ActionBar;
