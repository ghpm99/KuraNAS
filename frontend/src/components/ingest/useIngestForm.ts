import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { FolderPickerResult } from '@/components/folderPicker/folderPicker';
import { fetchFromUrl, getIngestPresets, getIngestTargets, getYtDlpStatus } from '@/service/ingest';
import { getFileLocation } from '@/service/files';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { validateIngestUrl } from './ingestUrlValidation';
import { activeIngestDownloadsQueryKey } from './ingestQueryKeys';
import { subfolderFromDiskPath } from './subfolderFromDiskPath';
import { loadLastIngestChoice, saveLastIngestChoice } from './lastIngestChoicePreference';

export const useIngestForm = () => {
    const { t } = useI18n();
    const queryClient = useQueryClient();
    const [url, setUrl] = useState('');
    const [lastChoice] = useState(loadLastIngestChoice);
    const [selectedTargetRoot, setSelectedTargetRoot] = useState(lastChoice?.targetRoot ?? '');
    const [selectedPreset, setSelectedPreset] = useState(lastChoice?.preset ?? '');
    const [chosenSubfolder, setChosenSubfolder] = useState(lastChoice?.subfolder ?? '');
    const [urlErrorKey, setUrlErrorKey] = useState<string | null>(null);
    const [subfolderError, setSubfolderError] = useState<string | null>(null);
    const [serverError, setServerError] = useState<string | null>(null);
    const [isFolderPickerOpen, setIsFolderPickerOpen] = useState(false);

    const targetsQuery = useQuery({
        queryKey: ['ingest', 'targets'],
        queryFn: getIngestTargets,
        refetchOnWindowFocus: false,
    });
    const presetsQuery = useQuery({
        queryKey: ['ingest', 'presets'],
        queryFn: getIngestPresets,
        refetchOnWindowFocus: false,
    });

    const ytDlpStatusQuery = useQuery({
        queryKey: ['ytdlp-status'],
        queryFn: getYtDlpStatus,
        retry: false,
        refetchOnWindowFocus: false,
    });
    const isYtDlpMissing = ytDlpStatusQuery.data?.installed === false;

    const targets = Array.isArray(targetsQuery.data) ? targetsQuery.data : [];
    const presets = Array.isArray(presetsQuery.data) ? presetsQuery.data : [];
    const isSelectedTargetRootAvailable = targets.some(
        (target) => target.path === selectedTargetRoot
    );
    const isSelectedPresetAvailable = presets.some(
        (availablePreset) => availablePreset.key === selectedPreset
    );
    const targetRoot = isSelectedTargetRootAvailable ? selectedTargetRoot : targets[0]?.path || '';
    const preset = isSelectedPresetAvailable ? selectedPreset : presets[0]?.key || '';
    const subfolder = isSelectedTargetRootAvailable ? chosenSubfolder : '';

    const submitMutation = useMutation({
        mutationFn: fetchFromUrl,
        onSuccess: (_response, sentRequest) => {
            saveLastIngestChoice({
                targetRoot: sentRequest.target_root,
                subfolder: sentRequest.subfolder,
                preset: sentRequest.preset,
            });
            setUrl('');
            setServerError(null);
            void queryClient.invalidateQueries({ queryKey: activeIngestDownloadsQueryKey });
        },
        onError: (error) => {
            setServerError(extractBackendErrorMessage(error) ?? t('INGEST_FORM_SUBMIT_ERROR'));
        },
    });

    const changeUrl = (value: string) => {
        setUrl(value);
        setUrlErrorKey(null);
        setServerError(null);
    };

    const pasteUrlFromClipboard = async () => {
        try {
            changeUrl((await navigator.clipboard.readText()).trim());
        } catch {
            setUrlErrorKey(null);
        }
    };

    const isClipboardAvailable =
        typeof navigator !== 'undefined' && typeof navigator.clipboard?.readText === 'function';

    const changeTargetRoot = (value: string) => {
        setSelectedTargetRoot(value);
        setChosenSubfolder('');
        setSubfolderError(null);
    };

    const selectFolder = async ({ folderId, path }: FolderPickerResult) => {
        setIsFolderPickerOpen(false);
        setSubfolderError(null);
        try {
            const folderDiskPath = folderId ? (await getFileLocation(folderId)).disk_path : path;
            if (folderDiskPath === undefined) {
                setChosenSubfolder('');
                return;
            }
            const relativeSubfolder = subfolderFromDiskPath(targetRoot, folderDiskPath);
            if (relativeSubfolder === null) {
                setSubfolderError(t('INGEST_FORM_SUBFOLDER_OUTSIDE_ROOT'));
                return;
            }
            setChosenSubfolder(relativeSubfolder);
        } catch (error) {
            setSubfolderError(
                extractBackendErrorMessage(error) ?? t('INGEST_FORM_SUBFOLDER_ERROR')
            );
        }
    };

    const submit = () => {
        const validationKey = validateIngestUrl(url);
        setUrlErrorKey(validationKey);
        if (validationKey) return;
        setServerError(null);
        submitMutation.mutate({
            url: url.trim(),
            preset,
            target_root: targetRoot,
            subfolder,
        });
    };

    return {
        t,
        url,
        targets,
        presets,
        targetRoot,
        preset,
        subfolder,
        urlError: urlErrorKey ? t(urlErrorKey) : null,
        subfolderError,
        serverError,
        isSubmitting: submitMutation.isPending,
        isFolderPickerOpen,
        isClipboardAvailable,
        isYtDlpMissing,
        canSubmit:
            targetRoot !== '' && preset !== '' && !isYtDlpMissing && !submitMutation.isPending,
        changeUrl,
        pasteUrlFromClipboard,
        changeTargetRoot,
        changePreset: setSelectedPreset,
        openFolderPicker: () => setIsFolderPickerOpen(true),
        closeFolderPicker: () => setIsFolderPickerOpen(false),
        selectFolder,
        submit,
    };
};
