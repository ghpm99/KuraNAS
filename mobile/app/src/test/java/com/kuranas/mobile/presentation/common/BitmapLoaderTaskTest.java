package com.kuranas.mobile.presentation.common;

import org.junit.Test;

import java.lang.reflect.Method;

import static org.junit.Assert.assertEquals;

public class BitmapLoaderTaskTest {

    @Test
    public void calculateInSampleSize_noDownscaleWhenSmallerThanTarget() throws Exception {
        int result = invokeCalculateInSampleSize(100, 100, 200, 200);

        assertEquals(1, result);
    }

    @Test
    public void calculateInSampleSize_downscalesLargeImage() throws Exception {
        int result = invokeCalculateInSampleSize(800, 800, 200, 200);

        assertEquals(4, result);
    }

    @Test
    public void calculateInSampleSize_downscalesVeryLargeImage() throws Exception {
        int result = invokeCalculateInSampleSize(3200, 3200, 200, 200);

        assertEquals(16, result);
    }

    @Test
    public void calculateInSampleSize_handlesRectangularImage() throws Exception {
        int result = invokeCalculateInSampleSize(1600, 800, 200, 200);

        assertEquals(4, result);
    }

    @Test
    public void calculateInSampleSize_exactMatchReturnsOne() throws Exception {
        int result = invokeCalculateInSampleSize(200, 200, 200, 200);

        assertEquals(1, result);
    }

    @Test
    public void calculateInSampleSize_zeroTargetReturnsOne() throws Exception {
        int result = invokeCalculateInSampleSize(0, 0, 200, 200);

        assertEquals(1, result);
    }

    @Test
    public void readAllBytes_readsFullStream() throws Exception {
        byte[] input = new byte[]{1, 2, 3, 4, 5};
        java.io.InputStream stream = new java.io.ByteArrayInputStream(input);

        Method method = BitmapLoaderTask.class.getDeclaredMethod(
                "readAllBytes", java.io.InputStream.class);
        method.setAccessible(true);
        byte[] result = (byte[]) method.invoke(null, stream);

        assertEquals(input.length, result.length);
        for (int i = 0; i < input.length; i++) {
            assertEquals(input[i], result[i]);
        }
    }

    @Test
    public void readAllBytes_emptyStreamReturnsEmptyArray() throws Exception {
        java.io.InputStream stream = new java.io.ByteArrayInputStream(new byte[0]);

        Method method = BitmapLoaderTask.class.getDeclaredMethod(
                "readAllBytes", java.io.InputStream.class);
        method.setAccessible(true);
        byte[] result = (byte[]) method.invoke(null, stream);

        assertEquals(0, result.length);
    }

    @Test
    public void readAllBytes_handlesLargePayload() throws Exception {
        byte[] input = new byte[10000];
        for (int i = 0; i < input.length; i++) {
            input[i] = (byte) (i % 256);
        }
        java.io.InputStream stream = new java.io.ByteArrayInputStream(input);

        Method method = BitmapLoaderTask.class.getDeclaredMethod(
                "readAllBytes", java.io.InputStream.class);
        method.setAccessible(true);
        byte[] result = (byte[]) method.invoke(null, stream);

        assertEquals(input.length, result.length);
        assertEquals(input[0], result[0]);
        assertEquals(input[9999], result[9999]);
    }

    private int invokeCalculateInSampleSize(int rawWidth, int rawHeight,
                                            int reqWidth, int reqHeight) throws Exception {
        Method method = BitmapLoaderTask.class.getDeclaredMethod(
                "calculateInSampleSize", int.class, int.class, int.class, int.class);
        method.setAccessible(true);
        return (Integer) method.invoke(null, rawWidth, rawHeight, reqWidth, reqHeight);
    }
}
