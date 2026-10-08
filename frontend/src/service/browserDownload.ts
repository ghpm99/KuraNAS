export const triggerBrowserDownload = (downloadUrl: string, fileName = ''): void => {
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = fileName;
    document.body.appendChild(link);
    link.click();
    link.remove();
};
