package com.kuranas.mobile.domain.model;

import org.junit.Test;

import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

public class FileItemTest {

    private FileItem createFileItem(int type, String format) {
        return new FileItem(1, "test", "/test", "/", type, format, 0L, "", "", false, 0);
    }

    @Test
    public void isDirectory_withDirectoryType_returnsTrue() {
        FileItem item = createFileItem(FileItem.TYPE_DIRECTORY, "");
        assertTrue(item.isDirectory());
        assertFalse(item.isFile());
    }

    @Test
    public void isFile_withFileType_returnsTrue() {
        FileItem item = createFileItem(FileItem.TYPE_FILE, "txt");
        assertTrue(item.isFile());
        assertFalse(item.isDirectory());
    }

    @Test
    public void isImage_withJpg_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "jpg").isImage());
    }

    @Test
    public void isImage_withJpeg_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "jpeg").isImage());
    }

    @Test
    public void isImage_withPng_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "png").isImage());
    }

    @Test
    public void isImage_withGif_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "gif").isImage());
    }

    @Test
    public void isImage_withWebp_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "webp").isImage());
    }

    @Test
    public void isImage_withBmp_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "bmp").isImage());
    }

    @Test
    public void isImage_withNonImageFormat_returnsFalse() {
        assertFalse(createFileItem(FileItem.TYPE_FILE, "mp3").isImage());
        assertFalse(createFileItem(FileItem.TYPE_FILE, "txt").isImage());
        assertFalse(createFileItem(FileItem.TYPE_FILE, "mp4").isImage());
    }

    @Test
    public void isAudio_withMp3_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "mp3").isAudio());
    }

    @Test
    public void isAudio_withFlac_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "flac").isAudio());
    }

    @Test
    public void isAudio_withOgg_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "ogg").isAudio());
    }

    @Test
    public void isAudio_withWav_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "wav").isAudio());
    }

    @Test
    public void isAudio_withM4a_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "m4a").isAudio());
    }

    @Test
    public void isAudio_withAac_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "aac").isAudio());
    }

    @Test
    public void isAudio_withNonAudioFormat_returnsFalse() {
        assertFalse(createFileItem(FileItem.TYPE_FILE, "jpg").isAudio());
        assertFalse(createFileItem(FileItem.TYPE_FILE, "mp4").isAudio());
        assertFalse(createFileItem(FileItem.TYPE_FILE, "pdf").isAudio());
    }

    @Test
    public void isVideo_withMp4_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "mp4").isVideo());
    }

    @Test
    public void isVideo_withMkv_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "mkv").isVideo());
    }

    @Test
    public void isVideo_withAvi_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "avi").isVideo());
    }

    @Test
    public void isVideo_withMov_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "mov").isVideo());
    }

    @Test
    public void isVideo_withWmv_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "wmv").isVideo());
    }

    @Test
    public void isVideo_withWebm_returnsTrue() {
        assertTrue(createFileItem(FileItem.TYPE_FILE, "webm").isVideo());
    }

    @Test
    public void isVideo_withNonVideoFormat_returnsFalse() {
        assertFalse(createFileItem(FileItem.TYPE_FILE, "jpg").isVideo());
        assertFalse(createFileItem(FileItem.TYPE_FILE, "mp3").isVideo());
        assertFalse(createFileItem(FileItem.TYPE_FILE, "doc").isVideo());
    }

    @Test
    public void isImage_withNullFormat_returnsFalse() {
        FileItem item = new FileItem(1, "test", "/test", "/", FileItem.TYPE_FILE, null, 0L, "", "", false, 0);
        assertFalse(item.isImage());
        assertFalse(item.isAudio());
        assertFalse(item.isVideo());
    }
}
