import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import useI18n from '@/components/i18n/provider/i18nContext';
import { invalidateMusicMetadataViews } from '@/features/music/queryKeys/invalidateMusicMetadataViews';
import { isGenreNameRefused, renameMusicGenre } from '@/service/music';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import type { MusicGenre } from '@/types/music';

export type RenameGenreStep = 'editing' | 'confirming';

export interface RenameGenreConfirmation {
    isMerge: boolean;
    movedTrackCount: number;
    resultingTrackCount: number;
    targetName: string;
}

interface UseRenameGenreOptions {
    genreKey: string;
    genreName: string;
    trackCount: number;
    onRenamed: (renamedGenreKey: string) => void;
    onClose: () => void;
}

export default function useRenameGenre({
    genreKey,
    genreName,
    trackCount,
    onRenamed,
    onClose,
}: UseRenameGenreOptions) {
    const { t } = useI18n();
    const { enqueueSnackbar } = useSnackbar();
    const queryClient = useQueryClient();
    const [newName, setNewName] = useState(genreName);
    const [hasTypedName, setHasTypedName] = useState(false);
    const [pickedGenre, setPickedGenre] = useState<MusicGenre | null>(null);
    const [step, setStep] = useState<RenameGenreStep>('editing');
    const [nameError, setNameError] = useState<string | undefined>(undefined);
    const [topError, setTopError] = useState<string | undefined>(undefined);

    const renameMutation = useMutation({
        mutationFn: () => renameMusicGenre(genreKey, newName),
        onMutate: () => {
            setNameError(undefined);
            setTopError(undefined);
        },
        onSuccess: (renamedGenre) => {
            const successMessage = renamedGenre.merged
                ? t('MUSIC_GENRE_RENAME_MERGED_DONE', {
                      count: String(renamedGenre.trackCount),
                      to: renamedGenre.name,
                  })
                : t('MUSIC_GENRE_RENAME_DONE');
            enqueueSnackbar(successMessage, { variant: 'success' });
            onRenamed(renamedGenre.key);
            onClose();
            void invalidateMusicMetadataViews(queryClient);
        },
        onError: (error) => {
            const message = extractBackendErrorMessage(error) ?? t('MUSIC_GENRE_RENAME_FAILED');
            if (isGenreNameRefused(error)) {
                setNameError(message);
                setStep('editing');
                return;
            }
            setTopError(message);
        },
    });

    const typeNewName = (typedName: string) => {
        setNewName(typedName);
        setHasTypedName(true);
        setPickedGenre(null);
        setNameError(undefined);
    };

    const pickExistingGenre = (existingGenre: MusicGenre) => {
        setNewName(existingGenre.genre);
        setPickedGenre(existingGenre);
        setNameError(undefined);
    };

    const isMergeTarget = pickedGenre !== null && pickedGenre.key !== genreKey;
    const confirmation: RenameGenreConfirmation = {
        isMerge: isMergeTarget,
        movedTrackCount: trackCount,
        resultingTrackCount: isMergeTarget ? pickedGenre.track_count + trackCount : trackCount,
        targetName: isMergeTarget ? pickedGenre.genre : newName.trim(),
    };

    const review = () => {
        setTopError(undefined);
        setStep('confirming');
    };

    const backToEditing = () => {
        setTopError(undefined);
        setStep('editing');
    };

    return {
        newName,
        hasTypedName,
        step,
        nameError,
        topError,
        confirmation,
        isSubmitting: renameMutation.isPending,
        canReview: newName.trim() !== '' && !renameMutation.isPending,
        typeNewName,
        pickExistingGenre,
        review,
        backToEditing,
        confirm: () => renameMutation.mutate(),
    };
}
