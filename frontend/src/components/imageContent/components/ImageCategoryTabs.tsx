import useI18n from '@/components/i18n/provider/i18nContext';
import styles from '../ImageContent.module.css';

type ImageCategoryTabsProps<TCategory extends string> = {
	activeCategory: TCategory;
	labels: Record<TCategory, string>;
	counts: Record<TCategory, number>;
	onSelect: (category: TCategory) => void;
};

export default function ImageCategoryTabs<TCategory extends string>({
	activeCategory,
	labels,
	counts,
	onSelect,
}: ImageCategoryTabsProps<TCategory>) {
	const { t } = useI18n();
	const categories = Object.keys(labels) as TCategory[];

	return (
		<div className={styles.categories} role='tablist' aria-label={t('IMAGES_CATEGORIES_ARIA')}>
			{categories.map((key) => (
				<button
					key={key}
					type='button'
					className={activeCategory === key ? `${styles.categoryPill} ${styles.categoryPillActive}` : styles.categoryPill}
					onClick={() => onSelect(key)}
				>
					<span>{labels[key]}</span>
					<strong>{counts[key]}</strong>
				</button>
			))}
		</div>
	);
}
