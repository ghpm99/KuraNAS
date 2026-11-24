import { Button, IconButton } from '@mui/material';
import { ArrowLeft, ChevronLeft, ChevronRight, Download, FolderOpen, Info } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import { getFileDownloadUrl } from '@/service/files';
import DocumentStarButton from '../DocumentStarButton';

type DocumentViewerToolbarProps = {
    fileId?: number;
    fileName?: string;
    isStarred: boolean;
    onBack: () => void;
    onPrevious?: () => void;
    onNext?: () => void;
    onToggleStar: () => void;
    onOpenFolder: () => void;
    onOpenDetails: () => void;
};

const DocumentViewerToolbar = ({
    fileId,
    fileName = '',
    isStarred,
    onBack,
    onPrevious,
    onNext,
    onToggleStar,
    onOpenFolder,
    onOpenDetails,
}: DocumentViewerToolbarProps) => {
    const { t } = useI18n();
    const hasFile = fileId !== undefined;

    return (
        <>
            <Button
                size="small"
                variant="outlined"
                startIcon={<ArrowLeft size={16} />}
                onClick={onBack}
            >
                {t('DOCUMENTS_VIEWER_BACK')}
            </Button>
            <IconButton
                size="small"
                aria-label={t('DOCUMENTS_VIEWER_PREVIOUS')}
                disabled={!onPrevious}
                onClick={onPrevious}
            >
                <ChevronLeft size={18} />
            </IconButton>
            <IconButton
                size="small"
                aria-label={t('DOCUMENTS_VIEWER_NEXT')}
                disabled={!onNext}
                onClick={onNext}
            >
                <ChevronRight size={18} />
            </IconButton>
            {hasFile ? (
                <>
                    <DocumentStarButton
                        documentName={fileName}
                        isStarred={isStarred}
                        onToggle={onToggleStar}
                    />
                    <Button
                        size="small"
                        variant="outlined"
                        startIcon={<Download size={16} />}
                        component="a"
                        href={getFileDownloadUrl(fileId)}
                        download
                    >
                        {t('DOCUMENTS_VIEWER_DOWNLOAD')}
                    </Button>
                    <Button
                        size="small"
                        variant="outlined"
                        startIcon={<FolderOpen size={16} />}
                        onClick={onOpenFolder}
                    >
                        {t('DOCUMENTS_VIEWER_OPEN_FOLDER')}
                    </Button>
                    <Button
                        size="small"
                        variant="outlined"
                        startIcon={<Info size={16} />}
                        onClick={onOpenDetails}
                    >
                        {t('DOCUMENTS_VIEWER_DETAILS')}
                    </Button>
                </>
            ) : null}
        </>
    );
};

export default DocumentViewerToolbar;
