{ stdenv, pkgs, ... }:

stdenv.mkDerivation {
  name = "kubernixos";

  # Rewrite this is a real programming language at some point
  # Bash is.. Annoying
  src = with pkgs;
  let
    kubectl = pkgs.kubectl.overrideAttrs (oldAttrs: rec {
      patches = [
       (fetchpatch {
         url    = "https://github.com/kubernetes/kubernetes/pull/69344.patch";
         sha256 = "0sz4766n5300l9wxk0yj4q7w5ybfkx9wfm5ljfrx5s5vhm1zj0bd";
       })
      ];
    });
  in
  writeText "kubernixos" ''
    #!${stdenv.shell}
    set -euo pipefail
    PACKAGES="''${PACKAGES:-${path}}"

    TOP=$(${nix}/bin/nix eval --arg packages "$PACKAGES" --arg modules "$MODULES" -f @out@/lib/kubernixos.nix kubernixos --json)
    SERVER=$(echo $TOP | ${jq}/bin/jq -rMc '.config.server')
    MANIFESTS=$(echo $TOP | ${jq}/bin/jq -Mc '.manifests')
    echo $MANIFESTS | \
      ${kubectl}/bin/kubectl -s $SERVER $@ -l reconciler=kubernixos --prune -f -
  '';

  phases = [ "installPhase" ];

  installPhase = ''
    mkdir -p $out/bin $out/lib
    cp -r ${./lib}/* $out/lib
    substitute $src $out/bin/kubernixos --subst-var out
    chmod +x $out/bin/kubernixos
  '';
}
