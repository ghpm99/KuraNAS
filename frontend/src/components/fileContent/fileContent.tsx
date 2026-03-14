import { FileType, formatSize } from '@/utils';
import FileCard from '../fileCard';
import useI18n from '../i18n/provider/i18nContext';
import useFile, { FileData } from '../providers/fileProvider/fileContext';
import FileViewer from './components/fileViewer/fileViewer';
import { getApiV1BaseUrl } from '@/service/apiUrl';
import styles from './fileContent.module.css';

interface FileContentProps {
	showHeading?: boolean;
	viewMode?: 'grid' | 'list';
}

const FileContent = ({ showHeading = true, viewMode = 'grid' }: FileContentProps) => {
	const { status, handleSelectItem, selectedItem, files, handleStarredItem, fileListFilter } = useFile();
	const { t } = useI18n();
	const currentListTitle =
		fileListFilter === 'starred' ? t('STARRED_FILES') : fileListFilter === 'recent' ? t('RECENT_FILES') : t('FILES');

	if (status === 'pending') {
		return <div className={styles.fileContent}>{t('LOADING')}</div>;
	}
	if (status === 'error') {
		return <div className={styles.fileContent}>{t('ERROR_LOADING_FILES')}</div>;
	}

	const fileMetadata = (file: FileData): string => {
		if (file.type === FileType.File) {
			const format = file.format ? `${file.format} - ` : '';
			const fileSize = formatSize(file.size);

			return `${format}${fileSize}`;
		}
		const directoryContentCount = file.directory_content_count;
		const countText = directoryContentCount > 1 ? t('ITENS') : t('ITEM');
		return `${t('FOLDER')} - ${directoryContentCount} ${countText}`;
	};

	const thumbnailUrl = (id: number) => `${getApiV1BaseUrl()}/files/thumbnail/${id}`;

	const renderCollection = (title: string, items: FileData[]) => {
		if (items.length === 0) {
			return (
				<div className={styles.fileContent}>
					{showHeading ? <h1 className={styles.title}>{title}</h1> : null}
					<div className={styles.emptyState}>{t('EMPTY_FILE_LIST')}</div>
				</div>
			);
		}

		return (
			<div className={styles.fileContent}>
				{showHeading ? <h1 className={styles.title}>{title}</h1> : null}
				{viewMode === 'list' ? (
					<div className={styles.fileList}>
						{items.map((file) => (
							<div key={file.id} className={styles.listRow}>
								<button
									type='button'
									className={styles.listButton}
									onClick={() => handleSelectItem(file.id)}
									aria-label={file.name}
								>
									<img
										src={thumbnailUrl(file.id)}
										alt={file.name}
										loading='lazy'
										className={styles.listThumbnail}
									/>
									<div className={styles.listContent}>
										<span className={styles.listTitle}>{file.name}</span>
										<span className={styles.listMetadata}>{fileMetadata(file)}</span>
									</div>
								</button>
								<button type='button' className={styles.listStarButton} onClick={() => handleStarredItem(file.id)}>
									{file.starred ? '★' : '☆'}
								</button>
							</div>
						))}
					</div>
				) : (
					<div className={styles.fileGrid}>
						{items.map((file) => (
							<FileCard
								key={file.id}
								title={file.name}
								starred={file.starred}
								metadata={fileMetadata(file)}
								thumbnail={thumbnailUrl(file.id)}
								onClick={() => handleSelectItem(file.id)}
								onClickStar={() => handleStarredItem(file.id)}
							/>
						))}
					</div>
				)}
			</div>
		);
	};

	if (!selectedItem) {
		return renderCollection(currentListTitle, files ?? []);
	}

	if (selectedItem.type === FileType.Directory) {
		return renderCollection(selectedItem.name, selectedItem.file_children ?? []);
	}

	return (
		<div className={styles.previewContainer}>
			<FileViewer file={selectedItem} />
		</div>
	);
};

export default FileContent;
