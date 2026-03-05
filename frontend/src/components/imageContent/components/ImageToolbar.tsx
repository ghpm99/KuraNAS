import { Search } from 'lucide-react';
import type { ImageGroupBy } from '@/components/providers/imageProvider/imageProvider';
import useI18n from '@/components/i18n/provider/i18nContext';
import controlsStyles from '../imageContentControls.module.css';

type ImageToolbarProps = {
	filteredCount: number;
	totalCount: number;
	search: string;
	groupBy: ImageGroupBy;
	groupByLabels: Record<ImageGroupBy, string>;
	onSearchChange: (value: string) => void;
	onGroupByChange: (value: ImageGroupBy) => void;
};

export default function ImageToolbar({
	filteredCount,
	totalCount,
	search,
	groupBy,
	groupByLabels,
	onSearchChange,
	onGroupByChange,
}: ImageToolbarProps) {
	const { t } = useI18n();

	return (
		<div className='images-toolbar'>
			<div>
				<h2>{t('IMAGES_TITLE')}</h2>
				<p>{t('IMAGES_COUNT_SUMMARY', { filtered: String(filteredCount), total: String(totalCount) })}</p>
			</div>
			<label className='images-search'>
				<Search size={16} />
				<input
					type='search'
					value={search}
					onChange={(event) => onSearchChange(event.target.value)}
					placeholder={t('IMAGES_SEARCH_PLACEHOLDER')}
				/>
			</label>
			<label className={controlsStyles.groupingSelect}>
				<span>{t('IMAGES_GROUP_BY_LABEL')}</span>
				<select aria-label={t('IMAGES_GROUP_BY_ARIA')} value={groupBy} onChange={(event) => onGroupByChange(event.target.value as ImageGroupBy)}>
					{(Object.keys(groupByLabels) as ImageGroupBy[]).map((key) => (
						<option key={key} value={key}>
							{groupByLabels[key]}
						</option>
					))}
				</select>
			</label>
		</div>
	);
}
