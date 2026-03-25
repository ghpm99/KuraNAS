package com.kuranas.mobile.presentation.settings;

import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.AdapterView;
import android.widget.ArrayAdapter;
import android.widget.Spinner;
import android.widget.TextView;

import androidx.fragment.app.Fragment;

import com.kuranas.mobile.R;
import com.kuranas.mobile.app.ServiceLocator;
import com.kuranas.mobile.domain.error.AppError;
import com.kuranas.mobile.domain.model.AppSettings;
import com.kuranas.mobile.domain.repository.ConfigRepository;
import com.kuranas.mobile.i18n.TranslationManager;
import com.kuranas.mobile.infra.http.ApiCallback;

import java.util.ArrayList;
import java.util.List;

public class SettingsFragment extends Fragment {

    private Spinner languageSpinner;
    private TextView aboutInfo;

    private ConfigRepository configRepository;
    private TranslationManager translationManager;

    private List<String> availableLanguages;
    private String currentLanguage;
    private boolean isInitialSelection = true;

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        View root = inflater.inflate(R.layout.fragment_settings, container, false);

        ServiceLocator locator = ServiceLocator.getInstance();
        configRepository = locator.getConfigRepository();
        translationManager = locator.getTranslationManager();

        languageSpinner = (Spinner) root.findViewById(R.id.language_spinner);
        aboutInfo = (TextView) root.findViewById(R.id.about_info);

        aboutInfo.setText("KuraNAS Mobile v1.0.0");

        loadSettings();

        return root;
    }

    private void loadSettings() {
        configRepository.getSettings(new ApiCallback<AppSettings>() {
            @Override
            public void onSuccess(AppSettings settings) {
                if (!isAdded()) {
                    return;
                }
                currentLanguage = settings.getCurrentLanguage();
                availableLanguages = settings.getAvailableLanguages();

                if (availableLanguages == null) {
                    availableLanguages = new ArrayList<String>();
                }

                populateLanguageSpinner();
            }

            @Override
            public void onError(AppError error) {
                // Settings could not be loaded
            }
        });
    }

    private void populateLanguageSpinner() {
        if (getActivity() == null) {
            return;
        }

        ArrayAdapter<String> spinnerAdapter = new ArrayAdapter<String>(
                getActivity(),
                android.R.layout.simple_spinner_item,
                availableLanguages);
        spinnerAdapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item);
        languageSpinner.setAdapter(spinnerAdapter);

        // Set current language selection
        int selectedIndex = 0;
        for (int i = 0; i < availableLanguages.size(); i++) {
            if (availableLanguages.get(i).equals(currentLanguage)) {
                selectedIndex = i;
                break;
            }
        }
        isInitialSelection = true;
        languageSpinner.setSelection(selectedIndex);

        languageSpinner.setOnItemSelectedListener(new AdapterView.OnItemSelectedListener() {
            @Override
            public void onItemSelected(AdapterView<?> parent, View view, int position, long id) {
                if (isInitialSelection) {
                    isInitialSelection = false;
                    return;
                }
                String selectedLanguage = availableLanguages.get(position);
                if (!selectedLanguage.equals(currentLanguage)) {
                    changeLanguage(selectedLanguage);
                }
            }

            @Override
            public void onNothingSelected(AdapterView<?> parent) {
                // No action needed
            }
        });
    }

    private void changeLanguage(final String language) {
        configRepository.updateLanguage(language, new ApiCallback<AppSettings>() {
            @Override
            public void onSuccess(AppSettings result) {
                if (!isAdded()) {
                    return;
                }
                currentLanguage = language;
                reloadTranslations();
            }

            @Override
            public void onError(AppError error) {
                // Revert spinner to previous selection
                if (availableLanguages != null && currentLanguage != null) {
                    for (int i = 0; i < availableLanguages.size(); i++) {
                        if (availableLanguages.get(i).equals(currentLanguage)) {
                            isInitialSelection = true;
                            languageSpinner.setSelection(i);
                            break;
                        }
                    }
                }
            }
        });
    }

    private void reloadTranslations() {
        translationManager.loadAsync(null);
    }
}
