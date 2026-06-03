package com.kuranas.mobile.app;

import org.junit.Test;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;

public class MainNavigationCoordinatorTest {

    private final MainNavigationCoordinator coordinator = new MainNavigationCoordinator();

    @Test
    public void resolveNavigation_homeDoesNotAddBackStack() {
        MainNavigationCoordinator.NavigationInstruction instruction =
                coordinator.resolveNavigation(MainNavigationCoordinator.NAV_HOME);

        assertEquals(MainNavigationCoordinator.NAV_HOME, instruction.getSelectedPosition());
        assertEquals(MainNavigationCoordinator.Destination.HOME, instruction.getDestination());
        assertFalse(instruction.shouldAddToBackStack());
    }

    @Test
    public void resolveNavigation_filesAddsBackStack() {
        MainNavigationCoordinator.NavigationInstruction instruction =
                coordinator.resolveNavigation(MainNavigationCoordinator.NAV_FILES);

        assertEquals(MainNavigationCoordinator.NAV_FILES, instruction.getSelectedPosition());
        assertEquals(MainNavigationCoordinator.Destination.FILES, instruction.getDestination());
        assertTrue(instruction.shouldAddToBackStack());
    }

    @Test
    public void resolveNavigation_invalidPositionFallsBackToHome() {
        MainNavigationCoordinator.NavigationInstruction instruction =
                coordinator.resolveNavigation(999);

        assertEquals(MainNavigationCoordinator.NAV_HOME, instruction.getSelectedPosition());
        assertEquals(MainNavigationCoordinator.Destination.HOME, instruction.getDestination());
        assertFalse(instruction.shouldAddToBackStack());
    }

    @Test
    public void resolveBackPress_whenDrawerOpen_closesDrawer() {
        MainNavigationCoordinator.BackPressDecision decision =
                coordinator.resolveBackPress(true, false, 0);

        assertEquals(MainNavigationCoordinator.BackAction.CLOSE_DRAWER, decision.getAction());
        assertFalse(decision.shouldUpdateSelection());
    }

    @Test
    public void resolveBackPress_whenFilesFragmentConsumesBack_isConsumed() {
        MainNavigationCoordinator.BackPressDecision decision =
                coordinator.resolveBackPress(false, true, 2);

        assertEquals(MainNavigationCoordinator.BackAction.CONSUMED, decision.getAction());
        assertFalse(decision.shouldUpdateSelection());
    }

    @Test
    public void resolveBackPress_withBackStack_popsAndSelectsHome() {
        MainNavigationCoordinator.BackPressDecision decision =
                coordinator.resolveBackPress(false, false, 1);

        assertEquals(MainNavigationCoordinator.BackAction.POP_BACK_STACK, decision.getAction());
        assertTrue(decision.shouldUpdateSelection());
        assertEquals(MainNavigationCoordinator.NAV_HOME, decision.getSelectedPosition());
    }

    @Test
    public void resolveBackPress_withoutBackStack_returnsNoOp() {
        MainNavigationCoordinator.BackPressDecision decision =
                coordinator.resolveBackPress(false, false, 0);

        assertEquals(MainNavigationCoordinator.BackAction.NO_OP, decision.getAction());
        assertFalse(decision.shouldUpdateSelection());
    }
}
