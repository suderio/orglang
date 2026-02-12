package net.technearts.lang.fun;

import org.junit.jupiter.api.AssertionFailureBuilder;

import java.math.BigDecimal;
import java.math.BigInteger;

import static java.math.BigDecimal.valueOf;
import static org.junit.jupiter.api.Assertions.assertEquals;

public class TestUtils {
    public static void assertNumbersEqual(BigDecimal expected, BigDecimal actual) {
        if (expected.compareTo(actual) != 0) {
            AssertionFailureBuilder.assertionFailure().message(null).expected(expected).actual(actual).buildAndThrow();
        }
    }

    public static void assertNumbersEqual(BigDecimal expected, Object actual) {
        if (actual instanceof BigDecimal) {
            assertNumbersEqual(expected, (BigDecimal) actual);
        } else if (actual instanceof BigInteger) {
            assertNumbersEqual(expected, new BigDecimal((BigInteger) actual));
        } else {
            AssertionFailureBuilder.assertionFailure().message(null).expected(expected).actual(actual).buildAndThrow();
        }
    }

    public static void assertNumbersEqual(Integer expected, Object actual) {
        if (actual instanceof BigDecimal) {
            assertNumbersEqual(valueOf(expected), (BigDecimal) actual);
        } else if (actual instanceof BigInteger) {
            assertNumbersEqual(valueOf(expected), new BigDecimal((BigInteger) actual));
        } else {
            AssertionFailureBuilder.assertionFailure().message(null).expected(expected).actual(actual).buildAndThrow();
        }
    }

    public static void assertSizeEquals(Integer expected, Object actual) {
        if (actual instanceof Table) {
            assertEquals(expected, ((Table) actual).size());
        } else {
            AssertionFailureBuilder.assertionFailure().message(null).expected(expected).actual(actual).buildAndThrow();
        }
    }

    public static void assertNumbersEqual(Double expected, Object actual) {
        if (actual instanceof BigDecimal) {
            assertNumbersEqual(valueOf(expected), (BigDecimal) actual);
        } else if (actual instanceof BigInteger) {
            assertNumbersEqual(valueOf(expected), new BigDecimal((BigInteger) actual));
        } else {
            AssertionFailureBuilder.assertionFailure().message(null).expected(expected).actual(actual).buildAndThrow();
        }
    }

    public static void assertAll(Table table) {
        table.values().stream().filter(o -> o instanceof Boolean b && !b).forEach(o -> AssertionFailureBuilder.assertionFailure().message(null).expected(true).actual(false).buildAndThrow());
        if (table.values().stream().noneMatch(o -> o instanceof Boolean)) {
            AssertionFailureBuilder.assertionFailure().message("At least one test expected.").expected("#tests > 0").actual("#tests = 0").buildAndThrow();
        }
    }
}
