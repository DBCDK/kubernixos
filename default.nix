{ stdenv, pkgs, ... }:

stdenv.mkDerivation {
  name = "kubernixos";

  # Rewrite this is a real programming language at some point
  # Bash is.. Annoying
  src = with pkgs; writeText "kubernixos" ''
    #!${stdenv.shell}
    set -euo pipefail
    PACKAGES="''${PACKAGES:-${path}}"

    TOP=$(${nix}/bin/nix eval --arg packages "$PACKAGES" --arg modules "$MODULES" -f @out@/lib/kubernixos.nix kubernixos --json)
    SERVER=$(echo $TOP | ${jq}/bin/jq -rMc '.config.server')

    OUT=$(mktemp)
    echo $TOP | ${jq}/bin/jq -Mc '.manifests' > $OUT

    if [ -n "''${DEBUG:-}" ]; then
      echo "Output written to: $OUT"
    else
      ${kubectl}/bin/kubectl -s $SERVER $@ -l reconciler=kubernixos --prune -f $OUT
      rm $OUT
    fi
  '';

  phases = [ "installPhase" ];

  installPhase = ''
    mkdir -p $out/bin $out/lib
    cp -r ${./lib}/* $out/lib
    substitute $src $out/bin/kubernixos --subst-var out
    chmod +x $out/bin/kubernixos
  '';
}
