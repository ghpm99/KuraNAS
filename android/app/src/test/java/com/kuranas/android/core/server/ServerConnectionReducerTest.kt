package com.kuranas.android.core.server

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class ServerConnectionReducerTest {

    private val offlineSnapshot = ConnectionSnapshot(ServerConnectionState.Offline, consecutiveFailures = 1)

    @Test
    fun `falha na verificacao leva de online para offline`() {
        val transition = reduceConnection(ConnectionSnapshot(), ConnectionObservation.Unreachable)

        assertEquals(ServerConnectionState.Offline, transition.snapshot.state)
        assertFalse(transition.hasReconnected)
    }

    @Test
    fun `servidor respondendo de novo leva de offline para online e sinaliza reconexao`() {
        val transition = reduceConnection(offlineSnapshot, ConnectionObservation.Reachable)

        assertEquals(ServerConnectionState.Online, transition.snapshot.state)
        assertEquals(0, transition.snapshot.consecutiveFailures)
        assertTrue(transition.hasReconnected)
    }

    @Test
    fun `continuar online nao sinaliza reconexao`() {
        val transition = reduceConnection(ConnectionSnapshot(), ConnectionObservation.Reachable)

        assertFalse(transition.hasReconnected)
    }

    @Test
    fun `sem servidor lembrado fica online sem sinalizar reconexao`() {
        val transition = reduceConnection(offlineSnapshot, ConnectionObservation.NoServer)

        assertEquals(ServerConnectionState.Online, transition.snapshot.state)
        assertFalse(transition.hasReconnected)
    }

    @Test
    fun `acesso negado leva ao estado proprio sem marcar offline`() {
        val transition = reduceConnection(ConnectionSnapshot(), ConnectionObservation.Denied(AccessDenial("10.0.0.9", null)))

        assertEquals(ServerConnectionState.AccessDenied, transition.snapshot.state)
        assertFalse(transition.hasReconnected)
    }

    @Test
    fun `voltar a responder depois de acesso negado conta como reconexao`() {
        val denied = ConnectionSnapshot(ServerConnectionState.AccessDenied, 1)

        val transition = reduceConnection(denied, ConnectionObservation.Reachable)

        assertEquals(ServerConnectionState.Online, transition.snapshot.state)
        assertTrue(transition.hasReconnected)
    }

    @Test
    fun `falhas consecutivas acumulam`() {
        val transition = reduceConnection(offlineSnapshot, ConnectionObservation.Unreachable)

        assertEquals(2, transition.snapshot.consecutiveFailures)
    }

    @Test
    fun `acesso negado vira observacao propria`() {
        val denial = AccessDenial("10.0.0.9", "blocked")
        assertEquals(ConnectionObservation.Denied(denial), ServerCheckResult.AccessDenied(denial).toObservation())
        assertEquals(ConnectionObservation.Reachable, ServerCheckResult.Confirmed.toObservation())
    }

    @Test
    fun `intervalo offline cresce ate o teto e online usa o intervalo longo`() {
        val policy = ConnectionCheckPolicy()

        assertEquals(60_000L, policy.nextCheckDelayMillis(ConnectionSnapshot()))
        assertEquals(10_000L, policy.nextCheckDelayMillis(ConnectionSnapshot(ServerConnectionState.Offline, 1)))
        assertEquals(20_000L, policy.nextCheckDelayMillis(ConnectionSnapshot(ServerConnectionState.Offline, 2)))
        assertEquals(30_000L, policy.nextCheckDelayMillis(ConnectionSnapshot(ServerConnectionState.Offline, 9)))
    }
}
