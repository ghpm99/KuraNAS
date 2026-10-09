import useI18n from '@/components/i18n/provider/i18nContext';
import { Breadcrumbs, useMediaQuery } from '@mui/material';
import type { BreadcrumbSegment } from './useFilesExplorerScreen';
import { viewportMediaQueries } from '@/theme/visualTokens';
import styles from './FilesBreadcrumb.module.css';

const narrowScreenQuery = viewportMediaQueries.belowPhone;
const narrowScreenMaxItems = 3;
const wideScreenMaxItems = 8;
const itemsBeforeCollapse = 1;
const itemsAfterCollapse = 2;

type FilesBreadcrumbProps = {
    segments?: BreadcrumbSegment[];
    onNavigate?: (segment: BreadcrumbSegment) => void;
};

const FilesBreadcrumb = ({ segments = [], onNavigate }: FilesBreadcrumbProps) => {
    const { t } = useI18n();
    const isNarrowScreen = useMediaQuery(narrowScreenQuery);

    return (
        <Breadcrumbs
            className={styles.breadcrumb}
            aria-label={t('FILES_CURRENT_LOCATION')}
            expandText={t('FILES_BREADCRUMB_EXPAND')}
            maxItems={isNarrowScreen ? narrowScreenMaxItems : wideScreenMaxItems}
            itemsBeforeCollapse={itemsBeforeCollapse}
            itemsAfterCollapse={itemsAfterCollapse}
            separator="/"
        >
            {segments.map((segment) =>
                segment.isCurrent ? (
                    <span
                        key={`${segment.id ?? 'root'}-${segment.label}`}
                        className={styles.segmentCurrent}
                        aria-current="page"
                        title={segment.label}
                    >
                        {segment.label}
                    </span>
                ) : (
                    <button
                        key={`${segment.id ?? 'root'}-${segment.label}`}
                        type="button"
                        className={styles.segmentButton}
                        title={segment.label}
                        onClick={() => onNavigate?.(segment)}
                    >
                        {segment.label}
                    </button>
                )
            )}
        </Breadcrumbs>
    );
};

export default FilesBreadcrumb;
