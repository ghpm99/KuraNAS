package com.kuranas.mobile.presentation.common;

import android.graphics.Bitmap;
import android.graphics.BitmapFactory;
import android.os.AsyncTask;
import android.widget.ImageView;

import com.kuranas.mobile.infra.cache.BitmapCache;
import com.kuranas.mobile.infra.logging.AppLogger;

import java.io.InputStream;
import java.lang.ref.WeakReference;
import java.net.HttpURLConnection;
import java.net.URL;

public class BitmapLoaderTask extends AsyncTask<String, Void, Bitmap> {

    private static final String LOG_TAG = "BitmapLoaderTask";

    private final BitmapCache cache;
    private final WeakReference<ImageView> imageViewRef;
    private final int targetWidth;
    private final int targetHeight;
    private String url;

    public BitmapLoaderTask(BitmapCache cache, ImageView imageView,
                            int targetWidth, int targetHeight) {
        this.cache = cache;
        this.imageViewRef = new WeakReference<ImageView>(imageView);
        this.targetWidth = targetWidth;
        this.targetHeight = targetHeight;
    }

    @Override
    protected Bitmap doInBackground(String... params) {
        if (params == null || params.length == 0) {
            return null;
        }
        url = params[0];

        HttpURLConnection connection = null;
        InputStream inputStream = null;
        InputStream sampleStream = null;
        try {
            connection = (HttpURLConnection) new URL(url).openConnection();
            connection.setConnectTimeout(10000);
            connection.setReadTimeout(15000);
            connection.setDoInput(true);
            connection.connect();

            int responseCode = connection.getResponseCode();
            if (responseCode != HttpURLConnection.HTTP_OK) {
                AppLogger.w(LOG_TAG, "HTTP " + responseCode + " for " + url);
                return null;
            }

            inputStream = connection.getInputStream();

            BitmapFactory.Options boundsOptions = new BitmapFactory.Options();
            boundsOptions.inJustDecodeBounds = true;
            BitmapFactory.decodeStream(inputStream, null, boundsOptions);

            closeQuietly(inputStream);
            connection.disconnect();

            int sampleSize = calculateInSampleSize(
                    boundsOptions.outWidth, boundsOptions.outHeight,
                    targetWidth, targetHeight);

            connection = (HttpURLConnection) new URL(url).openConnection();
            connection.setConnectTimeout(10000);
            connection.setReadTimeout(15000);
            connection.setDoInput(true);
            connection.connect();

            sampleStream = connection.getInputStream();

            BitmapFactory.Options decodeOptions = new BitmapFactory.Options();
            decodeOptions.inSampleSize = sampleSize;
            Bitmap bitmap = BitmapFactory.decodeStream(sampleStream, null, decodeOptions);

            if (bitmap != null && cache != null) {
                cache.put(url, bitmap);
            }

            return bitmap;
        } catch (Exception e) {
            AppLogger.e(LOG_TAG, "Failed to load bitmap: " + url, e);
            return null;
        } finally {
            closeQuietly(inputStream);
            closeQuietly(sampleStream);
            if (connection != null) {
                connection.disconnect();
            }
        }
    }

    @Override
    protected void onPostExecute(Bitmap bitmap) {
        if (isCancelled()) {
            return;
        }
        ImageView imageView = imageViewRef.get();
        if (imageView != null && bitmap != null) {
            imageView.setImageBitmap(bitmap);
        }
    }

    public static void load(String url, ImageView imageView, BitmapCache cache,
                            int targetWidth, int targetHeight) {
        if (url == null || imageView == null) {
            return;
        }

        if (cache != null) {
            Bitmap cached = cache.get(url);
            if (cached != null) {
                imageView.setImageBitmap(cached);
                return;
            }
        }

        BitmapLoaderTask task = new BitmapLoaderTask(cache, imageView,
                targetWidth, targetHeight);
        task.execute(url);
    }

    private static int calculateInSampleSize(int rawWidth, int rawHeight,
                                             int reqWidth, int reqHeight) {
        int inSampleSize = 1;
        if (rawHeight > reqHeight || rawWidth > reqWidth) {
            int halfHeight = rawHeight / 2;
            int halfWidth = rawWidth / 2;
            while ((halfHeight / inSampleSize) >= reqHeight
                    && (halfWidth / inSampleSize) >= reqWidth) {
                inSampleSize *= 2;
            }
        }
        return inSampleSize;
    }

    private static void closeQuietly(InputStream stream) {
        if (stream != null) {
            try {
                stream.close();
            } catch (Exception ignored) {
            }
        }
    }
}
