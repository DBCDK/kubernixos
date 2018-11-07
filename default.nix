{ stdenv, pkgs, ... }:

stdenv.mkDerivation {
  name = "kubernixos";

  src = with pkgs; writeText "kubernixos" ''
    #!${stdenv.shell}
    set -euo pipefail
    PACKAGES="''${PACKAGES:-${path}}"

    ${nix}/bin/nix eval --arg packages "$PACKAGES" --arg modules "$MODULES" -f ${./kubernixos.nix} manifests --json | \
      ${kubectl}/bin/kubectl apply -l reconciler=kubernixos --prune -f - $@
  '';

  phases = [ "installPhase" ];

  installPhase = ''
    mkdir -p $out/bin
    cp $src $out/bin/kubernixos
    chmod +x $out/bin/kubernixos
  '';
}
