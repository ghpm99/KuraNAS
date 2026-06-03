package com.kuranas.mobile.app;

public final class MainNavigationCoordinator {

    public static final int NAV_HOME = 0;
    public static final int NAV_FILES = 1;
    public static final int NAV_IMAGES = 2;
    public static final int NAV_MUSIC = 3;
    public static final int NAV_VIDEOS = 4;
    public static final int NAV_SEARCH = 5;
    public static final int NAV_SETTINGS = 6;

    public NavigationInstruction resolveNavigation(int requestedPosition) {
        int selectedPosition = normalizePosition(requestedPosition);
        return new NavigationInstruction(
                selectedPosition,
                destinationFor(selectedPosition),
                selectedPosition != NAV_HOME
        );
    }

    public BackPressDecision resolveBackPress(
            boolean isDrawerOpen,
            boolean filesHandledBackNavigation,
            int backStackEntryCount
    ) {
        if (isDrawerOpen) {
            return new BackPressDecision(BackAction.CLOSE_DRAWER, -1);
        }
        if (filesHandledBackNavigation) {
            return new BackPressDecision(BackAction.CONSUMED, -1);
        }
        if (backStackEntryCount > 0) {
            return new BackPressDecision(BackAction.POP_BACK_STACK, NAV_HOME);
        }
        return new BackPressDecision(BackAction.NO_OP, -1);
    }

    private int normalizePosition(int requestedPosition) {
        if (requestedPosition < NAV_HOME || requestedPosition > NAV_SETTINGS) {
            return NAV_HOME;
        }
        return requestedPosition;
    }

    private Destination destinationFor(int selectedPosition) {
        switch (selectedPosition) {
            case NAV_FILES:
                return Destination.FILES;
            case NAV_IMAGES:
                return Destination.IMAGES;
            case NAV_MUSIC:
                return Destination.MUSIC;
            case NAV_VIDEOS:
                return Destination.VIDEOS;
            case NAV_SEARCH:
                return Destination.SEARCH;
            case NAV_SETTINGS:
                return Destination.SETTINGS;
            case NAV_HOME:
            default:
                return Destination.HOME;
        }
    }

    public enum Destination {
        HOME,
        FILES,
        IMAGES,
        MUSIC,
        VIDEOS,
        SEARCH,
        SETTINGS
    }

    public enum BackAction {
        CLOSE_DRAWER,
        CONSUMED,
        POP_BACK_STACK,
        NO_OP
    }

    public static final class NavigationInstruction {
        private final int selectedPosition;
        private final Destination destination;
        private final boolean addToBackStack;

        NavigationInstruction(int selectedPosition, Destination destination, boolean addToBackStack) {
            this.selectedPosition = selectedPosition;
            this.destination = destination;
            this.addToBackStack = addToBackStack;
        }

        public int getSelectedPosition() {
            return selectedPosition;
        }

        public Destination getDestination() {
            return destination;
        }

        public boolean shouldAddToBackStack() {
            return addToBackStack;
        }
    }

    public static final class BackPressDecision {
        private final BackAction action;
        private final int selectedPosition;

        BackPressDecision(BackAction action, int selectedPosition) {
            this.action = action;
            this.selectedPosition = selectedPosition;
        }

        public BackAction getAction() {
            return action;
        }

        public int getSelectedPosition() {
            return selectedPosition;
        }

        public boolean shouldUpdateSelection() {
            return selectedPosition >= 0;
        }
    }
}
