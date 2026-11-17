export const connectionLostErrorKey = 'ERROR_UPLOAD_CONNECTION_LOST';

export class UploadConnectionLostError extends Error {
    constructor() {
        super(connectionLostErrorKey);
        this.name = 'UploadConnectionLostError';
    }
}
