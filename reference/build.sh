#!/bin/sh
# Builds Flying Saucer (core + pdf) from a copy of its source tree and compiles
# RefRender against it. Needs a JDK 21 or later and Maven. FLYINGSAUCER names a
# checkout of https://github.com/flyingsaucerproject/flyingsaucer (at the
# commit PORTING.md names); without it the repository is cloned.
#
#   FLYINGSAUCER=path/to/flyingsaucer JAVA_HOME=/usr/lib/jvm/java-25-openjdk-amd64 ./build.sh
#
# Everything is written under ./build, which is not committed.
set -e
here=$(cd "$(dirname "$0")" && pwd)
build="$here/build"
PATH="$JAVA_HOME/bin:$PATH"
rm -rf "$build"
mkdir -p "$build"
if [ -n "$FLYINGSAUCER" ]; then
	cp -r "$FLYINGSAUCER" "$build/flyingsaucer"
else
	git clone -q https://github.com/flyingsaucerproject/flyingsaucer.git "$build/flyingsaucer"
	git -C "$build/flyingsaucer" checkout -q 14f4747b65ca4abba0bd75edb06617af7da46036
fi
cd "$build/flyingsaucer"
mvn -q -pl flying-saucer-pdf -am -DskipTests -Dmaven.javadoc.skip=true -Denforcer.skip=true install
mvn -q -pl flying-saucer-pdf dependency:build-classpath -Dmdep.outputFile="$build/deps.txt"
core=$(ls flying-saucer-core/target/flying-saucer-core-*[!s].jar | grep -v sources | head -1)
pdf=$(ls flying-saucer-pdf/target/flying-saucer-pdf-*.jar | grep -v sources | head -1)
echo "$build/flyingsaucer/$pdf:$build/flyingsaucer/$core:$(cat "$build/deps.txt")" > "$build/classpath.txt"
javac -cp "$(cat "$build/classpath.txt")" -d "$build/classes" "$here/RefRender.java"
echo "built; run: $here/render.sh in.xhtml out.pdf out.boxes"
