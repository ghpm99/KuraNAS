package com.kuranas.android

import android.Manifest
import android.content.pm.PackageManager
import android.content.Intent
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.core.content.ContextCompat
import com.kuranas.android.core.server.ServerConnectionMonitor
import com.kuranas.android.feature.transfers.ui.TransfersIntents
import com.kuranas.android.navigation.KuranasRoot
import com.kuranas.android.ui.theme.KuranasTheme
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    private val requestNotifications =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) { /* nada a fazer */ }

    @Inject
    lateinit var serverConnectionMonitor: ServerConnectionMonitor

    private var shouldOpenTransfers by mutableStateOf(false)

    override fun onStart() {
        super.onStart()
        serverConnectionMonitor.start()
    }

    override fun onStop() {
        serverConnectionMonitor.stop()
        super.onStop()
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        if (TransfersIntents.requestsOpenTransfers(intent)) shouldOpenTransfers = true
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        ensureNotificationPermission()
        if (savedInstanceState == null) shouldOpenTransfers = TransfersIntents.requestsOpenTransfers(intent)
        setContent {
            KuranasTheme {
                KuranasRoot(
                    modifier = Modifier.fillMaxSize(),
                    shouldOpenTransfers = shouldOpenTransfers,
                    onOpenTransfersHandled = { shouldOpenTransfers = false },
                )
            }
        }
    }

    /**
     * minSdk 33 sempre exige POST_NOTIFICATIONS em runtime; sem ela o foreground service de
     * mídia não consegue exibir a notificação com os controles de reprodução.
     */
    private fun ensureNotificationPermission() {
        val granted = ContextCompat.checkSelfPermission(this, Manifest.permission.POST_NOTIFICATIONS) ==
            PackageManager.PERMISSION_GRANTED
        if (!granted) requestNotifications.launch(Manifest.permission.POST_NOTIFICATIONS)
    }
}
