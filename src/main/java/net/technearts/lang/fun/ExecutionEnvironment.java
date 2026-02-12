package net.technearts.lang.fun;

import java.math.MathContext;

import static java.math.RoundingMode.HALF_UP;

public class ExecutionEnvironment {
    private final MathContext mathContext;

    private final boolean debug;

    public ExecutionEnvironment() {
        mathContext = new MathContext(50, HALF_UP);
        debug = true;
    }

    public ExecutionEnvironment(Config config) {
        mathContext = new MathContext(config.precision(), HALF_UP);
        debug = config.debug();
    }

    public boolean isDebug() {
        return debug;
    }
}
