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

	const pickFile = (files: FileList | null) => {
		if (!files || files.length === 0) {
			return;
		}
		const file = files.item(0);
		if (!file) {
			return;
		}
		onSelectFile(file);
	};

	const handleInputChange = (event: ChangeEvent<HTMLInputElement>) => {
		pickFile(event.target.files);
	};

	const handleDrop = (event: DragEvent<HTMLDivElement>) => {
		event.preventDefault();
		pickFile(event.dataTransfer.files);
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
				accept=".zip,application/zip"
				className={styles.input}
				onChange={handleInputChange}
			/>
		</div>
	);
};

export default TakeoutDropZone;
