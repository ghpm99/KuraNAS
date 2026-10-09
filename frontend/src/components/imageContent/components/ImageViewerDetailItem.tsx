import { Check, Copy } from 'lucide-react';
import { useEffect, useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { ViewerDetailItem } from './useImageViewerModal';
import styles from './ImageViewerModal.module.css';

const copiedFeedbackDurationInMs = 2000;

type ImageViewerDetailItemProps = {
    item: ViewerDetailItem;
};

export default function ImageViewerDetailItem({ item }: ImageViewerDetailItemProps) {
    const { t } = useI18n();
    const [isCopied, setIsCopied] = useState(false);

    useEffect(() => {
        if (!isCopied) {
            return;
        }
        const timeoutId = window.setTimeout(() => setIsCopied(false), copiedFeedbackDurationInMs);
        return () => window.clearTimeout(timeoutId);
    }, [isCopied]);

    const copyValue = async () => {
        if (!item.copyValue || !navigator.clipboard) {
            return;
        }
        try {
            await navigator.clipboard.writeText(item.copyValue);
            setIsCopied(true);
        } catch {
            setIsCopied(false);
        }
    };

    return (
        <div className={styles.detailsItem}>
            <span className={styles.detailsLabel}>{item.label}</span>
            <span className={styles.detailsValue}>{item.value}</span>
            {item.link ? (
                <a
                    className={styles.detailsLink}
                    href={item.link.href}
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    {item.link.label}
                </a>
            ) : null}
            {item.copyValue ? (
                <button
                    type="button"
                    className={styles.detailsCopyButton}
                    onClick={copyValue}
                    aria-label={`${t('IMAGES_DETAIL_COPY')}: ${item.label}`}
                >
                    {isCopied ? <Check size={14} /> : <Copy size={14} />}
                    <span>{isCopied ? t('IMAGES_DETAIL_COPIED') : t('IMAGES_DETAIL_COPY')}</span>
                </button>
            ) : null}
        </div>
    );
}
