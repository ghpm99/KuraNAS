import useI18n from '@/components/i18n/provider/i18nContext';
import {
	completeTakeoutUpload,
	initTakeoutUpload,
	uploadTakeoutChunk,
} from '@/service/takeout';
import { formatSize } from '@/shared/utils/formatSize';
import { extractBackendErrorMessage } from '@/shared/utils/extractBackendErrorMessage';
import { useCallback, useMemo, useState } from 'react';

export type TakeoutUploadState = 'idle' | 'selecting' | 'uploading' | 'completing' | 'done' | 'error';

const useTakeoutUpload = () => {
	const { t } = useI18n();
	const [state, setState] = useState<TakeoutUploadState>('idle');
	const [selectedFile, setSelectedFile] = useState<File | null>(null);
	const [progress, setProgress] = useState(0);
	const [uploadedBytes, setUploadedBytes] = useState(0);
	const [jobId, setJobId] = useState<number | null>(null);
	const [importId, setImportId] = useState<number | null>(null);
	const [errorMessage, setErrorMessage] = useState('');

	const fileName = selectedFile?.name ?? '';

	const selectFile = useCallback(
		(file: File) => {
			if (!file.name.toLowerCase().endsWith('.zip')) {
				setState('error');
				setErrorMessage(t('TAKEOUT_FILE_TYPE_ERROR'));
				return;
			}
			setSelectedFile(file);
			setProgress(0);
			setUploadedBytes(0);
			setJobId(null);
		setImportId(null);
			setImportId(null);
			setErrorMessage('');
			setState('selecting');
		},
		[t]
	);

	const reset = useCallback(() => {
		setState('idle');
		setSelectedFile(null);
		setProgress(0);
		setUploadedBytes(0);
		setJobId(null);
		setImportId(null);
		setErrorMessage('');
	}, []);

	const startUpload = useCallback(async () => {
		if (!selectedFile) {
			return;
		}

		try {
			setState('uploading');
			setErrorMessage('');

			const initResult = await initTakeoutUpload(selectedFile.name, selectedFile.size);
			const chunkSize = Math.max(1, initResult.chunk_size);

			let offset = 0;
			while (offset < selectedFile.size) {
				const chunk = selectedFile.slice(offset, offset + chunkSize);
				await uploadTakeoutChunk(initResult.upload_id, chunk, offset);
				offset += chunk.size;
				setUploadedBytes(offset);
				setProgress(Math.min(100, Math.round((offset / selectedFile.size) * 100)));
			}

			setState('completing');
			const completeResult = await completeTakeoutUpload(initResult.upload_id);
			setJobId(completeResult.job_id);
			setImportId(completeResult.import_id ?? null);
			setProgress(100);
			setState('done');
		} catch (uploadError) {
			setState('error');
			setErrorMessage(extractBackendErrorMessage(uploadError) ?? t('TAKEOUT_UPLOAD_ERROR'));
		}
	}, [selectedFile, t]);

	const progressMessage = useMemo(() => {
		if (!selectedFile) {
			return '';
		}
		return t('TAKEOUT_UPLOAD_PROGRESS', {
			sent: formatSize(uploadedBytes),
			total: formatSize(selectedFile.size),
		});
	}, [selectedFile, t, uploadedBytes]);

	return {
		state,
		progress,
		fileName,
		jobId,
		importId,
		errorMessage,
		progressMessage,
		selectFile,
		startUpload,
		reset,
	};
};

export default useTakeoutUpload;
