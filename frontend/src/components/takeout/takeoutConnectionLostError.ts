export class TakeoutConnectionLostError extends Error {
    constructor() {
        super('TAKEOUT_UPLOAD_CONNECTION_LOST');
        this.name = 'TakeoutConnectionLostError';
    }
}
