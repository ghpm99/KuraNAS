import useI18n from '@/components/i18n/provider/i18nContext';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import styles from './ImageCollectionsPanel.module.css';

const thumbnailWidth = 960;
const thumbnailHeight = 720;

const thumbnailUrl = (id: number) =>
    `${getApiV1BaseUrl()}/files/thumbnail/${id}?width=${thumbnailWidth}&height=${thumbnailHeight}`;

export type ImageCollectionCard = {
    id: string;
    title: string;
    description: string;
    imageCount: number;
    coverImageId?: number;
};

type ImageCollectionsPanelProps = {
    cards: ImageCollectionCard[];
    emptyTitle: string;
    emptyDescription: string;
    selectedId?: string;
    onSelect: (id: string) => void;
};

const ImageCollectionsPanel = ({
    cards,
    emptyTitle,
    emptyDescription,
    selectedId,
    onSelect,
}: ImageCollectionsPanelProps) => {
    const { t } = useI18n();

    if (cards.length === 0) {
        return (
            <div className={styles.empty}>
                <h3>{emptyTitle}</h3>
                <p>{emptyDescription}</p>
            </div>
        );
    }

    return (
        <div className={styles.grid}>
            {cards.map((card) => (
                <button
                    type="button"
                    key={card.id}
                    onClick={() => onSelect(card.id)}
                    className={
                        card.id === selectedId ? `${styles.card} ${styles.cardActive}` : styles.card
                    }
                    aria-label={t('IMAGES_COLLECTION_OPEN', { name: card.title })}
                >
                    <div className={styles.cover}>
                        {card.coverImageId ? (
                            <img
                                src={thumbnailUrl(card.coverImageId)}
                                alt={card.title}
                                loading="lazy"
                            />
                        ) : (
                            <div className={styles.placeholder}>
                                {card.title.slice(0, 1).toUpperCase()}
                            </div>
                        )}
                    </div>
                    <div className={styles.body}>
                        <div className={styles.meta}>
                            <h3>{card.title}</h3>
                            <p>{card.description}</p>
                        </div>
                        <span className={styles.count}>
                            {t('IMAGES_PHOTOS_COUNT', { count: String(card.imageCount) })}
                        </span>
                    </div>
                </button>
            ))}
        </div>
    );
};

export default ImageCollectionsPanel;
