package com.kuranas.android.feature.files.ui

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.viewModels
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.ui.Modifier
import com.kuranas.android.core.ui.components.KNFrame
import com.kuranas.android.feature.files.data.sharedUris
import com.kuranas.android.feature.transfers.ui.TransfersIntents
import com.kuranas.android.ui.theme.KuranasTheme
import dagger.hilt.android.AndroidEntryPoint

/** Alvo de "Compartilhar" do sistema (ACTION_SEND e ACTION_SEND_MULTIPLE): escolhe o destino e enfileira o envio. */
@AndroidEntryPoint
class ShareTargetActivity : ComponentActivity() {

    private val uploadViewModel: UploadViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        val sharedUris = intent.sharedUris()
        if (sharedUris.isEmpty()) {
            finish()
            return
        }
        if (savedInstanceState == null) uploadViewModel.receiveShared(sharedUris)
        setContent {
            KuranasTheme {
                KNFrame(modifier = Modifier.fillMaxSize()) {
                    ShareTargetScreen(
                        selectedFileCount = sharedUris.size,
                        onClose = ::finish,
                        onViewTransfers = ::openTransfersAndFinish,
                        uploadViewModel = uploadViewModel,
                    )
                }
            }
        }
    }

    private fun openTransfersAndFinish() {
        startActivity(TransfersIntents.openTransfers(this))
        finish()
    }
}
