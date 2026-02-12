
# build and copy to ~/bin
native:
  mvn package -Pnative
  cp ./target/fun-runner ~/.local/bin/fun
  chmod ug+x ~/.local/bin/fun

# just build
build:
  mvn package
