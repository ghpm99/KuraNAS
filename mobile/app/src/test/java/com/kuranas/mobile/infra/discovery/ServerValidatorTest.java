package com.kuranas.mobile.infra.discovery;

import org.junit.Test;

import static org.junit.Assert.assertFalse;

public class ServerValidatorTest {

    @Test
    public void validateReturnsFalseForUnreachableHost() {
        ServerValidator validator = new ServerValidator();
        boolean result = validator.validate("http://192.0.2.1:8000");
        assertFalse(result);
    }

    @Test
    public void validateReturnsFalseForInvalidUrl() {
        ServerValidator validator = new ServerValidator();
        boolean result = validator.validate("not-a-valid-url");
        assertFalse(result);
    }

    @Test
    public void validateReturnsFalseForEmptyUrl() {
        ServerValidator validator = new ServerValidator();
        boolean result = validator.validate("");
        assertFalse(result);
    }
}
