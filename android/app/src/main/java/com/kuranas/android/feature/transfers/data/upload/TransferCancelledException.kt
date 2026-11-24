package com.kuranas.android.feature.transfers.data.upload

import kotlin.coroutines.cancellation.CancellationException

class TransferCancelledException : CancellationException("Transfer cancelled by the user")
