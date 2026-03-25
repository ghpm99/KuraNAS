package com.kuranas.mobile.presentation.connection;

import android.content.Intent;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.ProgressBar;
import android.widget.TextView;

import androidx.appcompat.app.AppCompatActivity;

import com.kuranas.mobile.R;
import com.kuranas.mobile.app.MainActivity;
import com.kuranas.mobile.app.ServiceLocator;
import com.kuranas.mobile.domain.model.DiscoveryResult;
import com.kuranas.mobile.domain.port.ServerDiscoveryPort;
import com.kuranas.mobile.i18n.TranslationManager;
import com.kuranas.mobile.infra.discovery.CachedServerStrategy;
import com.kuranas.mobile.infra.discovery.DiscoveryStrategy;
import com.kuranas.mobile.infra.discovery.NetworkScanStrategy;
import com.kuranas.mobile.infra.discovery.NsdDiscoveryStrategy;
import com.kuranas.mobile.infra.discovery.ServerDiscovery;
import com.kuranas.mobile.infra.discovery.ServerValidator;
import com.kuranas.mobile.infra.discovery.UdpDiscoveryStrategy;
import com.kuranas.mobile.infra.preferences.ServerPreferences;

import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class ConnectionActivity extends AppCompatActivity {

    private ProgressBar progressBar;
    private TextView statusText;
    private View manualInputSection;
    private TextView failedText;
    private EditText serverAddressInput;
    private Button connectButton;
    private Button retryButton;

    private ServerDiscovery serverDiscovery;
    private ServerPreferences serverPreferences;
    private ServerValidator serverValidator;
    private TranslationManager translationManager;
    private final Handler mainHandler = new Handler(Looper.getMainLooper());
    private final ExecutorService executor = Executors.newSingleThreadExecutor();

    private final Map<String, String> strategyLabels = new HashMap<String, String>();

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_connection);

        translationManager = TranslationManager.getInstance();

        progressBar = (ProgressBar) findViewById(R.id.discovery_progress);
        statusText = (TextView) findViewById(R.id.discovery_status);
        manualInputSection = findViewById(R.id.manual_input_section);
        failedText = (TextView) findViewById(R.id.discovery_failed_text);
        serverAddressInput = (EditText) findViewById(R.id.server_address_input);
        connectButton = (Button) findViewById(R.id.connect_button);
        retryButton = (Button) findViewById(R.id.retry_button);

        initStrategyLabels();
        applyTranslations();

        serverPreferences = new ServerPreferences(this);
        serverValidator = new ServerValidator();

        connectButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                onManualConnect();
            }
        });

        retryButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                showDiscoveryState();
                startDiscovery();
            }
        });

        startDiscovery();
    }

    private void initStrategyLabels() {
        strategyLabels.put("cache", translationManager.get("DISCOVERY_CHECKING_CACHE"));
        strategyLabels.put("mdns", translationManager.get("DISCOVERY_MDNS"));
        strategyLabels.put("udp", translationManager.get("DISCOVERY_UDP"));
        strategyLabels.put("scan", translationManager.get("DISCOVERY_SCANNING"));
    }

    private void applyTranslations() {
        serverAddressInput.setHint(translationManager.get("DISCOVERY_MANUAL_HINT"));
        connectButton.setText(translationManager.get("DISCOVERY_CONNECT"));
        retryButton.setText(translationManager.get("DISCOVERY_RETRY"));
    }

    private void startDiscovery() {
        List<DiscoveryStrategy> strategies = Arrays.<DiscoveryStrategy>asList(
                new CachedServerStrategy(serverPreferences, serverValidator),
                new NsdDiscoveryStrategy(this, serverValidator),
                new UdpDiscoveryStrategy(serverValidator),
                new NetworkScanStrategy(this, serverValidator)
        );

        serverDiscovery = new ServerDiscovery(strategies, serverPreferences);

        serverDiscovery.discover(new ServerDiscoveryPort.DiscoveryCallback() {
            @Override
            public void onDiscovered(final DiscoveryResult result) {
                mainHandler.post(new Runnable() {
                    @Override
                    public void run() {
                        onServerFound(result);
                    }
                });
            }

            @Override
            public void onFailed(final String reason) {
                mainHandler.post(new Runnable() {
                    @Override
                    public void run() {
                        showManualInput();
                    }
                });
            }

            @Override
            public void onProgress(final String strategyName) {
                mainHandler.post(new Runnable() {
                    @Override
                    public void run() {
                        String label = strategyLabels.get(strategyName);
                        if (label != null) {
                            statusText.setText(label);
                        }
                    }
                });
            }
        });
    }

    private void onServerFound(DiscoveryResult result) {
        ServiceLocator.initialize(result.getServerUrl());

        translationManager.loadAsync(new Runnable() {
            @Override
            public void run() {
                // Remote translations loaded
            }
        });

        navigateToMain();
    }

    private void showManualInput() {
        progressBar.setVisibility(View.GONE);
        statusText.setVisibility(View.GONE);
        manualInputSection.setVisibility(View.VISIBLE);
        failedText.setText(translationManager.get("DISCOVERY_FAILED"));
    }

    private void showDiscoveryState() {
        progressBar.setVisibility(View.VISIBLE);
        statusText.setVisibility(View.VISIBLE);
        manualInputSection.setVisibility(View.GONE);
    }

    private void onManualConnect() {
        final String input = serverAddressInput.getText().toString().trim();
        if (input.isEmpty()) {
            failedText.setText(translationManager.get("DISCOVERY_INVALID_ADDRESS"));
            return;
        }

        final String candidateUrl;
        if (input.startsWith("http://") || input.startsWith("https://")) {
            candidateUrl = input;
        } else {
            candidateUrl = "http://" + input + ":8000";
        }

        connectButton.setEnabled(false);
        connectButton.setText(translationManager.get("DISCOVERY_CONNECTING"));

        executor.submit(new Runnable() {
            @Override
            public void run() {
                final boolean valid = serverValidator.validate(candidateUrl);
                mainHandler.post(new Runnable() {
                    @Override
                    public void run() {
                        connectButton.setEnabled(true);
                        connectButton.setText(translationManager.get("DISCOVERY_CONNECT"));

                        if (valid) {
                            serverPreferences.saveServerUrl(candidateUrl);
                            onServerFound(new DiscoveryResult(candidateUrl, "manual"));
                        } else {
                            failedText.setText(translationManager.get("DISCOVERY_SERVER_NOT_FOUND"));
                        }
                    }
                });
            }
        });
    }

    private void navigateToMain() {
        Intent intent = new Intent(this, MainActivity.class);
        intent.setFlags(Intent.FLAG_ACTIVITY_NEW_TASK | Intent.FLAG_ACTIVITY_CLEAR_TASK);
        startActivity(intent);
        finish();
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        if (serverDiscovery != null) {
            serverDiscovery.cancel();
        }
    }
}
