import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { ViewerContent } from './useImageViewerModal';
import styles from './ImageViewerModal.module.css';

type ImageViewerContentSectionProps = {
    content?: Partial<ViewerContent>;
    onSearchTag: (tag: string) => void;
};

export default function ImageViewerContentSection({
    content,
    onSearchTag,
}: ImageViewerContentSectionProps) {
    const { t } = useI18n();
    const [isOcrExpanded, setIsOcrExpanded] = useState(false);
    const caption = content?.caption ?? '';
    const tags = content?.tags ?? [];
    const ocrText = content?.ocrText ?? '';

    if (!caption && tags.length === 0 && !ocrText) {
        return null;
    }

    return (
        <section className={styles.detailsSection}>
            <h4>{t('IMAGES_DETAILS_SECTION_CONTENT')}</h4>
            <div className={styles.detailsList}>
                {caption ? (
                    <div className={styles.detailsItem}>
                        <span className={styles.detailsLabel}>{t('IMAGES_DETAIL_CAPTION')}</span>
                        <span className={styles.detailsValue}>{caption}</span>
                    </div>
                ) : null}
                {tags.length > 0 ? (
                    <div className={styles.detailsItem}>
                        <span className={styles.detailsLabel}>{t('IMAGES_DETAIL_TAGS')}</span>
                        <div className={styles.tagList}>
                            {tags.map((tag) => (
                                <button
                                    key={tag}
                                    type="button"
                                    className={styles.tagChip}
                                    onClick={() => onSearchTag(tag)}
                                    aria-label={t('IMAGES_DETAIL_TAG_SEARCH', { tag })}
                                >
                                    {tag}
                                </button>
                            ))}
                        </div>
                    </div>
                ) : null}
                {ocrText ? (
                    <div className={styles.detailsItem}>
                        <button
                            type="button"
                            className={styles.detailsCopyButton}
                            onClick={() => setIsOcrExpanded((isExpanded) => !isExpanded)}
                            aria-expanded={isOcrExpanded}
                        >
                            {isOcrExpanded
                                ? t('IMAGES_DETAIL_OCR_HIDE')
                                : t('IMAGES_DETAIL_OCR_SHOW')}
                        </button>
                        {isOcrExpanded ? (
                            <span className={styles.detailsValue}>{ocrText}</span>
                        ) : null}
                    </div>
                ) : null}
            </div>
        </section>
    );
}
