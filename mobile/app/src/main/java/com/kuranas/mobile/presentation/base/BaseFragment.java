package com.kuranas.mobile.presentation.base;

import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import androidx.fragment.app.Fragment;
import androidx.swiperefreshlayout.widget.SwipeRefreshLayout;

import com.kuranas.mobile.R;
import com.kuranas.mobile.i18n.TranslationManager;

public abstract class BaseFragment extends Fragment {

    private static TranslationManager translationManagerInstance;

    protected View loadingView;
    protected View emptyView;
    protected View errorView;
    protected View contentView;

    public static void setTranslationManager(TranslationManager manager) {
        translationManagerInstance = manager;
    }

    protected void initStateViews(View root) {
        loadingView = root.findViewById(R.id.loading_view);
        emptyView = root.findViewById(R.id.empty_view);
        errorView = root.findViewById(R.id.error_view);

        contentView = root.findViewById(R.id.swipe_refresh);
        if (contentView == null) {
            contentView = findFirstSwipeRefreshLayout(root);
        }
    }

    protected void setState(ViewState state) {
        if (loadingView != null) {
            loadingView.setVisibility(state == ViewState.LOADING ? View.VISIBLE : View.GONE);
        }
        if (contentView != null) {
            contentView.setVisibility(state == ViewState.CONTENT ? View.VISIBLE : View.GONE);
        }
        if (emptyView != null) {
            emptyView.setVisibility(state == ViewState.EMPTY ? View.VISIBLE : View.GONE);
        }
        if (errorView != null) {
            errorView.setVisibility(state == ViewState.ERROR ? View.VISIBLE : View.GONE);
        }
    }

    protected void setErrorMessage(String msg) {
        if (errorView == null) {
            return;
        }
        TextView errorText = (TextView) errorView.findViewById(R.id.error_message);
        if (errorText != null) {
            errorText.setText(msg);
        }
    }

    protected void setEmptyMessage(String msg) {
        if (emptyView == null) {
            return;
        }
        TextView emptyText = (TextView) emptyView.findViewById(R.id.empty_message);
        if (emptyText != null) {
            emptyText.setText(msg);
        }
    }

    protected void setRetryListener(View.OnClickListener listener) {
        if (errorView == null) {
            return;
        }
        View retryButton = errorView.findViewById(R.id.btn_retry);
        if (retryButton != null) {
            retryButton.setOnClickListener(listener);
        }
    }

    protected TranslationManager getTranslations() {
        return translationManagerInstance;
    }

    protected String t(String key) {
        TranslationManager manager = getTranslations();
        if (manager == null) {
            return key;
        }
        return manager.t(key);
    }

    private View findFirstSwipeRefreshLayout(View root) {
        if (root instanceof SwipeRefreshLayout) {
            return root;
        }
        if (root instanceof ViewGroup) {
            ViewGroup group = (ViewGroup) root;
            for (int i = 0; i < group.getChildCount(); i++) {
                View child = group.getChildAt(i);
                if (child instanceof SwipeRefreshLayout) {
                    return child;
                }
            }
        }
        return null;
    }
}
