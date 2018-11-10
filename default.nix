{ stdenv, pkgs, ... }:

stdenv.mkDerivation {
  name = "kubernixos";

  src = with pkgs; writeText "kubernixos" ''
    #!${stdenv.shell}
    set -euo pipefail
    PACKAGES="''${PACKAGES:-${path}}"

    JSON=$(${nix}/bin/nix eval --arg packages "$PACKAGES" --arg modules "$MODULES" -f @out@/lib/kubernixos.nix manifests --json)
    echo $JSON | \
      ${kubectl}/bin/kubectl $@ -l reconciler=kubernixos --prune -f -
  '';

  phases = [ "installPhase" ];

  installPhase = ''
    mkdir -p $out/bin $out/lib
    cp -r ${./lib}/* $out/lib
    substitute $src $out/bin/kubernixos --subst-var out
    chmod +x $out/bin/kubernixos
  '';
}
