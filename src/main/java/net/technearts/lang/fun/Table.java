package net.technearts.lang.fun;

import java.math.BigInteger;
import java.util.LinkedHashMap;

import static java.math.BigInteger.valueOf;
import static java.util.Objects.requireNonNull;

public class Table extends LinkedHashMap<Object, Object> {
    private int lastIndex;
    private boolean stackOn = true;
    private String name;

    @Override
    public Object put(Object key, Object value) {
        if (stackOn) {
            if (requireNonNull(key) instanceof BigInteger) {
                if (((BigInteger) key).compareTo(valueOf(lastIndex + 1)) == 0) {
                    lastIndex++;
                }
            }
            return super.put(key, value);
        } else return null;
    }

    @Override
    public boolean remove(Object key, Object value) {
        if (stackOn) {
            if (requireNonNull(key) instanceof BigInteger) {
                if (((BigInteger) key).compareTo(valueOf(lastIndex)) == 0) {
                    lastIndex--;
                }
            }
            return super.remove(key, value);
        } else return false;
    }

    public void put(Object value) {
        if (stackOn) super.put(valueOf(lastIndex++), value);
    }

    public Object get(int i) {
        return super.get(valueOf(i));
    }

//    public Object push(Object value) {
//        if (stackOn) if (lastIndex > 0)
//            for (int i = lastIndex; i >= 0; i--) {
//                put(valueOf(i + 1), super.remove(valueOf(i)));
//            }
//        return put(ZERO, value);
//    }

    public void turnOn() {
        stackOn = true;
    }

    public void turnOff() {
        stackOn = false;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getName() {
        return name;
    }
}
