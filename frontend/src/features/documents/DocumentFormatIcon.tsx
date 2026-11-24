import { File, FileSpreadsheet, FileText, Presentation } from 'lucide-react';
import { getFileTypeInfo } from '@/utils';

const spreadsheetKey = 'DOCUMENT_SPREADSHEET';
const presentationKey = 'DOCUMENT_PRESENTATION';

type DocumentFormatIconProps = {
    format: string;
    size?: number;
};

const DocumentFormatIcon = ({ format, size = 20 }: DocumentFormatIconProps) => {
    const { type, description } = getFileTypeInfo(format);

    if (description === spreadsheetKey) return <FileSpreadsheet size={size} aria-hidden="true" />;
    if (description === presentationKey) return <Presentation size={size} aria-hidden="true" />;
    if (type === 'document' || type === 'text') return <FileText size={size} aria-hidden="true" />;
    return <File size={size} aria-hidden="true" />;
};

export default DocumentFormatIcon;
