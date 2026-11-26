import { ChangeEvent, DragEvent, useRef } from 'react';
import Button from '@mui/material/Button';
import useI18n from '@/components/i18n/provider/i18nContext';
import styles from './TakeoutDropZone.module.css';

type TakeoutDropZoneProps = {
	onSelectFile: (file: File) => void;
};

const TakeoutDropZone = ({ onSelectFile }: TakeoutDropZoneProps) => {
	const { t } = useI18n();
	const inputRef = useRef<HTMLInputElement | null>(null);

	const pickFiles = (files: FileList | null) => {
		if (!files) return;
		for (let fileIndex = 0; fileIndex < files.length; fileIndex += 1) {
			const file = files.item(fileIndex);
			if (file) onSelectFile(file);
		}
	};

	const handleInputChange = (event: ChangeEvent<HTMLInputElement>) => {
		pickFiles(event.target.files);
	};

	const handleDrop = (event: DragEvent<HTMLDivElement>) => {
		event.preventDefault();
		pickFiles(event.dataTransfer.files);
	};

	return (
		<div
			className={styles.dropZone}
			onDragOver={(event) => event.preventDefault()}
			onDrop={handleDrop}
			onClick={() => inputRef.current?.click()}
			role="button"
			tabIndex={0}
			onKeyDown={(event) => {
				if (event.key === 'Enter' || event.key === ' ') {
					event.preventDefault();
					inputRef.current?.click();
				}
			}}
		>
			<strong>{t('TAKEOUT_SELECT_FILE')}</strong>
			<span>{t('TAKEOUT_DRAG_DROP')}</span>
			<Button variant="outlined">{t('TAKEOUT_SELECT_FILE')}</Button>
			<input
				ref={inputRef}
				type="file"
				multiple
				accept=".zip,application/zip"
				className={styles.input}
				onChange={handleInputChange}
			/>
		</div>
	);
};

export default TakeoutDropZone;
