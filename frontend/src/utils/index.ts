export enum FileType {
	Directory = 1,
	File = 2,
}

export const formatSize = (size: number): string => {
	if (size < 1024) return `${size} B`;
	const units = ['KB', 'MB', 'GB', 'TB'];
	let unitIndex = -1;
	let formattedSize = size;

	do {
		formattedSize /= 1024;
		unitIndex++;
	} while (formattedSize >= 1024 && unitIndex < units.length - 1);

	return `${formattedSize.toFixed(2)} ${units[unitIndex]}`;
};

export const formatDate = (dateString: string): string => {
	try {
		const date = new Date(dateString);
		return date.toLocaleString();
	} catch (error) {
		console.error('Erro ao formatar a data:', error);
		return dateString;
	}
};

export const formatDuration = (seconds: number | undefined): string => {
	if (!seconds || seconds <= 0) return 'Em andamento';

	const hours = Math.floor(seconds / 3600);
	const minutes = Math.floor((seconds % 3600) / 60);
	const secs = seconds % 60;

	if (hours > 0) {
		return `${hours}h ${minutes}m ${secs}s`;
	} else if (minutes > 0) {
		return `${minutes}m ${secs}s`;
	} else {
		return `${secs}s`;
	}
};

export const formatDateTime = (date: Date): string => {
	return date.toLocaleString('pt-BR', {
		day: '2-digit',
		month: '2-digit',
		year: 'numeric',
		hour: '2-digit',
		minute: '2-digit',
		second: '2-digit',
	});
};

type formatType = {
	type: 'image' | 'audio' | 'video' | 'document' | 'archive' | 'unknown';
	mime: string;
	description: string;
};

export const getFileTypeInfo = (format: string): formatType => {
	const fileTypes: Record<string, formatType> = {
		// Imagens
		'.jpg': { type: 'image', mime: 'image/jpeg', description: 'IMAGE_JPEG' },
		'.jpeg': { type: 'image', mime: 'image/jpeg', description: 'IMAGE_JPEG' },
		'.png': { type: 'image', mime: 'image/png', description: 'IMAGE_PNG' },
		'.gif': { type: 'image', mime: 'image/gif', description: 'IMAGE_GIF' },
		'.bmp': { type: 'image', mime: 'image/bmp', description: 'IMAGE_BMP' },
		'.svg': { type: 'image', mime: 'image/svg+xml', description: 'IMAGE_SVG' },
		'.webp': { type: 'image', mime: 'image/webp', description: 'IMAGE_WEBP' },

		// Áudios
		'.mp3': { type: 'audio', mime: 'audio/mpeg', description: 'AUDIO_MP3' },
		'.wav': { type: 'audio', mime: 'audio/wav', description: 'AUDIO_WAV' },
		'.aac': { type: 'audio', mime: 'audio/aac', description: 'AUDIO_AAC' },
		'.flac': { type: 'audio', mime: 'audio/flac', description: 'AUDIO_FLAC' },

		// Vídeos
		'.mp4': { type: 'video', mime: 'video/mp4', description: 'VIDEO_MP4' },
		'.webm': { type: 'video', mime: 'video/webm', description: 'VIDEO_WEBM' },
		'.ogg': { type: 'video', mime: 'video/ogg', description: 'VIDEO_OGG' },
		'.mov': { type: 'video', mime: 'video/quicktime', description: 'VIDEO_MOV' },

		// Documentos
		'.pdf': { type: 'document', mime: 'application/pdf', description: 'DOCUMENT_PDF' },
		'.txt': { type: 'document', mime: 'text/plain', description: 'DOCUMENT_TXT' },
		'.html': { type: 'document', mime: 'text/html', description: 'DOCUMENT_HTML' },
		'.htm': { type: 'document', mime: 'text/html', description: 'DOCUMENT_HTML' },
		'.xml': { type: 'document', mime: 'application/xml', description: 'DOCUMENT_XML' },
		'.json': { type: 'document', mime: 'application/json', description: 'DOCUMENT_JSON' },
		'.csv': { type: 'document', mime: 'text/csv', description: 'DOCUMENT_CSV' },

		// Outros
		'.zip': { type: 'archive', mime: 'application/zip', description: 'ARCHIVE_ZIP' },
		'.rar': { type: 'archive', mime: 'application/vnd.rar', description: 'ARCHIVE_RAR' },
		'.7z': { type: 'archive', mime: 'application/x-7z-compressed', description: 'ARCHIVE_7Z' },
		'.tar': { type: 'archive', mime: 'application/x-tar', description: 'ARCHIVE_TAR' },
		'.gz': { type: 'archive', mime: 'application/gzip', description: 'ARCHIVE_GZIP' },
	};

	return fileTypes[format.toLowerCase()] || { type: 'unknown', mime: '', description: 'UNKNOWN_FORMAT' };
};
