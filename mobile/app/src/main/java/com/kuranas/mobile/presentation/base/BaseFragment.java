package com.kuranas.mobile.presentation.base;

import androidx.fragment.app.Fragment;

import com.kuranas.mobile.i18n.TranslationManager;

public abstract class BaseFragment extends Fragment {

    private static TranslationManager translationManagerInstance;

    public static void setTranslationManager(TranslationManager manager) {
        translationManagerInstance = manager;
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
}
