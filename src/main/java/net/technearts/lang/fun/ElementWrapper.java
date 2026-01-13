package net.technearts.lang.fun;

import jakarta.annotation.Nonnull;

import java.math.BigDecimal;
import java.math.BigInteger;
import java.math.MathContext;

import static java.lang.String.valueOf;
import static java.math.BigDecimal.ONE;
import static java.math.BigDecimal.ZERO;
import static java.math.RoundingMode.HALF_UP;
import static net.technearts.lang.fun.ElementWrapper.Nil.NULL;

public class ElementWrapper<T> implements Comparable<T> {
  private final T value;
  private final MathContext mathContext = new MathContext(50, HALF_UP);

  public ElementWrapper(T value) {
    this.value = value;
  }

  public static <V> ElementWrapper<V> wrap(V value) {
    return new ElementWrapper<>(value);
  }

  public BigDecimal getDecimal() {
    return switch (value) {
      case BigDecimal v -> v;
      case BigInteger v -> new BigDecimal(v);
      case Table v -> new BigDecimal(v.size());
      case String v -> new BigDecimal(v.length());
      case Boolean v -> v ? ONE : ZERO;
      case null, default -> null;
    };
  }

  public BigInteger getInteger() {
    return switch (value) {
      case BigDecimal v -> v.round(mathContext).toBigInteger();
      case BigInteger v -> v;
      case Table v -> new BigInteger(valueOf(v.size()));
      case String v -> new BigInteger(valueOf(v.length()));
      case Boolean v -> v ? BigInteger.ONE : BigInteger.ZERO;
      case null, default -> null;
    };
  }

  public Boolean getBoolean() {
    return switch (value) {
      case BigDecimal v -> v.compareTo(ZERO) != 0;
      case BigInteger v -> v.compareTo(BigInteger.ZERO) != 0;
      case Table v -> !v.isEmpty();
      case String v -> !v.isEmpty();
      case Boolean v -> v;
      case null, default -> null;
    };
  }

  public Object get() {
    return switch (value) {
      case BigDecimal v -> v;
      case null -> null;
      default -> this.getInteger();
    };
  }

  public Boolean isNull() {
    return this.value == NULL;
  }

  public Object negate() {
    return switch (value) {
      case BigDecimal v -> v.negate();
      case null -> null;
      default -> this.getInteger().negate();
    };
  }

  public Object add(ElementWrapper<?> ne) {
    if (this.getDecimal() == null || ne.getDecimal() == null) {
      return NULL;
    } else if (this.value instanceof BigDecimal || ne.value instanceof BigDecimal) {
      return this.getDecimal().add(ne.getDecimal(), mathContext);
    }
    return this.getInteger().add(ne.getInteger());
  }

  public Object subtract(ElementWrapper<?> ne) {
    if (this.getDecimal() == null || ne.getDecimal() == null) {
      return NULL;
    } else if (this.value instanceof BigDecimal || ne.value instanceof BigDecimal) {
      return this.getDecimal().subtract(ne.getDecimal(), mathContext);
    }
    return this.getInteger().subtract(ne.getInteger());
  }

  public Object multiply(ElementWrapper<?> ne) {
    if (this.getDecimal() == null || ne.getDecimal() == null) {
      return NULL;
    } else if (this.value instanceof BigDecimal || ne.value instanceof BigDecimal) {
      return this.getDecimal().multiply(ne.getDecimal(), mathContext);
    }
    return this.getInteger().multiply(ne.getInteger());
  }

  public Object divide(ElementWrapper<?> ne) {
    if (this.getDecimal() == null || ne.getDecimal() == null) {
      return NULL;
    } else if (this.value instanceof BigDecimal || ne.value instanceof BigDecimal) {
      return this.getDecimal().divide(ne.getDecimal(), mathContext);
    }
    return this.getInteger().divide(ne.getInteger());
  }

  public Object remainder(ElementWrapper<?> ne) {
    if (this.getDecimal() == null || ne.getDecimal() == null) {
      return NULL;
    } else if (this.value instanceof BigDecimal || ne.value instanceof BigDecimal) {
      return this.getDecimal().remainder(ne.getDecimal(), mathContext);
    }
    return this.getInteger().remainder(ne.getInteger());
  }

  public Object pow(ElementWrapper<?> ne) {
    if (this.getDecimal() == null || ne.getDecimal() == null) {
      return NULL;
    } else if (this.value instanceof BigDecimal || ne.value instanceof BigDecimal) {
      return Utils.pow(this.getDecimal(), ne.getDecimal());
    }
    // TODO
    return Utils.pow(this.getDecimal(), ne.getDecimal());
  }

  public Object shiftRight(ElementWrapper<?> ne) {
    return this.getInteger().shiftRight(ne.getInteger().intValue());
  }

  public Object shiftLeft(ElementWrapper<?> ne) {
    return this.getInteger().shiftLeft(ne.getInteger().intValue());
  }

  @Override
  public int compareTo(@Nonnull T ne) {
    if (this.getDecimal() == null || ((ElementWrapper<?>) ne).getDecimal() == null) {
      return -1;
    }
    return this.getDecimal().compareTo(((ElementWrapper<?>) ne).getDecimal());
  }

  public enum Nil {
    NULL
  }
}
