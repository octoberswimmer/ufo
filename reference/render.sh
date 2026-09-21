#!/bin/sh
# Renders a document with the Java Flying Saucer build made by build.sh:
#   render.sh in.xhtml out.pdf [out.boxes]
# out.boxes receives one line per box of the laid-out tree with its absolute
# position and size in dots, which is what the Go port's layout is compared
# with.
set -e
here=$(cd "$(dirname "$0")" && pwd)
PATH="$JAVA_HOME/bin:$PATH"
exec java -cp "$(cat "$here/build/classpath.txt"):$here/build/classes" RefRender "$@" 2>&1
