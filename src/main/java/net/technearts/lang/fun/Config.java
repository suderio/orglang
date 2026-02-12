package net.technearts.lang.fun;

import io.smallrye.config.ConfigMapping;
import io.smallrye.config.WithName;

@ConfigMapping(prefix = "fun")
public interface Config {
    @WithName("debug")
    Boolean debug();

    @WithName("math.precision")
    Integer precision();

}